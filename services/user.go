package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/zaibon/shortcut/domain"
)

type userStore interface {
	InsertUser(ctx context.Context, user *domain.User) error
	GetUser(ctx context.Context, email string) (*domain.User, error)
}

var (
	ErrInvalidCredentials = fmt.Errorf("invalid credentials")
	ErrUserNotFound       = fmt.Errorf("user not found")
)

type userService struct {
	store userStore
}

func NewUser(store userStore) *userService {
	return &userService{store: store}
}

func (s *userService) CreateUser(ctx context.Context, user *domain.User) error {
	return s.store.InsertUser(ctx, user)
}

func (s *userService) VerifyLogin(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.store.GetUser(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf(
			"failed to verify login for user %s: %w",
			email,
			err,
		)
	}

	if user.Password != password { //TODO: make it secure
		return nil, fmt.Errorf(
			"failed to verify login for user %s: %w",
			email,
			ErrInvalidCredentials,
		)
	}

	return user, nil
}
