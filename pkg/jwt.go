package pkg

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"air-social/internal/config"
)

const (
	ACCESS_TOKEN  = "access"
	REFRESH_TOKEN = "refresh"
)

type JWTAuth interface {
	GenerateAccessToken(userID int64) (string, error)
	GenerateRefreshToken(userID int64) (string, error)
	Validate(token string) (*jwt.Token, error)
}

type JWTAuthImpl struct {
	Secret          string
	Aud             string
	Iss             string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func NewJWTAuth(cfg config.JWTConfig) *JWTAuthImpl {
	return &JWTAuthImpl{
		Secret:          cfg.Secret,
		Aud:             cfg.Aud,
		Iss:             cfg.Iss,
		AccessTokenTTL:  cfg.AccessTokenTTL,
		RefreshTokenTTL: cfg.RefreshTokenTTL,
	}
}

func (j *JWTAuthImpl) GenerateAccessToken(userID int64) (string, error) {
	return j.generateToken(j.getClaims(ACCESS_TOKEN, userID))

}

func (j *JWTAuthImpl) GenerateRefreshToken(userID int64) (string, error) {
	return j.generateToken(j.getClaims(REFRESH_TOKEN, userID))
}

func (j *JWTAuthImpl) Validate(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(j.Secret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(j.Aud),
		jwt.WithIssuer(j.Iss),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
}

func (j *JWTAuthImpl) getClaims(typ string, userID int64) jwt.MapClaims {
	now := time.Now()
	var duration time.Duration
	if typ == REFRESH_TOKEN {
		duration = j.RefreshTokenTTL
	} else {
		duration = j.AccessTokenTTL
	}

	return jwt.MapClaims{
		"sub": fmt.Sprintf("%d", userID),
		"aud": j.Aud,
		"iss": j.Iss,
		"iat": now.Unix(),
		"nbf": now.Unix(),
		"exp": now.Add(duration).Unix(),
		"typ": typ,
	}
}

func (j *JWTAuthImpl) generateToken(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.Secret))
}
