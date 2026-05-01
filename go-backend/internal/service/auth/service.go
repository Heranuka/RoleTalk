// Package auth implements authentication, registration, and session management logic.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"go-backend/internal/config"
	"go-backend/internal/logger"
	"go-backend/internal/models/domain"
	tokenrepo "go-backend/internal/repository/verificationtoken"
	"go-backend/internal/service/user"
)

var tracer = otel.Tracer("internal/service/auth")

// Service provides authentication and identity operations.
type Service struct {
	userService UserService
	sessionRepo SessionRepository
	oauthRepo   OAuthConnectionRepository
	tokenRepo   VerificationTokenRepository
	emailSender EmailSender
	oauthClient OAuthClient
	transactor  Transactor
	authConfig  *config.Auth
	log         *zap.SugaredLogger
}

// NewService creates a new authentication service instance.
func NewService(
	userService UserService,
	sessionRepo SessionRepository,
	oauthRepo OAuthConnectionRepository,
	tokenRepo VerificationTokenRepository,
	emailSender EmailSender,
	oauthClient OAuthClient,
	transactor Transactor,
	authConfig *config.Auth,
	log *zap.SugaredLogger,
) *Service {
	return &Service{
		userService: userService,
		sessionRepo: sessionRepo,
		oauthRepo:   oauthRepo,
		tokenRepo:   tokenRepo,
		emailSender: emailSender,
		oauthClient: oauthClient,
		transactor:  transactor,
		authConfig:  authConfig,
		log:         log,
	}
}

// Register creates a new user, generates a verification token, and sends a welcome email.
func (s *Service) Register(ctx context.Context, input RegisterInput) error {
	ctx, span := tracer.Start(ctx, "Service.Register")
	defer span.End()

	log := logger.FromContext(ctx, s.log)
	var rawToken string

	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		// 1. Create the user
		userID, err := s.userService.CreateWithPassword(txCtx, user.CreateWithPasswordInput{
			Email:       input.Email,
			Password:    input.Password,
			Username:    input.Username,
			DisplayName: input.Email, // Default DisplayName to Email for initial registration
		})
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// 2. Generate verification token
		tokenStr, entity, err := domain.NewVerificationToken(userID, domain.TokenPurposeEmailVerification, s.authConfig.EmailVerificationTTL)
		if err != nil {
			return fmt.Errorf("generate token: %w", err)
		}

		// 3. Save token hash to DB
		if err := s.tokenRepo.Create(txCtx, entity); err != nil {
			return fmt.Errorf("save token: %w", err)
		}

		rawToken = tokenStr
		return nil
	})

	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("create token: %w", err)
	}

	// 4. Send email (outside the DB transaction)
	if err := s.emailSender.SendVerificationEmail(ctx, input.Email, rawToken); err != nil {
		log.Errorw("failed to send verification email", "email", input.Email, "error", err)
	}

	log.Infow("user registered successfully", "email", input.Email)
	return nil
}

// VerifyEmail validates the token and marks the user's email as verified.
func (s *Service) VerifyEmail(ctx context.Context, rawToken string) error {
	ctx, span := tracer.Start(ctx, "Service.VerifyEmail")
	defer span.End()

	log := logger.FromContext(ctx, s.log)
	tokenHash := domain.HashVerificationToken(rawToken)

	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		token, err := s.tokenRepo.GetByHash(txCtx, tokenHash, domain.TokenPurposeEmailVerification)
		if err != nil {
			if errors.Is(err, tokenrepo.ErrTokenNotFound) {
				return ErrInvalidVerificationToken
			}
			return fmt.Errorf("fetch token: %w", err)
		}

		if err := s.userService.MarkEmailVerified(txCtx, token.UserID); err != nil {
			return fmt.Errorf("update verification status: %w", err)
		}

		if err := s.tokenRepo.Delete(txCtx, tokenHash); err != nil {
			return fmt.Errorf("consume token: %w", err)
		}

		return nil
	})

	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("delete token: %w", err)
	}

	log.Info("email verified successfully")
	return nil
}

