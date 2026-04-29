// Package middleware provides HTTP middleware components for the Backforge API.
// It includes utilities for logging, authentication, authorization, and rate limiting.
package middleware

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"go-backend/internal/transport/http/render"
)

// ErrTooManyRequests is returned when a client exceeds their allocated request quota.
var ErrTooManyRequests = errors.New("too many requests")

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter returns a middleware that performs rate limiting using the Token Bucket algorithm.
//
// It identifies clients by their authenticated UserID if present in the context,
// otherwise it falls back to the remote IP address.
//
// The limit parameter defines the maximum number of requests allowed per second,
// while b (burst) defines the maximum number of requests that can be handled simultaneously.
// cleanupInterval determines how long a client remains in memory after their last activity.
//
// To prevent memory leaks, a background goroutine is started to periodically remove
// inactive clients from the tracking map.
func RateLimiter(log *zap.SugaredLogger, limit rate.Limit, b int, cleanupInterval time.Duration) func(http.Handler) http.Handler {
	var (
		mu       sync.RWMutex
		visitors = make(map[string]*visitor)
	)

	// Start a background goroutine to clean up inactive visitors.
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for k, v := range visitors {
				if time.Since(v.lastSeen) > cleanupInterval {
					delete(visitors, k)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Determine the client key: prioritize UserID over IP address.
			key := r.RemoteAddr
			if userID, ok := UserIDFromContext(r.Context()); ok {
				key = userID.String()
			}

			// Check if we already have a limiter for this key.
			mu.RLock()
			v, exists := visitors[key]
			mu.RUnlock()

			if !exists {
				mu.Lock()
				// Double-check existence after acquiring write lock.
				v, exists = visitors[key]
				if !exists {
					v = &visitor{limiter: rate.NewLimiter(limit, b)}
					visitors[key] = v
				}
				mu.Unlock()
			}

			// Update the last seen timestamp.
			mu.Lock()
			v.lastSeen = time.Now()
			mu.Unlock()

			// Check if the request is allowed.
			if !v.limiter.Allow() {
				log.With(zap.String("client_key", key)).Warn("rate limit exceeded")

				if err := render.Fail(w, http.StatusTooManyRequests, ErrTooManyRequests); err != nil {
					log.With(zap.Error(err)).Warn("failed to send rate limit response")
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
