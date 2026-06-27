package llm

import (
	"sync"
	"time"
)

// Session manages a conversation session with the LLM.
type Session struct {
	mu           sync.RWMutex
	messages     []Message
	maxMessages  int
	systemPrompt string
	lastUsed     time.Time
}

// NewSession creates a new LLM session.
func NewSession(systemPrompt string, maxMessages int) *Session {
	return &Session{
		messages:     make([]Message, 0),
		maxMessages:  maxMessages,
		systemPrompt: systemPrompt,
		lastUsed:     time.Now(),
	}
}

// AddMessage adds a message to the session.
func (s *Session) AddMessage(msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = append(s.messages, msg)
	s.lastUsed = time.Now()

	// Trim if exceeds max (keep system prompt)
	if len(s.messages) > s.maxMessages {
		// Keep the last N messages
		s.messages = s.messages[len(s.messages)-s.maxMessages:]
	}
}

// GetMessages returns all messages in the session.
func (s *Session) GetMessages() []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Build full message list with system prompt
	messages := make([]Message, 0, len(s.messages)+1)

	// Add system prompt
	if s.systemPrompt != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: s.systemPrompt,
		})
	}

	// Add conversation messages
	messages = append(messages, s.messages...)

	return messages
}

// Clear clears the session history.
func (s *Session) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = make([]Message, 0)
	s.lastUsed = time.Now()
}

// GetMessageCount returns the number of messages in the session.
func (s *Session) GetMessageCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.messages)
}

// LastUsed returns when the session was last used.
func (s *Session) LastUsed() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.lastUsed
}

// SessionManager manages multiple LLM sessions.
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	ttl      time.Duration
}

// NewSessionManager creates a new session manager.
func NewSessionManager(ttl time.Duration) *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
		ttl:      ttl,
	}

	// Start cleanup goroutine
	go sm.cleanup()

	return sm
}

// GetOrCreate gets or creates a session.
func (sm *SessionManager) GetOrCreate(id string, systemPrompt string, maxMessages int) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[id]
	if !ok {
		session = NewSession(systemPrompt, maxMessages)
		sm.sessions[id] = session
	}

	return session
}

// Get gets a session by ID.
func (sm *SessionManager) Get(id string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[id]
	return session, ok
}

// Delete deletes a session.
func (sm *SessionManager) Delete(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, id)
}

// cleanup removes expired sessions.
func (sm *SessionManager) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.mu.Lock()
		now := time.Now()
		for id, session := range sm.sessions {
			if now.Sub(session.LastUsed()) > sm.ttl {
				delete(sm.sessions, id)
			}
		}
		sm.mu.Unlock()
	}
}

// GetSessionCount returns the number of active sessions.
func (sm *SessionManager) GetSessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.sessions)
}
