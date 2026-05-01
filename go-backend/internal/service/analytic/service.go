// Package analytic implements business logic for user progress tracking and skill analysis.
package analytic

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"go-backend/internal/logger"
	"go-backend/internal/models/domain"
)

var tracer = otel.Tracer("internal/service/analytic")

// Service coordinates skill updates, retrieves performance metrics, and orchestrates AI evaluation.
type Service struct {
	repo            Repository
	message         MessageService
	practiceSession PracticeSessionRepository
	topicRepo       TopicRepository
	llmClient       OllamaClient
	prompt          Engine
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
	promptEngine Engine,
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
// It fetches history, renders an evaluation prompt, and updates user skill points.
func (s *Service) EvaluateSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "Service.EvaluateSession")
	defer span.End()

	log := logger.FromContext(ctx, s.log)

	// 1. Fetch Session & Topic context
	session, err := s.practiceSession.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session lookup: %w", err)
	}

	topic, err := s.topicRepo.GetByID(ctx, session.TopicID)
	if err != nil {
		return fmt.Errorf("topic lookup: %w", err)
	}

	// 2. Fetch dialog history
	messages, err := s.message.GetSessionHistory(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("history lookup: %w", err)
	}

	if len(messages) < 2 {
		log.Warnw("session too short for evaluation", "session_id", sessionID)
		return nil // Nothing to evaluate
	}

	// 3. Build transcript safely (handling pointer strings)
	var transcript strings.Builder
	for _, m := range messages {
		content := "[Audio Only]"
		if m.TextContent != nil {
			content = *m.TextContent
		}
		fmt.Fprintf(&transcript, "%s: %s\n", m.SenderRole, content)
	}

	// 4. Request AI Scores
	evalParams := domain.EvaluationParams{
		Goal:       topic.Goal,
		Transcript: transcript.String(),
	}

	evalPrompt, err := s.prompt.RenderEvaluation(evalParams)
	if err != nil {
		return fmt.Errorf("prompt render: %w", err)
	}

	scores, err := s.requestAIScores(ctx, evalPrompt)
	if err != nil {
		return fmt.Errorf("ai scoring: %w", err)
	}

	// 5. Atomic Progress Update
	return s.ProcessSessionProgress(ctx, userID, scores.Empathy, scores.Persuasion, scores.Structure, scores.Stress)
}

// ProcessSessionProgress applies skill increments to the user's profile within a transaction.
func (s *Service) ProcessSessionProgress(ctx context.Context, userID uuid.UUID, emp, pers, struc, stress int) error {
	if err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		skills, err := s.repo.GetByUserID(txCtx, userID)
		if err != nil {
			// If no skills exist yet, initialize them
			skills = domain.NewUserSkill(userID)
		}

		skills.ApplyProgress(emp, pers, struc, stress)

		return s.repo.UpdateSkills(txCtx, skills)
	}); err != nil {
		return fmt.Errorf("skill update transaction: %w", err)
	}
	return nil
}

// GetUserSkills retrieves the current skill profile for a specific user.
// This method must match the handler's interface exactly.
func (s *Service) GetUserSkills(ctx context.Context, userID uuid.UUID) (*domain.UserSkill, error) {
	ctx, span := tracer.Start(ctx, "Service.Analytic.GetUserSkills")
	defer span.End()

	log := logger.FromContext(ctx, s.log)

	skills, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		log.Errorw("failed to fetch user skills", "user_id", userID, "error", err)
		return nil, fmt.Errorf("fetch skills from repo: %w", err)
	}

	return skills, nil
}

func (s *Service) requestAIScores(ctx context.Context, promptStr string) (*sessionScores, error) {
	raw, err := s.llmClient.AnalyzeTranscript(ctx, promptStr)
	if err != nil {
		return nil, fmt.Errorf("ollama analysis failed: %w", err)
	}

	// Ensure keys match exactly what Python/Ollama returns in its JSON
	return &sessionScores{
		Empathy:    raw["empathy"],
		Persuasion: raw["persuasion"],
		Structure:  raw["structure"],
		Stress:     raw["stress_resistance"], // Updated to match domain naming
	}, nil
}
