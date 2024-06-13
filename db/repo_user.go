package db

import (
	"context"
	"fmt"

	"github.com/zaibon/shortcut/db/datastore"
	"github.com/zaibon/shortcut/domain"
)

type userStore struct {
	db datastore.Querier
}

func NewUserStore(db datastore.DBTX) *userStore {
	return &userStore{
		db: datastore.New(db),
	}
}

func (s *userStore) InsertUser(ctx context.Context, user datastore.InsertUserParams) error {
	_, err := s.db.InsertUser(ctx, user)
	if err != nil {
		return fmt.Errorf(
			"failed to insert user %s: %w",
			user.Username,
			err,
		)
	}

	return nil
}

func (s *userStore) UpdateUser(ctx context.Context, id domain.ID, user *domain.User) (*domain.User, error) {
	row, err := s.db.UpdateUser(ctx, datastore.UpdateUserParams{
		Username: user.Name,
		Email:    user.Email,
		ID:       int32(id),
	})
	if err != nil {
		return nil, fmt.Errorf(
			"failed to update user %s: %w",
			user.Name,
			err,
		)
	}

	return &domain.User{
		ID:        domain.ID(row.ID),
		Name:      row.Username,
		Email:     row.Email,
		Password:  "",
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (s *userStore) UpdatePassword(ctx context.Context, id domain.ID, newPassword, newSalt []byte) error {
	err := s.db.UpdatePassword(ctx, datastore.UpdatePasswordParams{
		Password:     newPassword,
		PasswordSalt: newSalt,
		ID:           int32(id),
	})
	if err != nil {
		return fmt.Errorf(
			"failed to update password for user %d: %w",
			id,
			err,
		)
	}
	return nil
}

func (s *userStore) GetUser(ctx context.Context, email string) (datastore.User, error) {
	row, err := s.db.GetUser(ctx, email)
	if err != nil {
		return datastore.User{}, fmt.Errorf(
			"failed to verify login for user %s: %w",
			email,
			err,
		)
	}

	return row, nil
}
