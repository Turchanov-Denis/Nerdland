package e2e

import (
	"account/internal/account"
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"
)

type TestUser struct {
	Username    string
	AccessToken string
}

func setupDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open(
		"postgres",
		"postgres://test:test@localhost:5433/testdb?sslmode=disable",
	)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	return db
}

func cleanDB(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec("DELETE FROM follows")
	require.NoError(t, err)

	_, err = db.Exec("DELETE FROM sessions")
	require.NoError(t, err)

	_, err = db.Exec("DELETE FROM profiles")
	require.NoError(t, err)

	_, err = db.Exec("DELETE FROM accounts")
	require.NoError(t, err)
}

func postRegister(t *testing.T, username string) TestUser {
	t.Helper()

	reqBody := account.RegisterRequest{
		Email:       username + "@test.com",
		Password:    "Password123!",
		Username:    username,
		DisplayName: username,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	resp, err := http.Post(
		"http://localhost:8080/auth/register",
		"application/json",
		bytes.NewBuffer(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var ans account.RegisterResponse

	err = json.NewDecoder(resp.Body).Decode(&ans)
	require.NoError(t, err)

	return TestUser{
		Username:    username,
		AccessToken: ans.Token.AccessToken,
	}
}

func postFollow(
	t *testing.T,
	token string,
	username string,
) {
	t.Helper()

	req, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:8080/follow/"+username,
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)

	client := http.Client{}

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func getFollowers(
	t *testing.T,
	username string,
	token string,
) []account.PublicProfile {

	t.Helper()

	req, err := http.NewRequest(
		http.MethodGet,
		"http://localhost:8080/followers/"+username,
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)

	client := http.Client{}

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var profiles []account.PublicProfile

	err = json.NewDecoder(resp.Body).Decode(&profiles)
	require.NoError(t, err)

	return profiles
}

func getFollowing(
	t *testing.T,
	username string,
	token string,
) []account.PublicProfile {

	t.Helper()

	req, err := http.NewRequest(
		http.MethodGet,
		"http://localhost:8080/following/"+username,
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)

	client := http.Client{}

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var profiles []account.PublicProfile

	err = json.NewDecoder(resp.Body).Decode(&profiles)
	require.NoError(t, err)

	return profiles
}

func TestRegistrationFollow(t *testing.T) {
	testDB := setupDB(t)
	t.Cleanup(func() {
		cleanDB(t, testDB)
	})
	alice := postRegister(t, "alice")
	bob := postRegister(t, "bob")
	charlie := postRegister(t, "charlie")
	david := postRegister(t, "david")
	eve := postRegister(t, "eve")

	postFollow(t, alice.AccessToken, bob.Username)
	postFollow(t, alice.AccessToken, charlie.Username)
	postFollow(t, alice.AccessToken, david.Username)
	postFollow(t, alice.AccessToken, eve.Username)

	followersBob := getFollowers(t, bob.Username, bob.AccessToken)
	require.Len(t, followersBob, 1)
	require.Equal(t, "alice", followersBob[0].Username)

	followingAlice := getFollowing(t, alice.Username, alice.AccessToken)
	require.Len(t, followingAlice, 4)
}
