package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestTokenBucketAllowsBurstAndRefills(t *testing.T) {
	now := time.Unix(100, 0)
	limiter := NewTokenBucketLimiter(2, 2, time.Hour)

	if allowed, _ := limiter.AllowAt("client", now); !allowed {
		t.Fatal("first request should be allowed")
	}
	if allowed, _ := limiter.AllowAt("client", now); !allowed {
		t.Fatal("second burst request should be allowed")
	}
	if allowed, retryAfter := limiter.AllowAt("client", now); allowed || retryAfter <= 0 {
		t.Fatalf("third request should be limited, retry_after=%s", retryAfter)
	}
	if allowed, _ := limiter.AllowAt("client", now.Add(500*time.Millisecond)); !allowed {
		t.Fatal("one token should refill after 500ms")
	}
}

func TestRateLimitMiddlewareReturns429(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimit(NewTokenBucketLimiter(1, 1, time.Hour), func(*gin.Context) string {
		return "client"
	}))
	r.GET("/resource", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	first := httptest.NewRecorder()
	r.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/resource", nil))
	if first.Code != http.StatusNoContent {
		t.Fatalf("first request: expected 204, got %d", first.Code)
	}

	second := httptest.NewRecorder()
	r.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/resource", nil))
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second request: expected 429, got %d", second.Code)
	}
	if second.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
}