// Login authenticates a user by email and password.
func (s *Service) Login(ctx context.Context, input LoginInput) (string, string, error) {
	ctx, span := tracer.Start(ctx, "Service.Login")
	defer span.End()

	log := logger.FromContext(ctx, s.log)

	u, err := s.userService.GetByEmail(ctx, input.Email)
	if err != nil {
		log.Warnw("failed login attempt: user not found", "email", input.Email)
		return "", "", ErrInvalidCredentials
	}

	if !u.CheckPassword(input.Password) {
		log.Warnw("failed login attempt: incorrect password", "user_id", u.ID)
		return "", "", ErrInvalidCredentials
	}

	if !u.IsEmailVerified {
		log.Warnw("failed login attempt: email not verified", "user_id", u.ID)
		return "", "", ErrEmailNotVerified
	}

	at, rt, err := s.issueTokens(ctx, u)
	if err != nil {
		span.RecordError(err)
		return "", "", err
	}

	log.Infow("user logged in", "user_id", u.ID)
	return at, rt, nil
}

// RequestPasswordReset generates a reset token and sends an email.
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	ctx, span := tracer.Start(ctx, "Service.RequestPasswordReset")
	defer span.End()

	log := logger.FromContext(ctx, s.log)
	u, err := s.userService.GetByEmail(ctx, email)
	if err != nil {
		return nil // Silent return to prevent enumeration
	}

	var rawToken string
	err = s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		_ = s.tokenRepo.DeleteAllForUser(txCtx, u.ID, domain.TokenPurposePasswordReset)

		tokenStr, entity, err := domain.NewVerificationToken(u.ID, domain.TokenPurposePasswordReset, s.authConfig.PasswordResetTTL)
		if err != nil {
			return fmt.Errorf("generate token: %w", err)
		}

		rawToken = tokenStr
		return s.tokenRepo.Create(txCtx, entity)
	})

	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("create token: %w", err)
	}

	if err := s.emailSender.SendPasswordResetEmail(ctx, u.Email, rawToken); err != nil {
		log.Errorw("failed to send reset email", "user_id", u.ID, "error", err)
	}

	log.Infow("password reset requested", "user_id", u.ID)
	return nil
}

// ResetPassword validates the reset token and updates the user's password.
func (s *Service) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	ctx, span := tracer.Start(ctx, "Service.ResetPassword")
	defer span.End()

	log := logger.FromContext(ctx, s.log)
	tokenHash := domain.HashVerificationToken(rawToken)

	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		token, err := s.tokenRepo.GetByHash(txCtx, tokenHash, domain.TokenPurposePasswordReset)
		if err != nil {
			if errors.Is(err, tokenrepo.ErrTokenNotFound) {
				return ErrInvalidVerificationToken
			}
			return fmt.Errorf("fetch token: %w", err)
		}

		if err := s.userService.SetNewPassword(txCtx, token.UserID, newPassword); err != nil {
			return fmt.Errorf("set new password: %w", err)
		}

		return s.tokenRepo.Delete(txCtx, tokenHash)
	})

	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("delete token: %w", err)
	}

	log.Info("password reset successfully")
	return nil
}

// LoginWithGoogle handles the OAuth flow.
func (s *Service) LoginWithGoogle(ctx context.Context, code string) (string, string, error) {
	ctx, span := tracer.Start(ctx, "Service.LoginWithGoogle")
	defer span.End()

	log := logger.FromContext(ctx, s.log)
	profile, err := s.oauthClient.ExchangeCode(ctx, code)
	if err != nil {
		span.RecordError(err)
		return "", "", ErrOAuthExchangeFailed
	}

	var targetUser *domain.User
	err = s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		conn, err := s.oauthRepo.GetByProviderUserID(txCtx, domain.OAuthProviderGoogle, profile.ProviderID)
		if err == nil {
			targetUser, err = s.userService.GetByID(txCtx, conn.UserID)
			return fmt.Errorf("failed to get oathauth: %w", err)
		}

		targetUser, err = s.userService.GetByEmail(txCtx, profile.Email)
		if err != nil {
			if !errors.Is(err, user.ErrUserNotFound) {
				return fmt.Errorf("fetch user: %w", err)
			}

			userID, createErr := s.userService.CreateOAuthUser(txCtx, user.CreateOAuthInput{
				Email:           profile.Email,
				DisplayName:     profile.DisplayName,
				PhotoURL:        profile.PhotoURL,
				IsEmailVerified: true,
			})
			if createErr != nil {
				return fmt.Errorf("create oauth user: %w", createErr)
			}
			targetUser, err = s.userService.GetByID(txCtx, userID)
			if err != nil {
				return fmt.Errorf("fetch user: %w", err)
			}
		}

		newConn := domain.NewOAuthConnection(targetUser.ID, domain.OAuthProviderGoogle, profile.ProviderID)
		return s.oauthRepo.Create(txCtx, newConn)
	})

	if err != nil {
		span.RecordError(err)
		return "", "", fmt.Errorf("oauth transaction failed: %w", err)
	}

	log.Infow("oauth login success", "user_id", targetUser.ID, "provider", "google")
	return s.issueTokens(ctx, targetUser)
}

