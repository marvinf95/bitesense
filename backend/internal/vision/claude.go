package vision

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	claudeModel    = "claude-sonnet-4-6"
	claudeEndpoint = "https://api.anthropic.com/v1/messages"
)

type ClaudeClient struct {
	APIKey string
	HTTP   *http.Client
}

func NewClaudeClient(apiKey string) *ClaudeClient {
	return &ClaudeClient{
		APIKey: apiKey,
		HTTP:   &http.Client{Timeout: 30 * time.Second},
	}
}

type claudeImageSource struct {
	Type      string `json:"type"`       // "base64"
	MediaType string `json:"media_type"` // image/jpeg etc.
	Data      string `json:"data"`
}

type claudeContentBlock struct {
	Type   string             `json:"type"` // "text" | "image"
	Text   string             `json:"text,omitempty"`
	Source *claudeImageSource `json:"source,omitempty"`
}

type claudeMessage struct {
	Role    string               `json:"role"`
	Content []claudeContentBlock `json:"content"`
}

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system,omitempty"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func (c *ClaudeClient) Analyze(ctx context.Context, imageBytes []byte, mimeType string) (*Result, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("anthropic api key missing")
	}
	body := claudeRequest{
		Model:     claudeModel,
		MaxTokens: 1024,
		System:    "You are a careful food-recognition assistant. Return strictly minified JSON.",
		Messages: []claudeMessage{
			{
				Role: "user",
				Content: []claudeContentBlock{
					{Type: "text", Text: jsonSchemaInstruction},
					{Type: "image", Source: &claudeImageSource{
						Type:      "base64",
						MediaType: mimeType,
						Data:      base64.StdEncoding.EncodeToString(imageBytes),
					}},
				},
			},
		},
	}
	buf, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, claudeEndpoint, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("claude status %d: %s", resp.StatusCode, raw)
	}
	var cr claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return nil, err
	}
	if len(cr.Content) == 0 {
		return nil, fmt.Errorf("claude empty response")
	}
	var result Result
	if err := json.Unmarshal([]byte(cr.Content[0].Text), &result); err != nil {
		return nil, fmt.Errorf("claude parse: %w", err)
	}
	result.Provider = "claude"
	return &result, nil
}
