package account

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	r  *Repository
	tm *TokenManager
}

type RegisterRequest struct {
	Email       string
	Password    string
	Username    string
	DisplayName string
}

type RegisterResponse struct {
	Profile Profile
	Token   Token
}

type LoginRequest struct {
	Email     string
	Password  string
	UserAgent string
}

type TokenManager struct {
	jwtSecret string
}

func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{jwtSecret: secret}
}

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

var usernameRegex = regexp.MustCompile(`^[a-z0-9_]{3,20}$`)

var (
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")

	ErrInvalidPassword = errors.New("Invalid password")

	ErrAccountByEmailNotFound = errors.New("account by email not found")

	ErrInvalidEmail       = errors.New("invalid email")
	ErrWeakPassword       = errors.New("weak password")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidDisplayName = errors.New("invalid display name")
)

func NewAuthService(r *Repository, tm *TokenManager) *AuthService {
	return &AuthService{r: r, tm: tm}
}

func (s *AuthService) Register(
	req RegisterRequest,
) (RegisterResponse, error) {
	// check validation
	err := validateRegisterRequest(req)
	if err != nil {
		return RegisterResponse{}, err
	}

	tx, err := s.r.db.Begin()
	if err != nil {
		return RegisterResponse{}, err
	}
	defer tx.Rollback()

	// https://pkg.go.dev/golang.org/x/crypto/bcrypt
	bytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterResponse{}, err
	}
	passwordHash := string(bytes)

	// create account
	accountID, err := s.r.createAccount(tx, req.Email, passwordHash)
	if err != nil {
		return RegisterResponse{}, err
	}
	//create profile
	err = s.r.createProfile(tx, accountID, req.Username, req.DisplayName)

	if err != nil {
		return RegisterResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return RegisterResponse{}, err
	}

	token, err := s.Login(LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return RegisterResponse{}, err
	}
	ans := RegisterResponse{
		Token: token,
		Profile: Profile{
			AccountID:   accountID,
			Username:    req.Username,
			DisplayName: req.DisplayName,
		},
	}
	return ans, nil
}
func (s *AuthService) RevokeSessionByRefreshTokenHash(refreshTokenHash string) error {
	err := s.r.RevokeSessionByRefreshTokenHash(refreshTokenHash)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) FindSessionByRefreshTokenHash(refreshTokenHash string) (Session, error) {
	session, err := s.r.FindSessionByRefreshTokenHash(refreshTokenHash)
	if err != nil {
		return session, err
	}
	return session, nil
}

func (s *AuthService) GenerateAccessToken(accountID int64) (string, error) {
	token, err := s.tm.generateAccessToken(accountID)
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func (s *AuthService) GetPublicProfile(username string) (PublicProfile, error) {
	profile, err := s.r.GetProfileByUsername(username)
	if err != nil {
		return PublicProfile{}, err
	}
	return *profile, nil
}
func (s *AuthService) Login(req LoginRequest) (Token, error) {
	account, err := s.r.getAccountByEmail(req.Email)
	if err != nil {
		return Token{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.Password))
	if err != nil {
		return Token{}, ErrInvalidPassword
	}

	token, err := s.tm.generateAccessToken(account.ID)
	if err != nil {
		return Token{}, err
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return Token{}, err
	}
	refreshString := hex.EncodeToString(b)
	token.RefreshToken = refreshString
	hash := sha256.Sum256([]byte(refreshString))
	refreshTokenHash := hex.EncodeToString(hash[:])

	session := Session{
		AccountID:        account.ID,
		RefreshTokenHash: refreshTokenHash,
		UserAgent:        req.UserAgent,
		ExpiresAt:        time.Now().Add(30 * 24 * time.Hour).UTC(),
	}
	err = s.r.CreateSession(session)
	if err != nil {
		return Token{}, err
	}

	return token, nil
}

func (t *TokenManager) generateAccessToken(accountID int64) (Token, error) {
	accessClaims := jwt.MapClaims{
		"sub": accountID,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)

	accessString, err := accessToken.SignedString([]byte(t.jwtSecret))
	if err != nil {
		return Token{}, err
	}

	return Token{
		AccessToken: accessString,
	}, nil
}

func validateRegisterRequest(req RegisterRequest) error {

	if err := validateEmail(req.Email); err != nil {
		return err
	}

	if err := validatePassword(req.Password); err != nil {
		return err
	}

	if err := validateUsername(req.Username); err != nil {
		return err
	}

	if err := validateDisplayName(req.DisplayName); err != nil {
		return err
	}

	return nil
}

func validateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}
	return nil
}

func validatePassword(password string) error {
	if utf8.RuneCountInString(password) < 8 || utf8.RuneCountInString(password) > 15 {
		return ErrWeakPassword
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range password {
		if hasUpper && hasLower && hasDigit && hasSpecial {
			break
		}
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return ErrWeakPassword
	}
	return nil
}

func validateUsername(username string) error {
	if !usernameRegex.MatchString(username) {
		return ErrInvalidUsername
	}

	return nil
}

func validateDisplayName(displayName string) error {
	displayName = strings.TrimSpace(displayName)

	if utf8.RuneCountInString(displayName) == 0 || utf8.RuneCountInString(displayName) > 50 {
		return ErrInvalidDisplayName
	}

	return nil
}