// Refresh exchanges a valid refresh token for new tokens.
func (s *Service) Refresh(ctx context.Context, oldRawToken string) (string, string, error) {
	ctx, span := tracer.Start(ctx, "Service.Refresh")
	defer span.End()

	log := logger.FromContext(ctx, s.log)
	var at, rt string

	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		tokenHash := hashToken(oldRawToken)
		session, err := s.sessionRepo.GetByTokenHash(txCtx, tokenHash)
		if err != nil {
			return ErrRefreshTokenInvalid
		}

		if session.Revoked || time.Now().After(session.ExpiresAt) {
			return ErrRefreshTokenInvalid
		}

		u, err := s.userService.GetByID(txCtx, session.UserID)
		if err != nil {
			return fmt.Errorf("fetch user: %w", err)
		}

		at, err = s.generateAccessToken(u)
		if err != nil {
			return fmt.Errorf("generate access token: %w", err)
		}

		_ = s.sessionRepo.Revoke(txCtx, tokenHash)
		rt, err = s.generateRefreshToken(txCtx, u.ID)
		return fmt.Errorf("failed to get refresh token: %w", err)
	})

	if err != nil {
		span.RecordError(err)
		return "", "", fmt.Errorf("refresh transaction failed: %w", err)
	}

	log.Infow("token rotated", "token_hash_preview", oldRawToken[:5])
	return at, rt, nil
}

// ResendVerificationEmail invalidates existing verification tokens and issues a new one.
func (s *Service) ResendVerificationEmail(ctx context.Context, email string) error {
	u, err := s.userService.GetByEmail(ctx, email)
	if err != nil {
		// Security: If user is not found, silently return nil to prevent email enumeration.
		return nil //nolint:nilerr
	}

	if u.IsEmailVerified {
		// It's safe to return an error here so the frontend can redirect them to login.
		return ErrEmailAlreadyVerified
	}

	var rawToken string

	err = s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		// 1. Invalidate any previously requested, unused verification tokens
		if err := s.tokenRepo.DeleteAllForUser(txCtx, u.ID, domain.TokenPurposeEmailVerification); err != nil {
			return fmt.Errorf("cleanup old tokens: %w", err)
		}

		// 2. Generate new token
		tokenStr, entity, err := domain.NewVerificationToken(
			u.ID, domain.TokenPurposeEmailVerification, s.authConfig.EmailVerificationTTL,
		)
		if err != nil {
			return fmt.Errorf("generate token: %w", err)
		}

		// 3. Save new token
		if err := s.tokenRepo.Create(txCtx, entity); err != nil {
			return fmt.Errorf("save token: %w", err)
		}

		rawToken = tokenStr
		return nil
	})

	if err != nil {
		return fmt.Errorf("create token: %w", err)
	}

	// 4. Send the new email
	if err := s.emailSender.SendVerificationEmail(ctx, u.Email, rawToken); err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}

	return nil
}

// Internal helpers

func (s *Service) issueTokens(ctx context.Context, u *domain.User) (string, string, error) {
	ctx, span := tracer.Start(ctx, "Service.issueTokens")
	defer span.End()

	at, err := s.generateAccessToken(u)
	if err != nil {
		return "", "", err
	}

	rt, err := s.generateRefreshToken(ctx, u.ID)
	return at, rt, err
}

func (s *Service) generateAccessToken(u *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  u.ID.String(),
		"role": u.Role,
		"iss":  "roleTalk",
		"exp":  time.Now().Add(s.authConfig.AccessTokenTTL).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.authConfig.Secret))
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return signed, nil
}

func (s *Service) generateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	rawToken := base64.RawURLEncoding.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(s.authConfig.RefreshTokenTTL)

	session := domain.NewAuthSession(userID, hashToken(rawToken), expiresAt)
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return "", fmt.Errorf("create auth session: %w", err)
	}

	return rawToken, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
