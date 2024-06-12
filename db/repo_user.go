package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zaibon/shortcut/db/datastore"
	"github.com/zaibon/shortcut/domain"
)

type userStore struct {
	db datastore.Querier
}

func NewUserStore(db *sql.DB) *userStore {
	return &userStore{
		db: datastore.New(db),
	}
}

func (s *userStore) InsertUser(ctx context.Context, user *domain.User) error {
	inserted, err := s.db.InsertUser(ctx, datastore.InsertUserParams{
		Username: user.Name,
		Email:    user.Email,
		Password: user.Password,
	})
	if err != nil {
		return fmt.Errorf(
			"failed to insert user %s: %w",
			user.Name,
			err,
		)
	}
	user.ID = inserted.ID //TODO: weirdo
	return nil
}

func (s *userStore) GetUser(ctx context.Context, email string) (*domain.User, error) {
	row, err := s.db.GetUser(ctx, email)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to verify login for user %s: %w",
			email,
			err,
		)
	}

	return &domain.User{
		ID:        row.ID,
		Name:      row.Username,
		Email:     row.Email,
		Password:  row.Password,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}
