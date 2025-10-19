package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// CORSMiddleware enforces the configured allowlist of origins.
func CORSMiddleware(allowed []string) gin.HandlerFunc {
	allowAll := len(allowed) == 0
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, origin := range allowed {
		if origin == "" {
			continue
		}
		allowedSet[origin] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		origin = strings.TrimSpace(origin)
		_, ok := allowedSet[origin]
		if !allowAll && !ok {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		allowedOrigin := origin
		if allowAll {
			allowedOrigin = "*"
		}

		c.Header("Vary", "Origin")
		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		if allowedOrigin != "*" {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// IPRateLimiter implements a simple per-IP token bucket limiter.
type IPRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
}

// NewIPRateLimiter constructs a limiter with the supplied rate and burst.
func NewIPRateLimiter(rps float64, burst int) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(rps),
		burst:    burst,
	}
}

func (l *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	if ip == "" {
		ip = "unknown"
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(l.rate, l.burst)
		l.limiters[ip] = limiter
	}

	return limiter
}

// Middleware returns the Gin middleware that enforces the rate limit.
func (l *IPRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !l.getLimiter(ip).Allow() {
			reset := time.Now().Add(time.Second)
			c.Header("Retry-After", reset.Format(time.RFC1123))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}
