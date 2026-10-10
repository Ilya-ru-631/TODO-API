package auth

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func Test_ParsingToken(t *testing.T) {
	goodSecret := []byte("secret")
	otherSecret := []byte("other-secret")

	tests := []struct {
		name         string
		claims       jwt.MapClaims
		verifySecret []byte
		wantErr      error
		wantUserID   int
	}{
		{
			name:         "no_error",
			claims:       jwt.MapClaims{"sub": "5", "exp": time.Now().Add(time.Hour).Unix()},
			verifySecret: goodSecret,
			wantErr:      nil,
			wantUserID:   5,
		},
		{
			name:         "someone_secret",
			claims:       jwt.MapClaims{"sub": "5", "exp": time.Now().Add(time.Hour).Unix()},
			verifySecret: otherSecret,
			wantErr:      jwt.ErrSignatureInvalid,
			wantUserID:   0,
		},
		{
			name:         "no_claims",
			claims:       jwt.MapClaims{},
			verifySecret: goodSecret,
			wantErr:      jwt.ErrTokenRequiredClaimMissing,
			wantUserID:   0,
		},
		{
			name:         "exp_in_the_past",
			claims:       jwt.MapClaims{"sub": "5", "exp": time.Now().Add(-time.Hour).Unix()},
			verifySecret: goodSecret,
			wantErr:      jwt.ErrTokenExpired,
			wantUserID:   0,
		},
		{
			name:         "no_sub",
			claims:       jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()},
			verifySecret: goodSecret,
			wantErr:      ErrSubIsMissing,
			wantUserID:   0,
		},
		{
			name:         "sub_integer",
			claims:       jwt.MapClaims{"sub": 5, "exp": time.Now().Add(time.Hour).Unix()},
			verifySecret: goodSecret,
			wantErr:      ErrSubIsNotStoredString,
			wantUserID:   0,
		},
		{
			name:         "sub_abc",
			claims:       jwt.MapClaims{"sub": "abc", "exp": time.Now().Add(time.Hour).Unix()},
			verifySecret: goodSecret,
			wantErr:      ErrSubNotNumber,
			wantUserID:   0,
		},
		{
			name:         "empty_sub",
			claims:       jwt.MapClaims{"sub": "", "exp": time.Now().Add(time.Hour).Unix()},
			verifySecret: goodSecret,
			wantErr:      ErrSubNotNumber,
			wantUserID:   0,
		},
		{
			name:         "no_exp",
			claims:       jwt.MapClaims{"sub": "5"},
			verifySecret: goodSecret,
			wantErr:      jwt.ErrTokenRequiredClaimMissing,
			wantUserID:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := tt.claims

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString(goodSecret)
			if err != nil {
				t.Fatalf("error when converting a token from JWT to string: %v", err)
			}

			userID, err := ParsingToken(tokenString, tt.verifySecret)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error: %v, want error: %v", err, tt.wantErr)
			}

			if userID != tt.wantUserID {
				t.Errorf("got user id: %v, want user id: %v", userID, tt.wantUserID)
			}
		})
	}
}

func Test_Circle_GenParseToken(t *testing.T) {
	userID := 5
	secret := []byte("secret")

	genRes, err := GenerateToken(userID, secret)
	if err != nil {
		t.Fatalf("error for generate token: %v", err)
	}

	parseUserID, err := ParsingToken(genRes, secret)
	if err != nil {
		t.Fatalf("error for parsing token: %v", err)
	}

	if parseUserID != userID {
		t.Errorf("got user id: %v, want user id: %v", parseUserID, userID)
	}
}

func Test_InvalidParamParse(t *testing.T) {
	secret := []byte("secret")
	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{
			name:    "empty_token",
			token:   "",
			wantErr: jwt.ErrTokenMalformed,
		},
		{
			name:    "invalid_token",
			token:   "abc",
			wantErr: jwt.ErrTokenMalformed,
		},
		{
			name:    "invalid_token_right_form",
			token:   "a.b.c",
			wantErr: jwt.ErrTokenMalformed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsingToken(tt.token, secret)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error: %v, want error: %v", err, tt.wantErr)
			}
		})
	}

}

func Test_ParsingToken_WrongAlgorithm(t *testing.T) {
	secret := []byte("secret")
	userID := 5
	claims := jwt.MapClaims{
		"sub": strconv.Itoa(userID),
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token512 := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token512.SignedString(secret)
	if err != nil {
		t.Fatalf("error for signing token")
	}

	parseUserID, err := ParsingToken(tokenString, secret)
	if !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		t.Errorf("got error: %v, want error: %v", err, jwt.ErrTokenSignatureInvalid)
	}

	if parseUserID != 0 {
		t.Errorf("want user id: 0")
	}
}
