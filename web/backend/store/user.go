package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Emqo/TradingAgent/web/backend/models"
	"golang.org/x/crypto/bcrypt"
)

// UserStore manages user data.
type UserStore struct {
	db *sql.DB
}

// NewUserStore creates a new user store.
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

// Init creates the users table if it doesn't exist.
func (s *UserStore) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		email VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("create users table: %w", err)
	}

	return nil
}

// Create creates a new user.
func (s *UserStore) Create(username, password, email string) (*models.User, error) {
	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Insert user
	query := `
	INSERT INTO users (username, password_hash, email, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, username, email, created_at, updated_at`

	now := time.Now()
	user := &models.User{}
	err = s.db.QueryRow(query, username, string(hash), email, now, now).Scan(
		&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return user, nil
}

// GetByUsername gets a user by username.
func (s *UserStore) GetByUsername(username string) (*models.User, error) {
	query := `
	SELECT id, username, password_hash, email, created_at, updated_at
	FROM users
	WHERE username = $1`

	user := &models.User{}
	err := s.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Email,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}

// GetByID gets a user by ID.
func (s *UserStore) GetByID(id int64) (*models.User, error) {
	query := `
	SELECT id, username, email, created_at, updated_at
	FROM users
	WHERE id = $1`

	user := &models.User{}
	err := s.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}

// UpdatePassword updates a user's password.
func (s *UserStore) UpdatePassword(id int64, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	query := `UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`
	_, err = s.db.Exec(query, string(hash), time.Now(), id)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}

// VerifyPassword verifies a user's password.
func (s *UserStore) VerifyPassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}
