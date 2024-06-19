package db

import (
	"context"
	"fmt"

	"github.com/zaibon/shortcut/db/datastore"
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
