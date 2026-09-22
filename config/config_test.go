package config

import (
	"os"
	"testing"
)

func setValidEnv(t *testing.T) {
	t.Helper()
	env := map[string]string{
		"DATABASE_URL":           "postgres://localhost/test",
		"SUPABASE_JWKS_URL":      "https://example.supabase.co/auth/v1/.well-known/jwks.json",
		"R2_ACCOUNT_ID":          "account",
		"R2_ACCESS_KEY_ID":       "key",
		"R2_SECRET_ACCESS_KEY":   "secret",
		"R2_BUCKET":              "bucket",
		"META_SYSTEM_USER_TOKEN": "token",
		"META_AD_ACCOUNT_ID":     "act_123",
		"ANTHROPIC_API_KEY":      "sk-ant",
	}
	for k, v := range env {
		t.Setenv(k, v)
	}
}

func TestLoad_Valid(t *testing.T) {
	setValidEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.RenderedRetentionDays != 7 {
		t.Errorf("expected default retention 7, got %d", cfg.RenderedRetentionDays)
	}
}

func TestLoad_MissingRequiredVar(t *testing.T) {
	setValidEnv(t)
	os.Unsetenv("DATABASE_URL")

	if _, err := Load(); err == nil {
		t.Fatal("expected error for missing DATABASE_URL, got nil")
	}
}

func TestLoad_NoAIProviderKey(t *testing.T) {
	setValidEnv(t)
	os.Unsetenv("ANTHROPIC_API_KEY")

	if _, err := Load(); err == nil {
		t.Fatal("expected error when no AI provider key is set, got nil")
	}
}

func TestLoad_NegativeRetentionDays(t *testing.T) {
	setValidEnv(t)
	t.Setenv("RENDERED_RETENTION_DAYS", "-1")

	if _, err := Load(); err == nil {
		t.Fatal("expected error for negative retention days, got nil")
	}
}

func TestLoad_CORSOrigins(t *testing.T) {
	setValidEnv(t)
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.com, https://b.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.CORSAllowedOrigins) != 2 {
		t.Fatalf("expected 2 origins, got %v", cfg.CORSAllowedOrigins)
	}
	if cfg.CORSAllowedOrigins[0] != "https://a.com" || cfg.CORSAllowedOrigins[1] != "https://b.com" {
		t.Errorf("unexpected origins: %v", cfg.CORSAllowedOrigins)
	}
}
