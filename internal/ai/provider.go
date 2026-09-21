package ai

import (
	"context"
	"encoding/json"
	"fmt"
)

type ScriptBeat struct {
	Text            string  `json:"text"`
	StartSeconds    float64 `json:"start_seconds"`
	DurationSeconds float64 `json:"duration_seconds"`
}

type GenerateResult struct {
	Caption string       `json:"caption"`
	Script  []ScriptBeat `json:"script"`
}

type Provider interface {
	Generate(ctx context.Context, model, prompt string) (GenerateResult, error)
}

type Registry struct {
	providers       map[string]Provider
	defaultProvider string
}

func NewRegistry(defaultProvider string) *Registry {
	return &Registry{
		providers:       make(map[string]Provider),
		defaultProvider: defaultProvider,
	}
}

func (r *Registry) Register(name string, provider Provider) {
	r.providers[name] = provider
}

func (r *Registry) Get(name string) (Provider, error) {
	if name == "" {
		name = r.defaultProvider
	}
	provider, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("unknown ai provider: %s", name)
	}
	return provider, nil
}

func parseGenerateResult(raw string) (GenerateResult, error) {
	var result GenerateResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return GenerateResult{}, fmt.Errorf("parse ai response as json: %w", err)
	}
	return result, nil
}
