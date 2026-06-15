package main

import (
	"account/internal/account"
	"account/internal/database"
	"account/logger"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Server struct {
	authService *account.AuthService
	log         *slog.Logger
	mux         *http.ServeMux
}

func NewServer(as *account.AuthService, log *slog.Logger) *Server {
	s := &Server{
		authService: as,
		log:         log,
		mux:         http.NewServeMux(),
	}
	s.mux.HandleFunc("/auth/login", s.handleLogin)
	s.mux.HandleFunc("/auth/register", s.handleRegistration)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.log.Info("incoming request", "method", r.Method, "path", r.URL.Path)
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req account.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Warn("failed to decode login request", "error", err)
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
		s.log.Warn("failed to decode login request", "error", err)
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

//func writeJSON(w http.ResponseWriter, status int, v any) {
//	w.Header().Set("Content-Type", "application/json")
//	w.WriteHeader(status)
//	_ = json.NewEncoder(w).Encode(v)
//}
//
//func writeError(w http.ResponseWriter, status int, msg string) {
//	writeJSON(w, status, map[string]string{"error": msg})
//}

func main() {
	log := logger.New()
	log.Info("The server is running, but it still attracts nerds.")

	err := godotenv.Load(".env", "./.env")
	if err != nil {
		log.Info("No .env found")
	}

	db, err := database.NewPostgress()
	if err != nil {
		log.Error(err.Error())
	}
	defer db.Close()

	// Make sure it works.
	err = db.Ping()
	if err != nil {
		log.Error(err.Error())
	}

	accountRepo := account.NewRepository(db)
	// db init
	err = accountRepo.Init()
	if err != nil {
		log.Error(err.Error())
		panic(err)
	}

	tm := account.NewTokenManager(os.Getenv("JWT_SECRET"))
	authService := account.NewAuthService(accountRepo, tm)

	server := NewServer(authService, log)

	log.Info("server is running on :8080")
	if err := http.ListenAndServe("localhost:8080", server); err != nil {
		log.Error("server stopped with error", "error", err)
	}

}
