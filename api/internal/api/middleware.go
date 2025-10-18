package api

import (
	"context"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"sprintspark/internal/auth"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
	// UserEmailKey is the context key for user email
	UserEmailKey contextKey = "user_email"
)

// JWTAuth middleware validates JWT tokens from Authorization header
func (s *Server) JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "missing authorization header", "unauthorized")
			return
		}

		// Check for Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondError(w, http.StatusUnauthorized, "invalid authorization header format", "unauthorized")
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := auth.ValidateToken(tokenString, s.config.JWTSecret)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			respondError(w, http.StatusUnauthorized, "invalid or expired token", "unauthorized")
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

		// Continue to next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logger middleware logs HTTP requests
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		log.Printf(
			"%s %s %d %s",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			time.Since(start),
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// GetUserID extracts user ID from request context
func GetUserID(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value(UserIDKey).(int64)
	return userID, ok
}

// GetUserEmail extracts user email from request context
func GetUserEmail(r *http.Request) (string, bool) {
	email, ok := r.Context().Value(UserEmailKey).(string)
	return email, ok
}

// tokenBucket implements a simple token bucket rate limiter
type tokenBucket struct {
	tokens      float64
	capacity    float64
	refillRate  float64
	lastRefill  time.Time
	mu          sync.Mutex
}

func newTokenBucket(capacity, refillRate float64) *tokenBucket {
	return &tokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *tokenBucket) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	// Refill tokens based on elapsed time
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefill = now

	// Check if we have tokens available
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

// rateLimiter manages rate limits for different endpoints
type rateLimiter struct {
	buckets map[string]*tokenBucket
	mu      sync.RWMutex
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{
		buckets: make(map[string]*tokenBucket),
	}
}

func (rl *rateLimiter) getBucket(key string, capacity, refillRate float64) *tokenBucket {
	rl.mu.RLock()
	bucket, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if exists {
		return bucket
	}

	// Create new bucket
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	bucket, exists = rl.buckets[key]
	if exists {
		return bucket
	}

	bucket = newTokenBucket(capacity, refillRate)
	rl.buckets[key] = bucket
	return bucket
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	limiter := newRateLimiter()
	capacity := float64(requestsPerMinute)
	refillRate := capacity / 60.0 // tokens per second

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use IP address as key
			ip := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ip = strings.Split(forwarded, ",")[0]
			}
			if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				ip = realIP
			}

			bucket := limiter.getBucket(ip, capacity, refillRate)
			if !bucket.allow() {
				respondError(w, http.StatusTooManyRequests, "rate limit exceeded", "rate_limit_exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
