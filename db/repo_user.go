package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

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

func (s *userStore) InsertUserOauth(ctx context.Context, user datastore.InsertUserOauthParams) error {
	if !user.Guid.Valid || user.Guid.Bytes == uuid.Nil {
		user.Guid = pgtype.UUID{
			Bytes: uuid.Must(uuid.NewV7()),
			Valid: true,
		}
	}

	_, err := s.db.InsertUserOauth(ctx, user)
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

func (s *userStore) InsertOauthState(ctx context.Context, state string, provider domain.OauthProvider) error {
	return s.db.InsertOauth2State(ctx, datastore.InsertOauth2StateParams{
		State:    state,
		Provider: string(provider),
	})
}

func (s *userStore) GetOauthState(ctx context.Context, state string) (datastore.Oauth2State, error) {
	return s.db.GetOauth2State(ctx, state)
}
