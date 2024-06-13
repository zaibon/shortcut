package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/zaibon/shortcut/db/datastore"
	"github.com/zaibon/shortcut/domain"
	"github.com/zaibon/shortcut/services/password"
)

type userStore interface {
	InsertUser(ctx context.Context, user datastore.InsertUserParams) error
	GetUser(ctx context.Context, email string) (datastore.User, error)
	UpdateUser(ctx context.Context, id domain.ID, user *domain.User) (*domain.User, error)
	UpdatePassword(ctx context.Context, id domain.ID, password, salt []byte) error
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
	hasher := password.DefaultArgon2iHasher()

	saltedPasswd, err := hasher.Hash([]byte(user.Password), nil)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.store.InsertUser(ctx, datastore.InsertUserParams{
		Username:     user.Name,
		Email:        user.Email,
		Password:     saltedPasswd.Hash,
		PasswordSalt: saltedPasswd.Salt,
	})
}

func (s *userService) UpdateUser(ctx context.Context, id domain.ID, user *domain.User) (*domain.User, error) {
	user, err := s.store.UpdateUser(ctx, id, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user %s: %w", user.Name, err)
	}

	return user, nil
}

func (s *userService) UpdatePassword(ctx context.Context, id domain.ID, newPassword string) error {
	hasher := password.DefaultArgon2iHasher()

	saltedPasswd, err := hasher.Hash([]byte(newPassword), nil)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.store.UpdatePassword(ctx, id, saltedPasswd.Hash, saltedPasswd.Salt); err != nil {
		return fmt.Errorf("failed to update user %d: %w", id, err)
	}

	return err
}

func (s *userService) VerifyLogin(ctx context.Context, email, passwd string) (*domain.User, error) {
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

	hasher := password.DefaultArgon2iHasher()
	err = hasher.Compare([]byte(user.Password), []byte(user.PasswordSalt), []byte(passwd))
	if err != nil {
		return nil, fmt.Errorf(
			"failed to verify login for user %s: %w",
			email,
			ErrInvalidCredentials,
		)
	}

	return &domain.User{
		ID:        domain.ID(user.ID),
		Name:      user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
	}, nil
}
