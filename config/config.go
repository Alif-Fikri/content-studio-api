package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port                string
	DatabaseURL         string
	SupabaseJWKSURL     string
	R2AccountID         string
	R2AccessKeyID       string
	R2SecretAccessKey   string
	R2Bucket            string
	AnthropicAPIKey     string
	MetaSystemUserToken string
	MetaAdAccountID     string
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
		MetaSystemUserToken: os.Getenv("META_SYSTEM_USER_TOKEN"),
		MetaAdAccountID:     os.Getenv("META_AD_ACCOUNT_ID"),
	}

	required := map[string]string{
		"DATABASE_URL":          cfg.DatabaseURL,
		"SUPABASE_JWKS_URL":     cfg.SupabaseJWKSURL,
		"R2_ACCOUNT_ID":         cfg.R2AccountID,
		"R2_ACCESS_KEY_ID":      cfg.R2AccessKeyID,
		"R2_SECRET_ACCESS_KEY":  cfg.R2SecretAccessKey,
		"R2_BUCKET":             cfg.R2Bucket,
		"ANTHROPIC_API_KEY":     cfg.AnthropicAPIKey,
		"META_SYSTEM_USER_TOKEN": cfg.MetaSystemUserToken,
		"META_AD_ACCOUNT_ID":    cfg.MetaAdAccountID,
	}

	for name, val := range required {
		if val == "" {
			return nil, fmt.Errorf("missing required env var: %s", name)
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
