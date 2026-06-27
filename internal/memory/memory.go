package memory

import (
	"fmt"
	"sync"
	"time"
)

// Manager manages the agent's memory system.
type Manager struct {
	mu            sync.RWMutex
	shortTerm     []MemoryEntry
	longTerm      []MemoryEntry
	working       map[string]any
	maxShortTerm  int
	maxLongTerm   int
}

// MemoryEntry represents a single memory entry.
type MemoryEntry struct {
	ID        string         `json:"id"`
	Type      MemoryType     `json:"type"`
	Content   string         `json:"content"`
	Data      map[string]any `json:"data,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	TTL       time.Duration  `json:"ttl,omitempty"`
	Tags      []string       `json:"tags,omitempty"`
}

// MemoryType represents the type of memory.
type MemoryType string

const (
	MemoryTypeDecision   MemoryType = "DECISION"
	MemoryTypeObservation MemoryType = "OBSERVATION"
	MemoryTypeAnalysis   MemoryType = "ANALYSIS"
	MemoryTypeTrade      MemoryType = "TRADE"
	MemoryTypeRisk       MemoryType = "RISK"
	MemoryTypeStrategy   MemoryType = "STRATEGY"
	MemoryTypeReflection MemoryType = "REFLECTION"
)

// NewManager creates a new memory manager.
func NewManager(maxShortTerm, maxLongTerm int) *Manager {
	return &Manager{
		shortTerm:    make([]MemoryEntry, 0),
		longTerm:     make([]MemoryEntry, 0),
		working:      make(map[string]any),
		maxShortTerm: maxShortTerm,
		maxLongTerm:  maxLongTerm,
	}
}

// AddToShortTerm adds an entry to short-term memory.
func (m *Manager) AddToShortTerm(entry MemoryEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry.Timestamp = time.Now()
	m.shortTerm = append(m.shortTerm, entry)

	// Trim if exceeds max
	if len(m.shortTerm) > m.maxShortTerm {
		m.shortTerm = m.shortTerm[len(m.shortTerm)-m.maxShortTerm:]
	}
}

// AddToLongTerm adds an entry to long-term memory.
func (m *Manager) AddToLongTerm(entry MemoryEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry.Timestamp = time.Now()
	m.longTerm = append(m.longTerm, entry)

	// Trim if exceeds max
	if len(m.longTerm) > m.maxLongTerm {
		m.longTerm = m.longTerm[len(m.longTerm)-m.maxLongTerm:]
	}
}

// GetShortTerm returns recent short-term memories.
func (m *Manager) GetShortTerm(limit int) []MemoryEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.shortTerm) {
		limit = len(m.shortTerm)
	}

	start := len(m.shortTerm) - limit
	if start < 0 {
		start = 0
	}

	return m.shortTerm[start:]
}

// GetLongTerm returns long-term memories.
func (m *Manager) GetLongTerm(limit int) []MemoryEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.longTerm) {
		limit = len(m.longTerm)
	}

	start := len(m.longTerm) - limit
	if start < 0 {
		start = 0
	}

	return m.longTerm[start:]
}

// SetWorkingMemory sets a value in working memory.
func (m *Manager) SetWorkingMemory(key string, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.working[key] = value
}

// GetWorkingMemory gets a value from working memory.
func (m *Manager) GetWorkingMemory(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.working[key]
	return val, ok
}

// ClearWorkingMemory clears all working memory.
func (m *Manager) ClearWorkingMemory() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.working = make(map[string]any)
}

// GetContext returns a formatted context string for the LLM.
func (m *Manager) GetContext() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	context := "=== Memory Context ===\n\n"

	// Recent decisions
	if len(m.shortTerm) > 0 {
		context += "Recent Decisions:\n"
		recent := m.shortTerm
		if len(recent) > 5 {
			recent = recent[len(recent)-5:]
		}
		for _, entry := range recent {
			context += "- [" + entry.Timestamp.Format("15:04:05") + "] " + entry.Content + "\n"
		}
		context += "\n"
	}

	// Working memory
	if len(m.working) > 0 {
		context += "Current State:\n"
		for key, val := range m.working {
			context += "- " + key + ": " + formatValue(val) + "\n"
		}
		context += "\n"
	}

	return context
}

// GetStats returns memory statistics.
func (m *Manager) GetStats() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]any{
		"short_term_count": len(m.shortTerm),
		"long_term_count":  len(m.longTerm),
		"working_count":    len(m.working),
		"max_short_term":   m.maxShortTerm,
		"max_long_term":    m.maxLongTerm,
	}
}

func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return formatFloat(val, 2)
	case int:
		return formatInt(val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return "..."
	}
}

func formatFloat(f float64, decimals int) string {
	return fmt.Sprintf("%.*f", decimals, f)
}

func formatInt(i int) string {
	return fmt.Sprintf("%d", i)
}
