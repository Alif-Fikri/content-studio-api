package claude

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/alchemist/content-studio-api/internal/content"
)

type generationResult struct {
	Caption string       `json:"caption"`
	Script  []scriptBeat `json:"script"`
}

type scriptBeat struct {
	Text            string  `json:"text"`
	StartSeconds    float64 `json:"start_seconds"`
	DurationSeconds float64 `json:"duration_seconds"`
}

func callClaude(ctx context.Context, apiKey, prompt string) (string, []content.ScriptBeat, error) {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeOpus5,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", nil, err
	}

	var result generationResult
	for _, block := range message.Content {
		if block.Type == "text" {
			if err := json.Unmarshal([]byte(block.Text), &result); err != nil {
				return "", nil, fmt.Errorf("parse claude response: %w", err)
			}
			break
		}
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
