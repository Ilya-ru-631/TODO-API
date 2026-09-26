package service

import (
	"context"
	"errors"
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
	ErrUsernameTaken   = errors.New("username is already taken")
	ErrPasswordOrLogin = errors.New("wrong password or login")
)

func (s *UserService) Register(ctx context.Context, username, password string) (string, error) {
	hashPas, err := auth.HashPassword(password)
	if err != nil {
		return "", err
	}

	user := models.User{
		Username:     username,
		PasswordHash: hashPas,
	}

	_, err = s.repo.GetByUsername(ctx, username)
	if err == nil {
		return "", ErrUsernameTaken
	}

	if !errors.Is(err, repository.ErrUserNotFound) {
		return "", err
	}

	result, err := s.repo.Create(ctx, user)
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
		return "", ErrPasswordOrLogin
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
