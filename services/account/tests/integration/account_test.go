package integration

import (
	"account/internal/account"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// потестить repo создание аккаунта\профиля и т.п.
func setupDB(t *testing.T) *sql.DB {
	db, err := sql.Open(
		"postgres",
		"postgres://test:test@localhost:5433/testdb?sslmode=disable",
	)

	require.NoError(t, err)
	require.NoError(t, db.Ping())

	return db
}
func cleanDB(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM follows")
	require.NoError(t, err)

	_, err = db.Exec("DELETE FROM accounts")
	require.NoError(t, err)
}

func TestAuthFlow(t *testing.T) {
	testDB := setupDB(t)
	t.Cleanup(
		func() {
			cleanDB(t, testDB)
		})

	repo := account.NewRepository(testDB)
	authService := account.NewAuthService(repo, &account.TokenManager{})

	signUpReq := account.RegisterRequest{
		Email:       "mary@test.com",
		Password:    "Password123!",
		Username:    "mary",
		DisplayName: "MaryRose",
	}

	// reg
	t.Run("Register_Success", func(t *testing.T) {
		res, err := authService.Register(signUpReq)
		assert.NoError(t, err)

		assert.Equal(t, "mary", res.Profile.Username)
		assert.Equal(t, "MaryRose", res.Profile.DisplayName)
		assert.NotZero(t, res.Profile.AccountID)
		assert.NotZero(t, res.Token.AccessToken)
	})

	// login, same credential
	t.Run("Login_Success", func(t *testing.T) {
		loginReq := account.LoginRequest{
			Email:    signUpReq.Email,
			Password: signUpReq.Password,
		}

		token, err := authService.Login(loginReq)

		assert.NoError(t, err)

		assert.NotZero(t, token.AccessToken, "AccessToken должен быть сгенерирован")
		assert.NotZero(t, token.RefreshToken, "RefreshToken должен быть сгенерирован")
	})

	// login wrong password
	t.Run("Login_WrongPassword_Failure", func(t *testing.T) {
		loginReq := account.LoginRequest{
			Email:    signUpReq.Email,
			Password: "WrongPassword!",
		}

		_, err := authService.Login(loginReq)
		assert.Error(t, err)
	})
	// get public profile
	t.Run("Get_Public_Profile", func(t *testing.T) {
		pr, err := authService.GetPublicProfile("mary")

		assert.NoError(t, err)
		assert.Equal(t, "mary", pr.Username)
		assert.Equal(t, "MaryRose", pr.DisplayName)
	})
}
