package middleware

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/gin-gonic/gin"
)

type tokenBucket struct {
	tokens   float64
	refilled time.Time
	lastSeen time.Time
}

// TokenBucketLimiter 是单实例令牌桶限流器。
type TokenBucketLimiter struct {
	mu            sync.Mutex
	ratePerSecond float64
	burst         float64
	idleTTL       time.Duration
	lastCleanup   time.Time
	buckets       map[string]*tokenBucket
}

// NewTokenBucketLimiter 创建按 key 隔离的令牌桶限流器。
func NewTokenBucketLimiter(ratePerSecond float64, burst int, idleTTL time.Duration) *TokenBucketLimiter {
	if ratePerSecond <= 0 {
		panic("ratePerSecond must be positive")
	}
	if burst <= 0 {
		panic("burst must be positive")
	}
	if idleTTL <= 0 {
		panic("idleTTL must be positive")
	}
	now := time.Now()
	return &TokenBucketLimiter{
		ratePerSecond: ratePerSecond,
		burst:         float64(burst),
		idleTTL:       idleTTL,
		lastCleanup:   now,
		buckets:       make(map[string]*tokenBucket),
	}
}

// Allow 判断 key 当前是否还有可用令牌。
func (l *TokenBucketLimiter) Allow(key string) (bool, time.Duration) {
	return l.AllowAt(key, time.Now())
}

// AllowAt 使用指定时间判断 key，便于测试令牌补充和过期清理。
func (l *TokenBucketLimiter) AllowAt(key string, now time.Time) (bool, time.Duration) {
	key = strings.TrimSpace(key)
	if key == "" {
		key = "anonymous"
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.cleanup(now)
	bucket, ok := l.buckets[key]
	if !ok {
		bucket = &tokenBucket{
			tokens:   l.burst,
			refilled: now,
			lastSeen: now,
		}
		l.buckets[key] = bucket
	}

	if now.After(bucket.refilled) {
		elapsed := now.Sub(bucket.refilled).Seconds()
		bucket.tokens = math.Min(l.burst, bucket.tokens+elapsed*l.ratePerSecond)
		bucket.refilled = now
	}
	bucket.lastSeen = now

	if bucket.tokens < 1 {
		wait := (1 - bucket.tokens) / l.ratePerSecond
		return false, time.Duration(math.Ceil(wait * float64(time.Second)))
	}
	bucket.tokens--
	return true, 0
}

func (l *TokenBucketLimiter) cleanup(now time.Time) {
	if now.Sub(l.lastCleanup) < time.Minute {
		return
	}
	for key, bucket := range l.buckets {
		if now.Sub(bucket.lastSeen) >= l.idleTTL {
			delete(l.buckets, key)
		}
	}
	l.lastCleanup = now
}

// ClientIPKey 使用 Gin 解析出的客户端 IP 作为限流 key。
func ClientIPKey(c *gin.Context) string {
	return c.ClientIP()
}

// RateLimit 将令牌桶应用为 Gin 中间件。
func RateLimit(limiter *TokenBucketLimiter, keyFn func(*gin.Context) string) gin.HandlerFunc {
	if limiter == nil {
		panic("limiter must not be nil")
	}
	if keyFn == nil {
		keyFn = ClientIPKey
	}
	return func(c *gin.Context) {
		allowed, retryAfter := limiter.Allow(keyFn(c))
		if !allowed {
			seconds := int(math.Ceil(retryAfter.Seconds()))
			if seconds < 1 {
				seconds = 1
			}
			c.Header("Retry-After", strconv.Itoa(seconds))
			resp.Fail(c, http.StatusTooManyRequests, "too many requests")
			c.Abort()
			return
		}
		c.Next()
	}
}
