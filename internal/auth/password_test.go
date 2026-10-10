package auth

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func Test_HashPassword(t *testing.T) {
	testPassword := "secret"
	result, err := HashPassword(testPassword)
	if err != nil {
		t.Fatalf("error for generate password: %v", err)
	}

	if result == testPassword {
		t.Errorf("error for hashing password")
	}

	if err := CheckPassword(result, testPassword); err != nil {
		t.Errorf("hash do not matched with password: %v", err)
	}
}

func Test_CheckPassword(t *testing.T) {
	goodPassword := "secret"
	otherPassword := "other"

	hash, hashErr := HashPassword(goodPassword)
	if hashErr != nil {
		t.Fatalf("error for hashing password: %v", hashErr)
	}

	tests := []struct {
		name         string
		wantErr      error
		hashForChech string
		password     string
	}{
		{
			name:         "check_completed",
			wantErr:      nil,
			hashForChech: hash,
			password:     goodPassword,
		},
		{
			name:         "wrong_password",
			wantErr:      bcrypt.ErrMismatchedHashAndPassword,
			hashForChech: hash,
			password:     otherPassword,
		},
		{
			name:         "wrong_hash",
			wantErr:      bcrypt.ErrHashTooShort,
			hashForChech: "123",
			password:     goodPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPassword(tt.hashForChech, tt.password)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error: %v, want error: %v", err, tt.wantErr)
			}

		})
	}
}

func Test_DoubleHash(t *testing.T) {
	testPassword := "secret"
	res1, err := HashPassword(testPassword)
	if err != nil {
		t.Fatalf("error for generate password: %v", err)
	}

	res2, err := HashPassword(testPassword)
	if err != nil {
		t.Fatalf("error for generate password: %v", err)
	}

	if res1 == res2 {
		t.Errorf("two hashes of identical passwords must be different")
	}
}

func Test_TooLongPassword(t *testing.T) {
	testPassword := strings.Repeat("u", 73)
	_, err := HashPassword(testPassword)

	if !errors.Is(err, bcrypt.ErrPasswordTooLong) {
		t.Errorf("got error: %v, want: %v", err, bcrypt.ErrPasswordTooLong)
	}
}


