package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 是访问令牌声明。
type Claims struct {
	UID      int64  `json:"uid"`
	OrgID    int64  `json:"org_id"`
	IsSuper  bool   `json:"is_super"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Manager 管理JWT签发与校验。
type Manager struct {
	secret []byte
	expire time.Duration
}

// NewManager 创建令牌管理器。
func NewManager(secret string, expire time.Duration) *Manager {
	return &Manager{secret: []byte(secret), expire: expire}
}

// ExpiresIn 返回令牌有效秒数。
func (m *Manager) ExpiresIn() int64 {
	return int64(m.expire.Seconds())
}

// Issue 签发访问令牌。
func (m *Manager) Issue(uid int64, orgID int64, username string, isSuper bool) (string, error) {
	now := time.Now()
	claims := Claims{
		UID:      uid,
		OrgID:    orgID,
		Username: username,
		IsSuper:  isSuper,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expire)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse 校验并解析访问令牌。
func (m *Manager) Parse(raw string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(raw, &Claims{}, func(parsed *jwt.Token) (any, error) {
		if parsed.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}
