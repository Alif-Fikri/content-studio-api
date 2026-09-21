package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port                  string
	DatabaseURL           string
	SupabaseJWKSURL       string
	R2AccountID           string
	R2AccessKeyID         string
	R2SecretAccessKey     string
	R2Bucket              string
	AnthropicAPIKey       string
	OpenAIAPIKey          string
	GeminiAPIKey          string
	DefaultAIProvider     string
	MetaSystemUserToken   string
	MetaAdAccountID       string
	RenderedRetentionDays int
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		SupabaseJWKSURL:     os.Getenv("SUPABASE_JWKS_URL"),
		R2AccountID:         os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyID:       os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey:   os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2Bucket:            os.Getenv("R2_BUCKET"),
		AnthropicAPIKey:     os.Getenv("ANTHROPIC_API_KEY"),
		OpenAIAPIKey:        os.Getenv("OPENAI_API_KEY"),
		GeminiAPIKey:        os.Getenv("GEMINI_API_KEY"),
		DefaultAIProvider:   getEnv("DEFAULT_AI_PROVIDER", "claude"),
		MetaSystemUserToken: os.Getenv("META_SYSTEM_USER_TOKEN"),
		MetaAdAccountID:     os.Getenv("META_AD_ACCOUNT_ID"),
	}

	retentionDays, err := strconv.Atoi(getEnv("RENDERED_RETENTION_DAYS", "7"))
	if err != nil {
		return nil, fmt.Errorf("invalid RENDERED_RETENTION_DAYS: %w", err)
	}
	if retentionDays <= 0 {
		return nil, fmt.Errorf("RENDERED_RETENTION_DAYS must be a positive number of days, got %d", retentionDays)
	}
	cfg.RenderedRetentionDays = retentionDays

	required := map[string]string{
		"DATABASE_URL":           cfg.DatabaseURL,
		"SUPABASE_JWKS_URL":      cfg.SupabaseJWKSURL,
		"R2_ACCOUNT_ID":          cfg.R2AccountID,
		"R2_ACCESS_KEY_ID":       cfg.R2AccessKeyID,
		"R2_SECRET_ACCESS_KEY":   cfg.R2SecretAccessKey,
		"R2_BUCKET":              cfg.R2Bucket,
		"META_SYSTEM_USER_TOKEN": cfg.MetaSystemUserToken,
		"META_AD_ACCOUNT_ID":     cfg.MetaAdAccountID,
	}

	for name, val := range required {
		if val == "" {
			return nil, fmt.Errorf("missing required env var: %s", name)
		}
	}

	if cfg.AnthropicAPIKey == "" && cfg.OpenAIAPIKey == "" && cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("at least one AI provider key must be set: ANTHROPIC_API_KEY, OPENAI_API_KEY, or GEMINI_API_KEY")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
