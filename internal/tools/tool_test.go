package tools

import (
	"context"
	"testing"
)

func TestNewSuccessResult(t *testing.T) {
	data := map[string]string{"key": "value"}
	result := NewSuccessResult(data)

	if !result.Success {
		t.Error("expected success to be true")
	}
	if result.Data == nil {
		t.Error("expected data to be set")
	}
	if result.Error != "" {
		t.Error("expected error to be empty")
	}
}

func TestNewErrorResult(t *testing.T) {
	errMsg := "test error"
	result := NewErrorResult(errMsg)

	if result.Success {
		t.Error("expected success to be false")
	}
	if result.Data != nil {
		t.Error("expected data to be nil")
	}
	if result.Error != errMsg {
		t.Errorf("expected error %q, got %q", errMsg, result.Error)
	}
}

func TestRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test empty registry
	if len(registry.List()) != 0 {
		t.Error("expected empty registry")
	}

	// Test get non-existent tool
	_, err := registry.Get("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent tool")
	}

	// Test register and get
	tool := &mockTool{name: "test_tool"}
	registry.Register(tool)

	got, err := registry.Get("test_tool")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got.Name() != "test_tool" {
		t.Errorf("expected name %q, got %q", "test_tool", got.Name())
	}

	// Test list
	if len(registry.List()) != 1 {
		t.Errorf("expected 1 tool, got %d", len(registry.List()))
	}

	// Test ToLLMTools
	llmTools := registry.ToLLMTools()
	if len(llmTools) != 1 {
		t.Errorf("expected 1 LLM tool, got %d", len(llmTools))
	}
	if llmTools[0].Name != "test_tool" {
		t.Errorf("expected name %q, got %q", "test_tool", llmTools[0].Name)
	}
}

type mockTool struct {
	name string
}

func (t *mockTool) Name() string {
	return t.name
}

func (t *mockTool) Description() string {
	return "mock tool"
}

func (t *mockTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"test": map[string]any{
				"type": "string",
			},
		},
	}
}

func (t *mockTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
	return NewSuccessResult("executed"), nil
}
