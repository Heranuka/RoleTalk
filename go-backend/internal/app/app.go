// Package app provides the application wiring and startup lifecycle logic.
// It assembles configuration, infrastructure, services, and transport layers into a single unit.
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"go-backend/internal/config"
	"go-backend/internal/infra/health"
	"go-backend/internal/infra/mailer"
	"go-backend/internal/infra/minio"
	"go-backend/internal/infra/oauth"
	"go-backend/internal/infra/postgres"
	"go-backend/internal/infra/telemetry"
	"go-backend/internal/logger"

	// Repositories
	repoanalytic "go-backend/internal/repository/analytic"
	repoauthsession "go-backend/internal/repository/auth_session"
	repomessage "go-backend/internal/repository/message"
	repooauth "go-backend/internal/repository/oauthconnection"
	repopractice "go-backend/internal/repository/practice_session"
	repotopic "go-backend/internal/repository/topic"
	repouser "go-backend/internal/repository/user"
	repotoken "go-backend/internal/repository/verificationtoken"

	// Services
	svcai "go-backend/internal/service/ai"
	svcanalytic "go-backend/internal/service/analytic"
	svcauth "go-backend/internal/service/auth"
	svcmessage "go-backend/internal/service/message"
	svcpractice "go-backend/internal/service/practice_session"
	svctopic "go-backend/internal/service/topic"
	svcuser "go-backend/internal/service/user"

	// Transport
	transporthttp "go-backend/internal/transport/http"
	handleranalytic "go-backend/internal/transport/http/handler/analytic"
	handlerauth "go-backend/internal/transport/http/handler/auth"
	handlermessage "go-backend/internal/transport/http/handler/message"
	handlerpractice "go-backend/internal/transport/http/handler/practice_session"
	handlertopic "go-backend/internal/transport/http/handler/topic"
	handleruser "go-backend/internal/transport/http/handler/user"

	pkggoogle "go-backend/pkg/google"
	pkgmailer "go-backend/pkg/mailer"
	"go-backend/pkg/transactor"
)

// App holds all application components and manages the server lifecycle.
type App struct {
	Config       *config.Config
	Logger       *zap.SugaredLogger
	Repositories *Repositories
	Services     *Services
	Server       *transporthttp.Server
	Telemetry    *telemetry.Telemetry
}

// Repositories aggregates all database access layer instances.
type Repositories struct {
	User              *repouser.Repository
	Message           *repomessage.Repository
	Topic             *repotopic.Repository
	Analytic          *repoanalytic.Repository
	AuthSession       *repoauthsession.Repository
	PracticeSession   *repopractice.Repository
	OAuthConnection   *repooauth.Repository
	VerificationToken *repotoken.Repository
}

// Services aggregates all core business logic layer instances.
type Services struct {
	User            *svcuser.Service
	Auth            *svcauth.Service
	Ai              *svcai.Service
	Analytic        *svcanalytic.Service
	Message         *svcmessage.Service
	Topic           *svctopic.Service
	PracticeSession *svcpractice.Service
}

