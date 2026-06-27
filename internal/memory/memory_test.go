package memory

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	manager := NewManager(10, 100)

	stats := manager.GetStats()
	if stats["short_term_count"] != 0 {
		t.Errorf("expected 0 short-term, got %v", stats["short_term_count"])
	}
	if stats["long_term_count"] != 0 {
		t.Errorf("expected 0 long-term, got %v", stats["long_term_count"])
	}
	if stats["max_short_term"] != 10 {
		t.Errorf("expected max 10, got %v", stats["max_short_term"])
	}
}

func TestAddToShortTerm(t *testing.T) {
	manager := NewManager(3, 100)

	// Add 5 entries (should keep only last 3)
	for i := 0; i < 5; i++ {
		manager.AddToShortTerm(MemoryEntry{
			Type:    MemoryTypeDecision,
			Content: "decision",
		})
	}

	entries := manager.GetShortTerm(10)
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestAddToLongTerm(t *testing.T) {
	manager := NewManager(100, 3)

	// Add 5 entries (should keep only last 3)
	for i := 0; i < 5; i++ {
		manager.AddToLongTerm(MemoryEntry{
			Type:    MemoryTypeTrade,
			Content: "trade",
		})
	}

	entries := manager.GetLongTerm(10)
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestWorkingMemory(t *testing.T) {
	manager := NewManager(100, 100)

	// Set and get
	manager.SetWorkingMemory("key1", "value1")
	manager.SetWorkingMemory("key2", 42)

	val, ok := manager.GetWorkingMemory("key1")
	if !ok || val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}

	val, ok = manager.GetWorkingMemory("key2")
	if !ok || val != 42 {
		t.Errorf("expected 42, got %v", val)
	}

	// Non-existent key
	_, ok = manager.GetWorkingMemory("nonexistent")
	if ok {
		t.Error("expected false for non-existent key")
	}

	// Clear
	manager.ClearWorkingMemory()
	_, ok = manager.GetWorkingMemory("key1")
	if ok {
		t.Error("expected false after clear")
	}
}

func TestGetContext(t *testing.T) {
	manager := NewManager(100, 100)

	// Empty context
	context := manager.GetContext()
	if context == "" {
		t.Error("expected non-empty context")
	}

	// Add some entries
	manager.AddToShortTerm(MemoryEntry{
		Type:    MemoryTypeDecision,
		Content: "test decision",
	})

	manager.SetWorkingMemory("btc_price", 60000)

	context = manager.GetContext()
	if context == "" {
		t.Error("expected non-empty context")
	}
}
