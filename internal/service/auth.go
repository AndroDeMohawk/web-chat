package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/AndroDeMohawk/web-chat/internal/repository"
	"github.com/AndroDeMohawk/web-chat/internal/repository/db"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrUserInvalidCredentials = errors.New("user invalid credentials")
)

type AuthService struct {
	repo *repository.Repository
}

func NewAuthService(repo *repository.Repository) *AuthService {
	return &AuthService{repo: repo}
}
func (s *AuthService) Register(ctx context.Context, username, password string) (*db.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	user, err := s.repo.CreateUser(ctx, db.CreateUserParams{
		Username:     username,
		PasswordHash: string(hashedPassword),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return &user, err
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*db.User, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, ErrUserInvalidCredentials
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrUserInvalidCredentials
	}
	return &user, nil
}