// New creates and wires all application dependencies using a clean injection pattern.
// It initializes infrastructure, repositories, services, and handlers.
func New(ctx context.Context) (*App, error) {
	// 1. Core: Configuration and Structured Logging
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("app.New: load config: %w", err)
	}

	log, err := logger.New(logger.Config{
		Environment: cfg.Env,
		Level:       cfg.Logging.Level,
		OutputPaths: []string{"stdout"},
	})
	if err != nil {
		return nil, fmt.Errorf("app.New: init logger: %w", err)
	}

	log.Infow("starting RoleTalk API", "version", cfg.App.Version, "env", cfg.Env)

	// 2. Infrastructure: Telemetry, Persistence, and Storage
	log.Info("initializing telemetry (Tempo/Loki)...")
	telemetryClient, err := telemetry.New(ctx, cfg, log)
	if err != nil {
		return nil, fmt.Errorf("app.New: init telemetry: %w", err)
	}

	log.Infof("connecting to postgres at %s...", cfg.Postgres.Host)
	pgPool, err := postgres.New(ctx, cfg.Postgres.ConnectionURL, &cfg.Postgres.Pool)
	if err != nil {
		return nil, fmt.Errorf("app.New: postgres: %w", err)
	}

	log.Infof("initializing minio storage (bucket: %s)...", cfg.MinIO.Bucket)
	fileStore, err := minio.FromConfig(ctx, &cfg.MinIO, log)
	if err != nil {
		return nil, fmt.Errorf("app.New: minio: %w", err)
	}

	// 3. Global Helpers: Transactor and External Clients
	tx := transactor.NewTransactor(postgres.NewPoolAdapter(pgPool), log)

	genericMailer := pkgmailer.New(pkgmailer.Config{
		Host: cfg.SMTP.Host, Port: cfg.SMTP.Port,
		Username: cfg.SMTP.Username, Password: cfg.SMTP.Password,
		FromAddress: cfg.SMTP.FromAddress,
	})
	authSender, err := mailer.NewAuthSender(genericMailer, cfg.Client.URL)
	if err != nil {
		return nil, fmt.Errorf("app.New: auth_sender: %w", err)
	}

	googleClient := pkggoogle.New(pkggoogle.Config{
		ClientID:     cfg.OAuth.Google.ClientID,
		ClientSecret: cfg.OAuth.Google.ClientSecret,
	}, log)
	oauthAdapter := oauth.NewGoogleAdapter(googleClient)

	// 4. Repositories: Data Access Layer
	log.Debug("wiring repositories...")
	repos := &Repositories{
		User:              repouser.NewRepository(pgPool, log),
		Topic:             repotopic.NewRepository(pgPool, log),
		Analytic:          repoanalytic.NewRepository(pgPool, log),
		Message:           repomessage.NewRepository(pgPool, log),
		PracticeSession:   repopractice.NewRepository(pgPool, log),
		OAuthConnection:   repooauth.NewRepository(pgPool, log),
		VerificationToken: repotoken.NewRepository(pgPool, log),
	}

	// 5. Services: Business Logic Layer
	// Initialization order follows the dependency hierarchy
	log.Debug("wiring services...")
	msgSvc := svcmessage.NewService(repos.Message, fileStore, log)
	analyticSvc := svcanalytic.NewService(repos.Analytic, tx, log)
	topicSvc := svctopic.NewService(repos.Topic, tx, log)
	practiceSvc := svcpractice.NewService(repos.PracticeSession, repos.Topic, log)
	userSvc := svcuser.NewService(repos.User, tx, fileStore, log)
	aiSvc := svcai.NewService(fileStore, msgSvc, cfg.API.URL, log)
	authSvc := svcauth.NewService(userSvc, repos.AuthSession, repos.OAuthConnection, repos.VerificationToken, authSender, oauthAdapter, tx, &cfg.Auth, log)

	svcs := &Services{
		User: userSvc, Auth: authSvc, Ai: aiSvc,
		Analytic: analyticSvc, Message: msgSvc,
		Topic: topicSvc, PracticeSession: practiceSvc,
	}

	// 6. Transport: Handlers, Health Checks, and Server
	log.Info("configuring health monitors...")
	healthHandler, err := health.NewHandler(ctx, cfg, log, pgPool, fileStore)
	if err != nil {
		return nil, fmt.Errorf("app.New: init health: %w", err)
	}

	log.Info("assembling http router and server...")
	handlers := transporthttp.Handlers{
		Auth:            handlerauth.NewHandler(svcs.Auth, log),
		User:            handleruser.NewHandler(svcs.User, log),
		Topic:           handlertopic.NewHandler(svcs.Topic, log),
		Message:         handlermessage.NewHandler(svcs.Ai, log), // Ai service orchestrates ProcessVoiceTurn
		Analytic:        handleranalytic.NewHandler(svcs.Analytic, log),
		PracticeSession: handlerpractice.NewHandler(svcs.PracticeSession, log),
	}

	router := transporthttp.NewRouter(cfg, log, handlers, userSvc, healthHandler)
	server := transporthttp.NewServer(cfg, log, router)

	log.Info("app components successfully wired")

	return &App{
		Config:       cfg,
		Logger:       log,
		Repositories: repos,
		Services:     svcs,
		Server:       server,
		Telemetry:    telemetryClient,
	}, nil
}

// Run starts the HTTP server and blocks until an OS interrupt signal is received.
// It performs a graceful shutdown, flushing telemetry and stopping the server.
func (a *App) Run(ctx context.Context) error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	errServer := make(chan error, 1)
	go func() {
		a.Logger.Infof("RoleTalk API listening on %s", a.Config.HTTP.Port)
		errServer <- a.Server.Run(ctx)
	}()

	select {
	case err := <-errServer:
		return fmt.Errorf("app.Run: server encountered error: %w", err)

	case sig := <-shutdown:
		a.Logger.Infof("app.Run: shutdown signal received: %v", sig)

		// Create a grace period context for cleanup operations
		graceCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// 1. Flush telemetry data to Tempo
		if a.Telemetry != nil {
			if err := a.Telemetry.Shutdown(graceCtx); err != nil {
				a.Logger.Errorf("app.Run: telemetry shutdown failure: %v", err)
			}
		}

		// 2. Stop the HTTP server
		return a.Server.Shutdown(graceCtx)
	}
}
