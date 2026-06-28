package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/Emqo/TradingAgent/web/backend/models"
	"golang.org/x/crypto/bcrypt"
)

// MemoryUserStore is an in-memory user store for development.
type MemoryUserStore struct {
	mu    sync.RWMutex
	users map[string]*models.User
}

// NewMemoryUserStore creates a new memory user store.
func NewMemoryUserStore() *MemoryUserStore {
	return &MemoryUserStore{
		users: make(map[string]*models.User),
	}
}

// Init initializes the store (no-op for memory store).
func (s *MemoryUserStore) Init() error {
	return nil
}

// Create creates a new user.
func (s *MemoryUserStore) Create(username, password, email string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user exists
	if _, exists := s.users[username]; exists {
		return nil, fmt.Errorf("username already exists")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:           int64(len(s.users) + 1),
		Username:     username,
		PasswordHash: string(hash),
		Email:        email,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	s.users[username] = user
	return user, nil
}

// GetByUsername gets a user by username.
func (s *MemoryUserStore) GetByUsername(username string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// GetByID gets a user by ID.
func (s *MemoryUserStore) GetByID(id int64) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.ID == id {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

// UpdatePassword updates a user's password.
func (s *MemoryUserStore) UpdatePassword(id int64, newPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, user := range s.users {
		if user.ID == id {
			hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("hash password: %w", err)
			}
			user.PasswordHash = string(hash)
			user.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("user not found")
}

// VerifyPassword verifies a user's password.
func (s *MemoryUserStore) VerifyPassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}
