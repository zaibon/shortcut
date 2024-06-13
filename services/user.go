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
		Password:     string(saltedPasswd.Hash),
		PasswordSalt: string(saltedPasswd.Salt),
	})
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
		ID:        user.ID,
		Name:      user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
	}, nil
}
