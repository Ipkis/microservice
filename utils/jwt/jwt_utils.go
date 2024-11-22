package jwt_utils

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserInfo struct {
	UserID  int64
	IsAdmin bool
}

func NewUserInfo() UserInfo {
	return UserInfo{
		UserID:  -1,
		IsAdmin: false,
	}
}

type Claims struct {
	UserInfo UserInfo
	jwt.RegisteredClaims
}

var jwt_secret string

func InitJwtSecret(secret string) {
	jwt_secret = secret
}

// TokenRevocationChecker - интерфейс для проверки отозванности токенов
type TokenRevocationChecker interface {
	IsTokenRevoked(token string) (bool, error)
}

func hashUsername(username string) string {
	hash := sha256.Sum256([]byte(username)) // Вычисляем SHA-256
	return hex.EncodeToString(hash[:])      // Конвертируем в строку
}

func GenerateJWT(username string) (string, error) {
	claims := &Claims{
		UserInfo: NewUserInfo(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "auth_service",                                      // Кто выдал токен
			Subject:   hashUsername(username),                              // Кто является владельцем
			IssuedAt:  jwt.NewNumericDate(time.Now()),                      // Время выдачи токена
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)), // Время истечения срока действия
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwt_secret))
}

func GetJwtTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is missing")
	}

	// "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}

func GetJWTClaims(token_str string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		token_str,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwt_secret), nil
		})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("Invalid token")
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("Unable to extract claims from token")
}

func JwtMiddleware(next http.Handler, revocationChecker TokenRevocationChecker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenSigned, err := GetJwtTokenFromHeader(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		isRevoked, err := revocationChecker.IsTokenRevoked(tokenSigned)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to check token revocation: %v", err), http.StatusInternalServerError)
			return
		}
		if isRevoked {
			http.Error(w, "token is revoked", http.StatusUnauthorized)
			return
		}

		claims, err := GetJWTClaims(tokenSigned)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "claims", claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetClaimsFromContext(req *http.Request) (*Claims, error) {
	claims, ok := req.Context().Value("claims").(*Claims)
	if !ok || claims == nil {
		return nil, fmt.Errorf("Unable to extract claims from contest")
	}

	return claims, nil
}
