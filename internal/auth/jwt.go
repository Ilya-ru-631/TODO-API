package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrValidationToken      = errors.New("error for validation token")
	ErrUnexpectedType       = errors.New("claims have an unexpected type")
	ErrSubIsMissing         = errors.New("sub is missing in payload")
	ErrSubIsNotStoredString = errors.New("sub is not stored as a string")
	ErrSubNotNumber         = errors.New("sub is not a number")
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
	}, jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, ErrValidationToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, ErrUnexpectedType
	}

	subValue, ok := claims["sub"]
	if !ok {
		return 0, ErrSubIsMissing
	}

	subStr, ok := subValue.(string)
	if !ok {
		return 0, ErrSubIsNotStoredString
	}

	userID, err := strconv.Atoi(subStr)
	if err != nil {
		return 0, ErrSubNotNumber
	}
	return userID, nil
}
