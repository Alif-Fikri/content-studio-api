package ai

import (
	"context"

	"github.com/alchemist/content-studio-api/internal/content"
)

func (r *Registry) Generate(ctx context.Context, providerName, model, prompt string) (string, []content.ScriptBeat, error) {
	provider, err := r.Get(providerName)
	if err != nil {
		return "", nil, err
	}

	result, err := provider.Generate(ctx, model, prompt)
	if err != nil {
		return "", nil, err
	}

	beats := make([]content.ScriptBeat, len(result.Script))
	for i, b := range result.Script {
		beats[i] = content.ScriptBeat{
			Text:            b.Text,
			StartSeconds:    b.StartSeconds,
			DurationSeconds: b.DurationSeconds,
		}
	}

	return result.Caption, beats, nil
}
