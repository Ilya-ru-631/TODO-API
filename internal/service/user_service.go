package service

import (
	"context"
	"errors"
	"fmt"
	"todo_api/internal/auth"
	"todo_api/internal/models"
	"todo_api/internal/repository"
)

type UserService struct {
	repo   repository.UserRepository
	secret []byte
}

func NewUserService(repo repository.UserRepository, secret []byte) *UserService {
	return &UserService{repo: repo, secret: secret}
}

var (
	ErrPasswordOrLogin = errors.New("wrong password or login")
	ErrUsernameIsTaken = errors.New("username is already taken in service")
)

func (s *UserService) Register(ctx context.Context, username, password string) (string, error) {
	_, err := s.repo.GetByUsername(ctx, username)
	if err == nil {
		return "", ErrUsernameIsTaken
	}

	if !errors.Is(err, repository.ErrUserNotFound) {
		return "", fmt.Errorf("error for get by username: %w", err)
	}

	hashPas, err := auth.HashPassword(password)
	if err != nil {
		return "", err
	}

	user := models.User{
		Username:     username,
		PasswordHash: hashPas,
	}

	result, err := s.repo.Create(ctx, user)
	if errors.Is(err, repository.ErrUsernameTaken) {
		return "", errors.Join(err, ErrUsernameIsTaken)
	}

	if err != nil {
		return "", err
	}

	genRes, err := auth.GenerateToken(result.ID, s.secret)
	if err != nil {
		return "", err
	}

	return genRes, nil
}

func (s *UserService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.repo.GetByUsername(ctx, username)

	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrPasswordOrLogin
		}
		return "", fmt.Errorf("get user by username: %w", err)
	}

	err = auth.CheckPassword(user.PasswordHash, password)
	if err != nil {
		return "", ErrPasswordOrLogin
	}

	genRes, err := auth.GenerateToken(user.ID, s.secret)
	if err != nil {
		return "", err
	}
	return genRes, nil
}
