// Package ollama provides a client for interacting with the local Ollama LLM API.
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"go-backend/internal/logger"
)

var tracer = otel.Tracer("internal/infra/ollama")

// Client handles requests to the Ollama API with integrated observability.
type Client struct {
	url        string
	model      string
	httpClient *http.Client
	log        *zap.SugaredLogger
}

// NewClient creates a new Ollama client instance.
// Timeout is set to 60s as complex transcript analysis can be resource-intensive.
func NewClient(url, model string, log *zap.SugaredLogger) *Client {
	return &Client{
		url:   url,
		model: model,
		log:   log,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ChatRequest represents the payload for the Ollama chat API.
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Format   string    `json:"format,omitempty"` // "json" forces the model to output valid JSON
}

// Message represents a single turn in the AI conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse represents the response from the Ollama chat API.
type ChatResponse struct {
	Message Message `json:"message"`
}

// AnalyzeTranscript sends a prompt to the LLM and extracts skill scores from the response.
// It uses regex as a fallback to ensure JSON can be parsed even if the LLM adds conversational filler.
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
		Format: "json", // Instruct Ollama to enforce JSON output
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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Errorw("ollama returned error status", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("ollama error status: %d", resp.StatusCode)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Extract JSON using multi-line aware regex
	// (?s) flag allows the dot (.) to match newlines
	re := regexp.MustCompile(`(?s)\{.*\}`)
	cleanJSON := re.FindString(chatResp.Message.Content)
	if cleanJSON == "" {
		log.Errorw("LLM response contains no JSON object", "content", chatResp.Message.Content)
		return nil, fmt.Errorf("no valid JSON found in LLM response")
	}

	var scores map[string]int
	if err := json.Unmarshal([]byte(cleanJSON), &scores); err != nil {
		log.Errorw("failed to unmarshal scores", "json", cleanJSON, "error", err)
		return nil, fmt.Errorf("unmarshal scores: %w", err)
	}

	span.SetAttributes(
		attribute.String("llm.model", c.model),
		attribute.Int("scores.count", len(scores)),
	)

	return scores, nil
}
