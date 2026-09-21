package claude

import (
	"context"

	"github.com/alchemist/content-studio-api/internal/content"
)

type Client struct {
	apiKey string
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey}
}

func (c *Client) Generate(ctx context.Context, product, brief string) (string, []content.ScriptBeat, error) {
	prompt := BuildPrompt(product, brief)
	return callClaude(ctx, c.apiKey, prompt)
}
