// Package app handles the initialization, dependency injection, and lifecycle
// management of the RoleTalk backend application.
package app

import (
	"context"
	"fmt"
	v1 "go-backend/internal/transport/http/v1"
	handleranalytic "go-backend/internal/transport/http/v1/handler/analytic"
	handlerauth "go-backend/internal/transport/http/v1/handler/auth"
	handlermessage "go-backend/internal/transport/http/v1/handler/message"
	handlerpractice "go-backend/internal/transport/http/v1/handler/practice_session"
	handlertopic "go-backend/internal/transport/http/v1/handler/topic"
	handleruser "go-backend/internal/transport/http/v1/handler/user"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-backend/internal/infra/rabbitmq"
	"go-backend/internal/infra/redis"

	"go.uber.org/zap"

	"go-backend/internal/config"
	"go-backend/internal/infra/ai"
	"go-backend/internal/infra/health"
	"go-backend/internal/infra/mailer"
	"go-backend/internal/infra/minio"
	"go-backend/internal/infra/oauth"
	"go-backend/internal/infra/ollama"
	"go-backend/internal/infra/postgres"
	"go-backend/internal/infra/prompt"
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

// Services aggregates all business logic layer instances.
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
func New(ctx context.Context) (*App, error) {
	// 1. Core: Configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("app.New: load config: %w", err)
	}

	// 2. Infrastructure: Telemetry
	// We init telemetry first to get the OTel log core for the actual logger
	tel, otelCore, err := telemetry.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("app.New: init telemetry: %w", err)
	}

	// 3. Core: Structured Logging
	log, err := logger.New(logger.Config{
		Environment: cfg.Env,
		Level:       cfg.Logging.Level,
		OTelCore:    otelCore,
	})
	if err != nil {
		return nil, fmt.Errorf("app.New: init logger: %w", err)
	}

	log.Infow("starting RoleTalk API", "version", cfg.App.Version, "env", cfg.Env)

	// 4. Infrastructure: Database
	log.Infof("connecting to postgres at %s...", cfg.Postgres.Host)
	pgPool, err := postgres.New(ctx, cfg.Postgres.ConnectionURL, &cfg.Postgres.Pool)
	if err != nil {
		return nil, fmt.Errorf("app.New: postgres: %w", err)
	}
	log.Info("postgres connection established")

	// 5. Infrastructure: Storage
	log.Infof("initializing minio storage (bucket: %s)...", cfg.MinIO.Bucket)
	fileStore, err := minio.FromConfig(ctx, &cfg.MinIO, log)
	if err != nil {
		return nil, fmt.Errorf("app.New: minio: %w", err)
	}
	log.Info("minio storage ready")

	// 6. Infrastructure: Ai
	log.Infof("initializing AI client at %s...", cfg.AI.Addr)
	aiClient, err := ai.NewClient(cfg.AI.Addr, log)
	if err != nil {
		return nil, fmt.Errorf("app.New: init ai: %w", err)
	}
	log.Info("initializing AI client")

	// 7. Infrastructure: Redis
	log.Infof("initializing redis client at %s...", cfg.Redis.Addr)
	_, err = redis.New(ctx, &cfg.Redis, log)
	if err != nil {
		return nil, fmt.Errorf("app.New: redis: %w", err)
	}
	log.Info("redis client ready")

	// 8. Infrastructure: RabbitMQ
	log.Infof("initializing rabbitmq client at %s...", cfg.RabbitMQ.Port)
	_, err = rabbitmq.NewClient(&cfg.RabbitMQ, log)
	if err != nil {
		return nil, fmt.Errorf("app.New: rabbitmq: %w", err)
	}
	log.Info("rabbitmq client ready")

	// 9. Global Helpers & Engines
	log.Info("setting up engines and external clients...")
	tx := transactor.NewTransactor(postgres.NewPoolAdapter(pgPool), log)

	genericMailer := pkgmailer.New(pkgmailer.Config{
		Host: cfg.SMTP.Host, Port: cfg.SMTP.Port,
		Username: cfg.SMTP.Username, Password: cfg.SMTP.Password,
		FromAddress: cfg.SMTP.FromAddress,
	})
	authSender, err := mailer.NewAuthSender(genericMailer, cfg.Client.URL, cfg.API.URL)
	if err != nil {
		return nil, fmt.Errorf("app.New: auth_sender: %w", err)
	}

	googleClient := pkggoogle.New(pkggoogle.Config{
		ClientID: cfg.OAuth.Google.ClientID, ClientSecret: cfg.OAuth.Google.ClientSecret,
	}, log)
	oauthAdapter := oauth.NewGoogleAdapter(googleClient)

	promptEngine, err := prompt.NewEngine()
	if err != nil {
		return nil, fmt.Errorf("app.New: prompt_engine: %w", err)
	}

	ollamaClient := ollama.NewClient(cfg.Ollama.URL, cfg.Ollama.Model, log)
	log.Info("all engines (Prompt, Ollama, OAuth) initialized")

	// 10. Repositories
	log.Debug("wiring repositories...")
	repos := &Repositories{
		User:              repouser.NewRepository(pgPool, log),
		Message:           repomessage.NewRepository(pgPool, log),
		Topic:             repotopic.NewRepository(pgPool, log),
		Analytic:          repoanalytic.NewRepository(pgPool, log),
		AuthSession:       repoauthsession.NewRepository(pgPool, log),
		PracticeSession:   repopractice.NewRepository(pgPool, log),
		OAuthConnection:   repooauth.NewRepository(pgPool, log),
		VerificationToken: repotoken.NewRepository(pgPool, log),
	}

	// 11. Services
	log.Debug("wiring services...")
	messageSvc := svcmessage.NewService(repos.Message, fileStore, log)
	analyticSvc := svcanalytic.NewService(repos.Analytic, tx, messageSvc, repos.PracticeSession, repos.Topic, ollamaClient, promptEngine, log)
	topicSvc := svctopic.NewService(repos.Topic, tx, log)
	practiceSvc := svcpractice.NewService(repos.PracticeSession, repos.Topic, analyticSvc, log)
	userSvc := svcuser.NewService(repos.User, tx, fileStore, log)
	aiSvc := svcai.NewService(fileStore, messageSvc, repos.Topic, repos.PracticeSession, promptEngine, aiClient, log)
	authSvc := svcauth.NewService(userSvc, repos.AuthSession, repos.OAuthConnection, repos.VerificationToken, authSender, oauthAdapter, tx, &cfg.Auth, log)

	svcs := &Services{
		User:            userSvc,
		Auth:            authSvc,
		Ai:              aiSvc,
		Analytic:        analyticSvc,
		Message:         messageSvc,
		Topic:           topicSvc,
		PracticeSession: practiceSvc,
	}

	// 12. Transport Layer (Handlers, Router, Health)
	log.Info("assembling transport layer and health monitors...")
	healthHandler, err := health.NewHandler(ctx, cfg, log, pgPool, fileStore)
	if err != nil {
		return nil, fmt.Errorf("app.New: init health: %w", err)
	}

	handlers := v1.Handlers{
		Auth:            handlerauth.NewHandler(svcs.Auth, log),
		User:            handleruser.NewHandler(svcs.User, log),
		Topic:           handlertopic.NewHandler(svcs.Topic, log),
		Message:         handlermessage.NewHandler(svcs.Ai, log),
		Analytic:        handleranalytic.NewHandler(svcs.Analytic, log),
		PracticeSession: handlerpractice.NewHandler(svcs.PracticeSession, log),
	}

	router := v1.NewRouter(cfg, log, handlers, userSvc, healthHandler)
	server := transporthttp.NewServer(cfg, log, router)

	log.Info("app dependencies successfully wired")

	return &App{
		Config:       cfg,
		Logger:       log,
		Repositories: repos,
		Services:     svcs,
		Server:       server,
		Telemetry:    tel,
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
		graceCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		// 1. Flush telemetry data to Tempo
		if a.Telemetry != nil {
			if err := a.Telemetry.Shutdown(graceCtx); err != nil {
				a.Logger.Errorf("app.Run: telemetry shutdown failure: %v", err)
			}
		}

		// 2. Stop the HTTP server
		if err := a.Server.Shutdown(graceCtx); err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}
		return nil
	}
}
