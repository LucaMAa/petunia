package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	AccessTokenType  TokenType = "access"
	RefreshTokenType TokenType = "refresh"
)

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

func getEnvDuration(name string, fallback time.Duration) time.Duration {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		if seconds, err := strconv.Atoi(value); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return fallback
}

func jwtSecrets() ([][]byte, error) {
	configured := []string{}

	if secret := strings.TrimSpace(os.Getenv("JWT_SECRET")); secret != "" {
		configured = append(configured, secret)
	}

	if rotated := strings.TrimSpace(os.Getenv("JWT_ROTATION_SECRETS")); rotated != "" {
		for _, secret := range strings.Split(rotated, ",") {
			secret = strings.TrimSpace(secret)
			if secret != "" {
				configured = append(configured, secret)
			}
		}
	}

	if len(configured) == 0 {
		return nil, errors.New("JWT_SECRET is required in production; set it to a strong random value")
	}

	secrets := make([][]byte, 0, len(configured))
	for _, secret := range configured {
		if len(secret) < 32 {
			return nil, fmt.Errorf("JWT secret is too short: %q; use at least 32 chars", secret)
		}
		secrets = append(secrets, []byte(secret))
	}

	return secrets, nil
}

func jwtIssuer() string {
	if issuer := strings.TrimSpace(os.Getenv("JWT_ISSUER")); issuer != "" {
		return issuer
	}
	return "petunia"
}

func jwtAudience() []string {
	if audience := strings.TrimSpace(os.Getenv("JWT_AUDIENCE")); audience != "" {
		return strings.Split(audience, ",")
	}
	return nil
}

func jwtAccessTTL() time.Duration {
	return getEnvDuration("JWT_ACCESS_TTL_SECONDS", 15*time.Minute)
}

func generateToken(userID uuid.UUID, tokenType TokenType, ttl time.Duration) (string, error) {
	secrets, err := jwtSecrets()
	if err != nil {
		return "", err
	}

	claims := Claims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    jwtIssuer(),
			Audience:  jwt.ClaimStrings(jwtAudience()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secrets[0])
}

func GenerateAccessToken(userID uuid.UUID) (string, error) {
	return generateToken(userID, AccessTokenType, jwtAccessTTL())
}

func GenerateRefreshToken(userID uuid.UUID) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateToken(userID uuid.UUID) (string, error) {
	return GenerateAccessToken(userID)
}

func parseTokenWithSecret(tokenStr string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signature method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("token not valid")
	}

	return claims, nil
}

func ParseToken(tokenStr string) (*Claims, error) {
	return ParseTokenOfType(tokenStr, "")
}

func ParseTokenOfType(tokenStr string, expectedType TokenType) (*Claims, error) {
	secrets, err := jwtSecrets()
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, secret := range secrets {
		claims, err := parseTokenWithSecret(tokenStr, secret)
		if err != nil {
			lastErr = err
			continue
		}

		if claims.TokenType == "" {
			claims.TokenType = AccessTokenType
		}

		if expectedType != "" && claims.TokenType != expectedType {
			return nil, errors.New("invalid token type")
		}

		if expectedIssuer := jwtIssuer(); expectedIssuer != "" && claims.Issuer != expectedIssuer {
			return nil, errors.New("invalid issuer")
		}

		if expectedAudiences := jwtAudience(); len(expectedAudiences) > 0 {
			matched := false
			for _, expectedAudience := range expectedAudiences {
				for _, claimAudience := range claims.Audience {
					if claimAudience == expectedAudience {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
			if !matched {
				return nil, errors.New("invalid audience")
			}
		}

		return claims, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("token not valid")
}
