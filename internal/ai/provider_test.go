package ai

import (
	"context"
	"testing"
)

func TestParseGenerateResult_Valid(t *testing.T) {
	raw := `{"caption": "hello", "script": [{"text": "hi", "start_seconds": 0, "duration_seconds": 2}]}`

	result, err := parseGenerateResult(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Caption != "hello" {
		t.Errorf("caption = %q, want %q", result.Caption, "hello")
	}
	if len(result.Script) != 1 || result.Script[0].Text != "hi" {
		t.Errorf("unexpected script: %+v", result.Script)
	}
}

func TestParseGenerateResult_InvalidJSON(t *testing.T) {
	_, err := parseGenerateResult("not json")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestRegistry_GetDefault(t *testing.T) {
	registry := NewRegistry("claude")
	stub := &stubProvider{}
	registry.Register("claude", stub)

	provider, err := registry.Get("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider != stub {
		t.Error("expected default provider to be returned")
	}
}

func TestRegistry_GetUnknown(t *testing.T) {
	registry := NewRegistry("claude")
	if _, err := registry.Get("nonexistent"); err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}
}

type stubProvider struct{}

func (s *stubProvider) Generate(_ context.Context, model, prompt string) (GenerateResult, error) {
	return GenerateResult{}, nil
}
