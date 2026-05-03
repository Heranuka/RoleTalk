// Package ai implements the gRPC client for interacting with the AI microservice.
package ai

import (
	"context"
	"fmt"
	"time"

	aiv1 "go-backend/api/proto/ai/v1" // Путь к сгенерированному коду

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client wraps the gRPC connection and the generated service client.
type Client struct {
	conn   *grpc.ClientConn
	client aiv1.AIServiceClient
	log    *zap.SugaredLogger
}

// NewClient initializes a new gRPC connection to the AI microservice.
// It uses insecure credentials for local/internal communication and
// attaches OpenTelemetry interceptors for distributed tracing.
func NewClient(addr string, log *zap.SugaredLogger) (*Client, error) {
	// otelgrpc.NewClientHandler() ensures that every gRPC call is tracked in Tempo.
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client for %s: %w", addr, err)
	}

	return &Client{
		conn:   conn,
		client: aiv1.NewAIServiceClient(conn),
		log:    log,
	}, nil
}

// ProcessVoiceTurn invokes the remote AI processing logic.
// It converts raw audio and context into transcription and synthesized AI response.
func (c *Client) ProcessVoiceTurn(
	ctx context.Context,
	audio []byte,
	lang string,
	systemPrompt string,
) (userText string, aiText string, aiAudio []byte, err error) {
	// Set a deadline for the gRPC call to prevent resource hanging.
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	req := &aiv1.ProcessVoiceTurnRequest{
		AudioData:    audio,
		Language:     lang,
		SystemPrompt: systemPrompt,
		RequestId:    "", // Trace ID is handled by otelgrpc interceptor automatically
	}

	c.log.Infow("[gRPC] Sending voice turn to AI service",
		"audio_size", len(audio),
		"language", lang,
	)

	resp, err := c.client.ProcessVoiceTurn(ctx, req)
	if err != nil {
		c.log.Errorw("[gRPC] Failed to process voice turn", "error", err)
		return "", "", nil, fmt.Errorf("grpc call failed: %w", err)
	}

	return resp.UserTranscription, resp.AiResponseText, resp.AiAudioData, nil
}

// Close gracefully shuts down the gRPC connection.
func (c *Client) Close() error {
	c.log.Info("[gRPC] Closing AI service connection...")
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close gRPC connection: %w", err)
	}
	return nil
}
