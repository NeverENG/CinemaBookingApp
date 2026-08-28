package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 业务声明：用户/管理员 ID + 角色。
type Claims struct {
	UserID   int64  `json:"uid"`
	Role     string `json:"role"`
	CinemaID *int64 `json:"cid,omitempty"`
	jwt.RegisteredClaims
}

// Manager 签发与校验 JWT。
type Manager struct {
	secret []byte
	ttl    time.Duration
}

func New(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl}
}

func (m *Manager) Generate(userID int64, role string, cinemaID *int64) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Role:     role,
		CinemaID: cinemaID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse 校验签名与有效期，返回业务声明。
func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
