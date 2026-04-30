// Package analytic implements business logic for user progress tracking and skill analysis.
package analytic

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

	"go-backend/internal/infra/prompt"
	"go-backend/internal/logger"
	"go-backend/internal/models/domain"
)

var tracer = otel.Tracer("internal/service/analytic")

// Service coordinates skill updates, retrieves performance metrics, and orchestrates AI evaluation.
type Service struct {
	repo            Repository
	message         MessageService
	practiceSession PracticeSessionRepository
	topicRepo       TopicRepository // Added to get the actual Goal
	llmClient       OllamaClient
	prompt          *prompt.Engine
	transactor      Transactor
	log             *zap.SugaredLogger
}

// NewService creates a new Analytic service instance with all dependencies.
func NewService(
	repo Repository,
	transactor Transactor,
	message MessageService,
	practiceSession PracticeSessionRepository,
	topicRepo TopicRepository,
	llmClient OllamaClient,
	promptEngine *prompt.Engine,
	log *zap.SugaredLogger,
) *Service {
	return &Service{
		repo:            repo,
		transactor:      transactor,
		message:         message,
		practiceSession: practiceSession,
		topicRepo:       topicRepo,
		llmClient:       llmClient,
		prompt:          promptEngine,
		log:             log,
	}
}

// EvaluateSession analyzes the completed dialog using the session's specific goal.
func (s *Service) EvaluateSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "Service.Analytic.EvaluateSession")
	defer span.End()

	log := s.logger(ctx)
	log.Infow("starting session evaluation", "session_id", sessionID, "user_id", userID)

	// 1. Fetch Session metadata to get TopicID
	session, err := s.practiceSession.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to fetch session: %w", err)
	}

	// 2. Fetch Topic metadata to get the specific Goal
	topic, err := s.topicRepo.GetByID(ctx, session.TopicID)
	if err != nil {
		return fmt.Errorf("failed to fetch topic context: %w", err)
	}

	// 3. Fetch dialog history
	messages, err := s.message.GetSessionHistory(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to fetch conversation history: %w", err)
	}

	// 4. Build transcript for the LLM
	var transcript strings.Builder
	for _, m := range messages {
		transcript.WriteString(fmt.Sprintf("%s: %s\n", m.SenderRole, m.TextContent))
	}

	// 5. Render prompt using the actual Topic Goal and Transcript
	evalParams := domain.EvaluationParams{
		Goal:       topic.Goal, // CRITICAL: Using the real business goal
		Transcript: transcript.String(),
	}

	evalPrompt, err := s.prompt.RenderEvaluation(evalParams)
	if err != nil {
		return fmt.Errorf("failed to render evaluation prompt: %w", err)
	}

	span.SetAttributes(
		attribute.String("session.goal", topic.Goal),
		attribute.Int("dialogue.turns", len(messages)),
	)

	// 6. Request and Parse AI Scores
	scores, err := s.requestAIScores(ctx, evalPrompt)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "ai_evaluation_failed")
		return fmt.Errorf("ai scoring logic failed: %w", err)
	}

	log.Infow("AI evaluation received",
		"empathy", scores.Empathy,
		"persuasion", scores.Persuasion,
	)

	// 7. Persist the progress using the existing ProcessSessionProgress logic
	return s.ProcessSessionProgress(ctx, userID, scores.Empathy, scores.Persuasion, scores.Structure, scores.Stress)
}

// GetUserSkills retrieves the current skill profile for a specific user.
// This method must match the handler's interface exactly.
func (s *Service) GetUserSkills(ctx context.Context, userID uuid.UUID) (*domain.UserSkill, error) {
	ctx, span := tracer.Start(ctx, "Service.Analytic.GetUserSkills")
	defer span.End()

	skills, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger(ctx).Errorw("failed to fetch user skills", "user_id", userID, "error", err)
		return nil, fmt.Errorf("fetch skills from repo: %w", err)
	}

	return skills, nil
}

// ProcessSessionProgress applies skill increments to the user's total profile in a transaction.
func (s *Service) ProcessSessionProgress(
	ctx context.Context,
	userID uuid.UUID,
	empInc, persInc, strucInc, stressInc int,
) error {
	ctx, span := tracer.Start(ctx, "Service.Analytic.ProcessSessionProgress")
	defer span.End()

	return s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		skills, err := s.repo.GetByUserID(txCtx, userID)
		if err != nil {
			return err
		}

		// Apply increments and clamp values 0-100 in the domain layer
		skills.ApplyProgress(empInc, persInc, strucInc, stressInc)

		if err := s.repo.UpdateSkills(txCtx, skills); err != nil {
			return err
		}

		return nil
	})
}

// requestAIScores performs the actual call to the LLM infrastructure.
func (s *Service) requestAIScores(ctx context.Context, promptStr string) (*sessionScores, error) {
	// We assume you have an 'llmClient' field in your Service struct
	// that points to the ollama.Client or an interface.

	rawScores, err := s.llmClient.AnalyzeTranscript(ctx, promptStr)
	if err != nil {
		return nil, err
	}

	// Map raw scores to our internal struct
	return &sessionScores{
		Empathy:    rawScores["empathy"],
		Persuasion: rawScores["persuasion"],
		Structure:  rawScores["structure"],
		Stress:     rawScores["stress"],
	}, nil
}

type sessionScores struct {
	Empathy    int `json:"empathy"`
	Persuasion int `json:"persuasion"`
	Structure  int `json:"structure"`
	Stress     int `json:"stress"`
}

func (s *Service) logger(ctx context.Context) *zap.SugaredLogger {
	return logger.FromContext(ctx, s.log)
}
