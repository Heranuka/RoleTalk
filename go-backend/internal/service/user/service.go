// Package user implements the application service layer for user management.
// It handles business logic, coordinates transactions, and maps repository errors
// to service-level errors while providing observability through structured logging.
package user

import (
	"context"
	"errors"
	"fmt"
	"go-backend/internal/logger"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"go-backend/internal/models/domain"
	repouser "go-backend/internal/repository/user"
)

// tracer is the package-level OpenTelemetry tracer.
var tracer = otel.Tracer("internal/service/user")

// Service manages user business operations and coordinates with the repository layer.
type Service struct {
	userRepo   Repository
	transactor Transactor
	minio      Minio
	log        *zap.SugaredLogger
}

// NewService creates and returns a new User service instance with observability support.
func NewService(userRepo Repository, transactor Transactor, minio Minio, log *zap.SugaredLogger) *Service {
	return &Service{
		userRepo:   userRepo,
		transactor: transactor,
		minio:      minio,
		log:        log,
	}
}

// CreateWithPassword creates a new user via standard email and password registration.
func (s *Service) CreateWithPassword(ctx context.Context, input CreateWithPasswordInput) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "Service.CreateWithPassword")
	defer span.End()

	log := logger.FromContext(ctx, s.log)

	u, err := s.processDomainUser(ctx, input)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to process domain user")
		return uuid.Nil, err
	}

	id, err := s.saveUser(ctx, u)
	if err != nil {
		span.RecordError(err)
		return uuid.Nil, err
	}

	log.Infow("user created with password", "user_id", id, "email", u.Email)
	return id, nil
}

// CreateOAuthUser handles user creation through third-party identity providers.
func (s *Service) CreateOAuthUser(ctx context.Context, input CreateOAuthInput) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "Service.CreateOAuthUser")
	defer span.End()

	log := logger.FromContext(ctx, s.log)

	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" {
		return uuid.Nil, ErrUserInvalidData
	}

	u := domain.NewUserFromOAuth(
		email,
		input.DisplayName,
		input.Username,
		input.PhotoURL,
		input.IsEmailVerified,
	)

	id, err := s.saveUser(ctx, u)
	if err != nil {
		span.RecordError(err)
		return uuid.Nil, err
	}

	log.Infow("user created via oauth", "user_id", id, "email", email)
	return id, nil
}

// Update modifies an existing user's profile based on the provided input.
func (s *Service) Update(ctx context.Context, input UpdateInput) error {
	ctx, span := tracer.Start(ctx, "Service.Update")
	defer span.End()

	log := logger.FromContext(ctx, s.log)

	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		u, err := s.userRepo.GetByID(txCtx, input.ID)
		if err != nil {
			if errors.Is(err, repouser.ErrUserNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("fetch user for update: %w", err)
		}

		if input.DisplayName != nil {
			u.DisplayName = *input.DisplayName
		}
		if input.Username != nil {
			u.Username = input.Username
		}

		if input.InterfaceLang != nil {
			u.InterfaceLang = *input.InterfaceLang
		}
		if input.PracticeLang != nil {
			u.PracticeLang = *input.PracticeLang
		}

		if input.PhotoURL != nil {
			u.PhotoURL = input.PhotoURL // Normalized to PhotoURL in domain
		}
		if input.Role != nil {
			u.Role = *input.Role
		}

		if err := s.userRepo.Update(txCtx, u); err != nil {
			return s.mapRepoError(txCtx, err, "update user profile")
		}

		log.Infow("user profile updated", "user_id", u.ID)
		return nil
	})
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("fetch user for update: %w", err)
	}
	return nil
}

// SetNewPassword securely updates a user's password.
func (s *Service) SetNewPassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	ctx, span := tracer.Start(ctx, "Service.SetNewPassword")
	defer span.End()

	log := logger.FromContext(ctx, s.log)

	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		u, err := s.userRepo.GetByID(txCtx, id)
		if err != nil {
			if errors.Is(err, repouser.ErrUserNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("fetch user: %w", err)
		}

		if err := u.SetPassword(newPassword); err != nil {
			if errors.Is(err, domain.ErrPasswordTooLong) {
				return ErrPasswordTooLong
			}
			return fmt.Errorf("domain password logic: %w", err)
		}

		if err := s.userRepo.Update(txCtx, u); err != nil {
			return s.mapRepoError(txCtx, err, "update password")
		}

		log.Infow("user password changed", "user_id", id)
		return nil
	})
	if err != nil {
		return fmt.Errorf("set new password: %w", err)
	}
	return nil
}

// MarkEmailVerified updates the user's status to indicate email confirmation.
func (s *Service) MarkEmailVerified(ctx context.Context, id uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "Service.MarkEmailVerified")
	defer span.End()

	log := logger.FromContext(ctx, s.log)

	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		u, err := s.userRepo.GetByID(txCtx, id)
		if err != nil {
			if errors.Is(err, repouser.ErrUserNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("fetch user: %w", err)
		}

		u.IsEmailVerified = true

		if err := s.userRepo.Update(txCtx, u); err != nil {
			return s.mapRepoError(txCtx, err, "verify email")
		}

		log.Infow("user email verified", "user_id", id)
		return nil
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to mark email verified")
		return fmt.Errorf("mark email verified: %w", err)
	}
	return nil
}

