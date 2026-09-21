package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userId int, secret []byte) (string, error) {
	claims := jwt.MapClaims{
		"sub": strconv.Itoa(userId),
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ParsingToken(tokenString string, secret []byte) (int, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return 0, err
	}

	if !token.Valid{
		return 0, fmt.Errorf("error for validation token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("claims have an unexpected type")
	}

	subValue, ok := claims["sub"]
	if !ok {
		return 0, fmt.Errorf("sub is missing in payload")
	}

	subStr, ok := subValue.(string)
	if !ok {
		return 0, fmt.Errorf("sub is not stored as a string")
	}

	userID, err := strconv.Atoi(subStr)
	if err != nil {
		return 0, fmt.Errorf("sub is not a number")
	}
	return userID, nil
}
