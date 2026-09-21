package ai

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type ClaudeProvider struct {
	client anthropic.Client
}

func NewClaudeProvider(apiKey string) *ClaudeProvider {
	return &ClaudeProvider{
		client: anthropic.NewClient(option.WithAPIKey(apiKey)),
	}
}

func (p *ClaudeProvider) Generate(ctx context.Context, model, prompt string) (GenerateResult, error) {
	if model == "" {
		model = "claude-opus-5"
	}

	message, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return GenerateResult{}, err
	}

	var raw string
	for _, block := range message.Content {
		if block.Type == "text" {
			raw = block.Text
			break
		}
	}

	return parseGenerateResult(raw)
}