// GetByID retrieves a user by their unique UUID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	ctx, span := tracer.Start(ctx, "Service.GetByID", trace.WithAttributes())
	defer span.End()

	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repouser.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by ID: %w", err)
	}
	return u, nil
}

// GetByEmail retrieves a user by their unique email address.
func (s *Service) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, span := tracer.Start(ctx, "Service.GetByEmail")
	defer span.End()

	cleanEmail := strings.TrimSpace(strings.ToLower(email))
	u, err := s.userRepo.GetByEmail(ctx, cleanEmail)
	if err != nil {
		if errors.Is(err, repouser.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

// IsAdmin checks if the given user has administrative rights.
func (s *Service) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	ctx, span := tracer.Start(ctx, "Service.IsAdmin")
	defer span.End()

	isAdmin, err := s.userRepo.IsAdmin(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("admin check: %w", err)
	}
	return isAdmin, nil
}

// saveUser is an internal helper to persist a user.
func (s *Service) saveUser(ctx context.Context, u *domain.User) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "Service.saveUser")
	defer span.End()

	id, err := s.userRepo.Create(ctx, u)
	if err != nil {
		return uuid.Nil, s.mapRepoError(ctx, err, "persist user")
	}
	return id, nil
}

// mapRepoError translates repository errors into service errors and records them in the current span.
func (s *Service) mapRepoError(ctx context.Context, err error, op string) error {
	log := logger.FromContext(ctx, s.log)
	span := trace.SpanFromContext(ctx)

	switch {
	case errors.Is(err, repouser.ErrUserEmailTaken):
		return ErrUserEmailTaken
	case errors.Is(err, repouser.ErrUserUsernameTaken):
		return ErrUserUsernameTaken
	case errors.Is(err, repouser.ErrUserInvalidRole):
		return ErrUserInvalidRole
	case errors.Is(err, repouser.ErrUserNotFound):
		return ErrUserNotFound
	default:
		span.RecordError(err)
		span.SetStatus(codes.Error, "unexpected repository error")
		log.Errorw("unexpected repository error", "operation", op, "error", err)
		return fmt.Errorf("%s: %w", op, err)
	}
}

// UploadAvatar handles receiving a file stream, uploading it to the S3 storage (MinIO),
// and updating the user's profile with the new photo URL.
func (s *Service) UploadAvatar(ctx context.Context, userID uuid.UUID, fileName string, file io.Reader) (string, error) {
	ctx, span := tracer.Start(ctx, "Service.User.UploadAvatar")
	defer span.End()

	log := logger.FromContext(ctx, s.log)

	// 1. Prepare metadata for tracing
	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("file.name", fileName),
	)

	// 2. Generate a unique filename using the userID and original extension
	// Example: avatars/550e8400-e29b-41d4-a716-446655440000.jpg
	ext := filepath.Ext(fileName)
	if ext == "" {
		ext = ".png" // Fallback extension
	}
	storedFileName := fmt.Sprintf("%s%s", userID.String(), ext)

	// 3. Upload to MinIO (subdir: "avatars")
	// Note: size is removed as the infra layer now handles streaming (-1)
	path, err := s.minio.Upload(ctx, "avatars", storedFileName, file)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "storage_upload_failed")
		log.Errorw("failed to upload avatar to storage", "user_id", userID, "error", err)
		return "", fmt.Errorf("storage upload: %w", err)
	}

	// 4. Update user record in database with the new PhotoURL
	// We use the Update method of the current service to ensure consistency
	updateInput := UpdateInput{
		ID:       userID,
		PhotoURL: &path,
	}

	if err := s.Update(ctx, updateInput); err != nil {
		// If DB update fails, we don't necessarily delete the file from S3 in MVP,
		// but we must log it and return an error.
		return "", fmt.Errorf("failed to update user profile with new photo: %w", err)
	}

	log.Infow("user avatar uploaded and profile updated", "user_id", userID, "path", path)
	return path, nil
}

// processDomainUser validates input data and initializes a new domain.User entity.
func (s *Service) processDomainUser(ctx context.Context, input CreateWithPasswordInput) (*domain.User, error) {
	_, span := tracer.Start(ctx, "Service.processDomainUser")
	defer span.End()

	email := strings.TrimSpace(strings.ToLower(input.Email))
	displayName := strings.TrimSpace(input.DisplayName)

	if email == "" {
		return nil, ErrUserInvalidData
	}

	if displayName == "" {
		displayName = strings.Split(email, "@")[0]
	}

	u, err := domain.NewUserWithPassword(
		email,
		input.Password,
		displayName,
		input.Username,
		input.PhotoURL,
	)

	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return u, nil
}
