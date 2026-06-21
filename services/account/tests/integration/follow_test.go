package integration_test

import (
	"account/internal/account"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"
)

func setupDB(t *testing.T) *sql.DB {
	db, err := sql.Open(
		"postgres",
		"postgres://test:test@localhost:5432/testdb?sslmode=disable",
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

func TestFollowFlow(t *testing.T) {
	// Arrange
	db := setupDB(t)
	cleanDB(t, db)

	repo := account.NewRepository(db)
	service := account.NewFollowService(repo)

	var followerID int64
	var followingID int64

	err := db.QueryRow(`
		INSERT INTO accounts (email, password_hash)
		VALUES ('a@test.com', 'x')
		RETURNING id
	`).Scan(&followerID)
	require.NoError(t, err)

	err = db.QueryRow(`
		INSERT INTO accounts (email, password_hash)
		VALUES ('b@test.com', 'x')
		RETURNING id
	`).Scan(&followingID)
	require.NoError(t, err)

	// Act
	err = service.Follow(followerID, followingID)

	// Assert
	require.NoError(t, err)

	var count int

	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM follows
		WHERE follower_id = $1
		AND following_id = $2
	`, followerID, followingID).Scan(&count)

	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSelfFollow(t *testing.T) {
	db := setupDB(t)
	cleanDB(t, db)

	repo := account.NewRepository(db)
	service := account.NewFollowService(repo)

	err := service.Follow(1, 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, account.ErrSelfFollowing)
}

func TestUnfollowFlow(t *testing.T) {
	db := setupDB(t)
	cleanDB(t, db)

	repo := account.NewRepository(db)
	service := account.NewFollowService(repo)

	var followerID int64
	var followingID int64

	require.NoError(t,
		db.QueryRow(`
		INSERT INTO accounts (email, password_hash)
		VALUES ('a@test.com','x')
		RETURNING id
	`).Scan(&followerID))

	require.NoError(t,
		db.QueryRow(`
		INSERT INTO accounts (email, password_hash)
		VALUES ('b@test.com','x')
		RETURNING id
	`).Scan(&followingID))

	require.NoError(t, service.Follow(followerID, followingID))

	err := service.Unfollow(followerID, followingID)

	require.NoError(t, err)

	var count int

	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM follows
		WHERE follower_id = $1
		AND following_id = $2
	`, followerID, followingID).Scan(&count)

	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
