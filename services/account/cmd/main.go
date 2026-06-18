package main

import (
	"account/internal/account"
	"account/internal/database"
	"account/logger"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Server struct {
	authService   *account.AuthService
	followService *account.FollowService
	log           *slog.Logger
	mux           *http.ServeMux
}
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

func NewServer(as *account.AuthService, fs *account.FollowService, log *slog.Logger) *Server {
	s := &Server{
		authService:   as,
		followService: fs,
		log:           log,
		mux:           http.NewServeMux(),
	}
	s.mux.HandleFunc("/auth/login", s.handleLogin)
	s.mux.HandleFunc("/auth/register", s.handleRegistration)
	s.mux.HandleFunc("/auth/logout", s.handleLogout)
	s.mux.HandleFunc("/auth/refresh", s.handleRefresh)
	s.mux.HandleFunc("GET /{username}", s.handlePublicProfile)
	// POST /follow
	follow := JWTMiddleware(s.authService.TokenSecret(), s.log, http.HandlerFunc(s.handleFollow))
	s.mux.Handle("POST /follow/{username}", follow)
	// GET /followers
	getFollowers := JWTMiddleware(s.authService.TokenSecret(), s.log, http.HandlerFunc(s.handleGetFollowers))
	s.mux.Handle("GET /followers/{username}", getFollowers)
	// GET /following
	getFollowing := JWTMiddleware(s.authService.TokenSecret(), s.log, http.HandlerFunc(s.handleGetFollowing))
	s.mux.Handle("GET /following/{username}", getFollowing)
	return s
}

//type FollowRequest struct {
//	Username string `json:"username"`
//}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.log.Info("incoming request", "method", r.Method, "path", r.URL.Path)
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleFollow(w http.ResponseWriter, r *http.Request) {
	// из контекста после jwtMiddleWare получить account id
	//username из r.Body
	ulog := s.log.With(
		slog.String("path", r.URL.Path),
		slog.String("method", r.Method),
		slog.String("adress", r.RemoteAddr))

	followerID, ok := r.Context().Value("account_id").(int64)
	if !ok {
		ulog.Error("Unauthorized")
		writeError(w, http.StatusUnauthorized, "Unauthorized")
	}

	username := r.PathValue("username")
	if utf8.RuneCountInString(username) < 3 {
		writeError(w, http.StatusBadRequest, "bad username")
	}

	followingID, err := s.authService.GetAccountIdByUsername(username)
	if err != nil {
		ulog.Error("no found by username", slog.Any("error", err))
		writeError(w, http.StatusNotFound, "no found")
		return
	}
	err = s.followService.Follow(followerID, followingID)
	if err != nil {
		ulog.Error("follow error", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal service error")
		return
	}

	writeJSON(w, http.StatusOK, username)

}

func (s *Server) handleGetFollowers(w http.ResponseWriter, r *http.Request) {
	// из контекста после jwtMiddleWare получить account id
	ulog := s.log.With(
		slog.String("path", r.URL.Path),
		slog.String("method", r.Method),
		slog.String("adress", r.RemoteAddr))

	username := r.PathValue("username")
	if utf8.RuneCountInString(username) < 3 {
		writeError(w, http.StatusBadRequest, "bad username")
	}

	targetID, err := s.authService.GetAccountIdByUsername(username)
	if err != nil {
		ulog.Error("no found by username", slog.Any("error", err))
		writeError(w, http.StatusNotFound, "no found")
		return
	}

	followers, err := s.followService.GetFollowers(targetID)
	if err != nil {
		ulog.Error("error with found followers", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal service error")
		return
	}

	writeJSON(w, http.StatusOK, followers)
}

func (s *Server) handleGetFollowing(w http.ResponseWriter, r *http.Request) {
	// из контекста после jwtMiddleWare получить account id
	ulog := s.log.With(
		slog.String("path", r.URL.Path),
		slog.String("method", r.Method),
		slog.String("adress", r.RemoteAddr))

	username := r.PathValue("username")
	if utf8.RuneCountInString(username) < 3 {
		writeError(w, http.StatusBadRequest, "bad username")
	}

	targetID, err := s.authService.GetAccountIdByUsername(username)
	if err != nil {
		ulog.Error("no found by username", slog.Any("error", err))
		writeError(w, http.StatusNotFound, "no found")
		return
	}

	following, err := s.followService.GetFollowing(targetID)
	if err != nil {
		ulog.Error("error with found followers", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal service error")
		return
	}

	writeJSON(w, http.StatusOK, following)
}
func (s *Server) handlePublicProfile(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if utf8.RuneCountInString(username) < 3 {
		writeError(w, http.StatusBadRequest, "invalid username")
		return
	}
	profile, err := s.authService.GetPublicProfile(username)
	if err != nil {
		s.log.Error("failed to get public profile", "error", err)
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LogoutRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	hash := sha256.Sum256([]byte(req.RefreshToken))
	refreshTokenHash := hex.EncodeToString(hash[:])
	err = s.authService.RevokeSessionByRefreshTokenHash(refreshTokenHash)
	if err != nil {
		s.log.Error(r.URL.Path, "error", err.Error())
		writeError(w, http.StatusInternalServerError, "invalid logout")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// refresh_token → refresh session → new access token
func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RefreshRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	hash := sha256.Sum256([]byte(req.RefreshToken))
	refreshTokenHash := hex.EncodeToString(hash[:])
	session, err := s.authService.FindSessionByRefreshTokenHash(refreshTokenHash)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid refresh")
		return
	}
	if session.RevokedAt != nil && session.ExpiresAt.Before(time.Now()) {
		writeError(w, http.StatusInternalServerError, "invalid refresh")
		return
	}
	newAccessToken, err := s.authService.GenerateAccessToken(session.AccountID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid refresh")
		return
	}

	writeJSON(w, http.StatusOK, RefreshResponse{
		AccessToken: newAccessToken})
}

func (s *Server) handleRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req account.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Error("failed to decode login request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	ans, err := s.authService.Register(req)
	if err != nil {
		switch {
		case errors.Is(err, account.ErrEmailAlreadyExists):
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		case errors.Is(err, account.ErrUsernameAlreadyExists):
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		default:
			s.log.Error("registration failed", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return

		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ans); err != nil {
		s.log.Error("failed to encode ans response", "error", err)
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req account.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Error("failed to decode login request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tokens, err := s.authService.Login(req)
	if err != nil {
		s.log.Error("login failed", "email", req.Email, "error", err)
		if errors.Is(err, account.ErrAccountByEmailNotFound) {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		if errors.Is(err, account.ErrInvalidPassword) {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tokens); err != nil {
		s.log.Error("failed to encode tokens response", "error", err)
	}
}

func JWTMiddleware(jwtSecret string, log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}
		uctx := context.WithValue(r.Context(), "X-Request-ID", reqID)

		ulog := log.With(slog.String("request_id", reqID),
			slog.String("path", r.URL.Path),
			slog.String("method", r.Method))

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "missing token")
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ulog.ErrorContext(uctx, "invalid token format")
			writeError(w, http.StatusUnauthorized, "invalid token format")
			return
		}
		tokenStr := parts[1]
		// parse jwt
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			// context for trace?
			ulog.ErrorContext(uctx, "bad parse token", slog.Any("error", err))
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		// Go parse json in  float64, fan fact
		sub, ok := claims["sub"].(float64)
		if !ok {
			ulog.ErrorContext(uctx, "invalid sub", slog.Any("error", err))
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		accountID := int64(sub)

		ctx := context.WithValue(uctx, "account_id", accountID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func main() {
	log := logger.New()
	log.Info("The server is running, but it still attracts nerds.")

	err := godotenv.Load(".env", "./.env")
	if err != nil {
		log.Info(".env not found, using environment variables")
	}

	db, err := database.NewPostgress()
	if err != nil {
		log.Error(err.Error())
		panic(err)
	}
	defer db.Close()

	// Make sure it works.
	err = db.Ping()
	if err != nil {
		log.Error(err.Error())
		panic(err)
	}

	accountRepo := account.NewRepository(db)

	tm := account.NewTokenManager(os.Getenv("JWT_SECRET"))
	authService := account.NewAuthService(accountRepo, tm)
	followService := account.NewFollowService(accountRepo)
	server := NewServer(authService, followService, log)

	log.Info("server is running on :8080")
	if err := http.ListenAndServe("0.0.0.0:8080", server); err != nil {
		log.Error("server stopped with error", "error", err)
	}

}
