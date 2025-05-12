package utils

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// GenerateJWT create jwt token with user_id, is_refresh fields and TTL
func GenerateJWT(userID, jwtSecret string, isRefresh bool, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":        userID,
		"exp":        time.Now().Add(ttl).Unix(),
		"iat":        time.Now().Unix(),
		"is_refresh": isRefresh,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// ParseJWT parse jwt token, validate it and return user_id and is_refresh
func ParseJWT(tokenStr, jwtSecret string) (string, bool, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return "", false, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", false, fmt.Errorf("invalid JWT claims")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", false, err
	}

	isRefresh, _ := claims["is_refresh"].(bool)

	return userID, isRefresh, nil
}
