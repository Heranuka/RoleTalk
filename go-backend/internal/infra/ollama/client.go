// Package ollama provides a client for interacting with the local Ollama LLM API.
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"go-backend/internal/logger"
)

var tracer = otel.Tracer("internal/infra/ollama")

// Client handles requests to the Ollama API.
type Client struct {
	url        string
	model      string
	httpClient *http.Client
	log        *zap.SugaredLogger
}

// NewClient creates a new Ollama client instance.
func NewClient(url, model string, log *zap.SugaredLogger) *Client {
	return &Client{
		url:   url,
		model: model,
		log:   log,
		httpClient: &http.Client{
			Timeout: 45 * time.Second, // Evaluations can take time
		},
	}
}

// ChatRequest represents the payload for the Ollama chat API.
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse represents the response from the Ollama chat API.
type ChatResponse struct {
	Message Message `json:"message"`
}

// AnalyzeTranscript sends a prompt to the LLM and parses the resulting JSON.
func (c *Client) AnalyzeTranscript(ctx context.Context, prompt string) (map[string]int, error) {
	ctx, span := tracer.Start(ctx, "Ollama.AnalyzeTranscript")
	defer span.End()

	log := logger.FromContext(ctx, c.log)

	reqBody := ChatRequest{
		Model: c.model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		Stream: false,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/api/chat", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("http execute: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status: %d", resp.StatusCode)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// 1. Extract JSON from LLM response using Regex
	// LLMs often wrap JSON in ```json ... ``` blocks.
	re := regexp.MustCompile(`\{.*\}`)
	cleanJSON := re.FindString(chatResp.Message.Content)
	if cleanJSON == "" {
		log.Errorw("LLM failed to return valid JSON", "raw_content", chatResp.Message.Content)
		return nil, fmt.Errorf("no valid JSON found in LLM response")
	}

	// 2. Parse the cleaned JSON into a map
	var scores map[string]int
	if err := json.Unmarshal([]byte(cleanJSON), &scores); err != nil {
		return nil, fmt.Errorf("unmarshal scores: %w", err)
	}

	span.SetAttributes(attribute.Int("response.length", len(chatResp.Message.Content)))
	return scores, nil
}
