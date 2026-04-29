// Package http manages the routing and server configuration for the RoleTalk API.
package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/riandyrn/otelchi"
	httpswagger "github.com/swaggo/http-swagger"
	"go-backend/internal/config"
	handleranalytic "go-backend/internal/transport/http/handler/analytic"
	handlerauth "go-backend/internal/transport/http/handler/auth"
	handlermessage "go-backend/internal/transport/http/handler/message"
	handlerpractice "go-backend/internal/transport/http/handler/practice_session"
	handlertopic "go-backend/internal/transport/http/handler/topic"
	handleruser "go-backend/internal/transport/http/handler/user"
	mw "go-backend/internal/transport/http/middleware"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// Handlers bundles all handler instances for the RoleTalk application.
type Handlers struct {
	Auth            *handlerauth.Handler
	User            *handleruser.Handler
	Topic           *handlertopic.Handler
	Message         *handlermessage.Handler
	Analytic        *handleranalytic.Handler
	PracticeSession *handlerpractice.Handler
}

// NewRouter initializes a Chi router with production-ready middleware,
// observability endpoints, and business-logic routes.
func NewRouter(
	cfg *config.Config,
	log *zap.SugaredLogger,
	handlers Handlers,
	userSvc mw.UserRoleChecker,
	healthHandler http.Handler, // Injected from app.go (health-go library)
) http.Handler {
	r := chi.NewRouter()

	// --- GLOBAL MIDDLEWARE ---
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(mw.Logger(log))
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(90 * time.Second)) // Extended for heavy AI processing
	r.Use(mw.Observability(log))
	r.Use(otelchi.Middleware("RoleTalk-API"))

	// --- RATE LIMITING ---
	if cfg.RateLimit.Enabled {
		r.Use(mw.RateLimiter(
			log,
			rate.Limit(cfg.RateLimit.Global.Limit),
			cfg.RateLimit.Global.Burst,
			cfg.RateLimit.CleanupInterval,
		))
	}

	// --- CORS ---
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.Client.URL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-Practice-Language"},
		ExposedHeaders:   []string{"Link", "X-STT-Transcription", "X-LLM-Response"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// --- SYSTEM ENDPOINTS ---
	r.Handle("/metrics", promhttp.Handler())
	r.Handle("/health", healthHandler) // Professional health check with DB ping
	r.Get("/swagger/*", httpswagger.Handler(httpswagger.URL("/swagger/doc.json")))

	// --- API VERSION 1 ---
	r.Route("/api/v1", func(r chi.Router) {

		// PUBLIC AUTH ROUTES
		r.Route("/auth", func(r chi.Router) {
			if cfg.RateLimit.Enabled {
				r.Use(mw.RateLimiter(log, rate.Limit(cfg.RateLimit.Auth.Limit), cfg.RateLimit.Auth.Burst, cfg.RateLimit.CleanupInterval))
			}
			r.Post("/register", handlers.Auth.Register)
			r.Post("/login", handlers.Auth.Login)
			r.Post("/verify-email", handlers.Auth.VerifyEmail)
			r.Post("/forgot-password", handlers.Auth.RequestPasswordReset)
			r.Post("/reset-password", handlers.Auth.ResetPassword)
			r.Post("/google/callback", handlers.Auth.GoogleCallback)
			r.Post("/refresh", handlers.Auth.Refresh)
			r.Post("/resend-verification", handlers.Auth.ResendVerification)
		})

		// PROTECTED ROUTES (Requires Bearer Token)
		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(cfg.Auth.Secret, log))

			// USER & PROFILE & ANALYTICS
			r.Route("/users", func(r chi.Router) {
				r.Get("/me", handlers.User.GetProfile)
				r.Patch("/me", handlers.User.UpdateProfile)
				r.Get("/me/skills", handlers.Analytic.GetMySkills) // Added Analytic handler
				// Future: r.Get("/me/achievements", handlers.Analytic.GetBadges)
			})

			// TOPICS (SCENARIOS)
			r.Route("/topics", func(r chi.Router) {
				r.Get("/official", handlers.Topic.GetOfficial)   // Solo AI recommendations
				r.Get("/community", handlers.Topic.GetCommunity) // UGC / "People" hub
				r.Post("/", handlers.Topic.Create)               // Create new scenario (+)
				r.Post("/{id}/like", handlers.Topic.AddLike)     // "Genius" rating system
				r.Delete("/{id}/like", handlers.Topic.RemoveLike)
				// Future: r.Get("/trending", handlers.Topic.GetTrending)
			})

			// VOICE & AI SESSIONS
			r.Route("/sessions", func(r chi.Router) {
				r.Post("/", handlers.PracticeSession.Start)                 // Start new practice
				r.Get("/{id}", handlers.PracticeSession.GetByID)            // Get session info
				r.Post("/{id}/complete", handlers.PracticeSession.Complete) // Finish practice

				// AI Core Loop: Receives m4a, talks to S3 and Python API
				r.Post("/{id}/voice", handlers.Message.ProcessVoiceTurn)
				r.Get("/{id}/history", handlers.Message.GetHistory)
			})

			// SOCIAL HUB (FUTURE)
			/*
				r.Route("/social", func(r chi.Router) {
					r.Get("/friends", handlers.Friends.List)
					r.Post("/friends/invite", handlers.Friends.Invite)
					r.Get("/leaderboard", handlers.Analytic.GetGlobalRankings)
				})
			*/

			// ADMIN SECTION
			r.Route("/admin", func(r chi.Router) {
				r.Use(mw.AdminOnly(log, userSvc))
				r.Delete("/topics/{id}", handlers.Topic.Delete)
				// Future: r.Get("/stats/usage", handlers.Admin.GetDetailedMetrics)
			})
		})
	})

	return r
}
