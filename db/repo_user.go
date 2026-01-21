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

func (s *userStore) InsertUserOauth(ctx context.Context, user datastore.InsertUserOauthParams) (domain.GUID, error) {
	if !user.Guid.Valid || user.Guid.Bytes == uuid.Nil {
		user.Guid = pgtype.UUID{
			Bytes: uuid.Must(uuid.NewV7()),
			Valid: true,
		}
	}

	_, err := s.db.InsertUserOauth(ctx, user)
	if err != nil {
		return domain.GUID(uuid.Nil), fmt.Errorf(
			"failed to insert user %s: %w",
			user.Username,
			err,
		)
	}

	return domain.GUID(user.Guid.Bytes), nil
}

func (s *userStore) GetUser(ctx context.Context, email string) (datastore.User, error) {
	row, err := s.db.GetUserByEmail(ctx, email)
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

func (s *userStore) InsertUserProvider(ctx context.Context, userID domain.GUID, provider domain.OauthProvider, providerUserID string) error {
	_, err := s.db.InsertUserProvider(ctx, datastore.InsertUserProviderParams{
		UserID: pgtype.UUID{
			Bytes: userID,
			Valid: uuid.UUID(userID) != uuid.Nil,
		},
		Provider:       string(provider),
		ProviderUserID: providerUserID,
	})
	if err != nil {
		return fmt.Errorf("failed to insert user provider %s: %w", provider, err)
	}
	return nil
}

func (s *userStore) GetUserProvider(ctx context.Context, userID uuid.UUID, provider domain.OauthProvider) (datastore.UserProvider, error) {
	row, err := s.db.GetUserProvider(ctx, datastore.GetUserProviderParams{
		UserID: pgtype.UUID{
			Bytes: userID,
			Valid: userID != uuid.Nil,
		},
		Provider: string(provider),
	})
	if err != nil {
		return datastore.UserProvider{}, fmt.Errorf("failed to get user provider %s: %w", provider, err)
	}
	return row, nil
}

func (s *userStore) GetUserProviderByProviderUserId(ctx context.Context, provider domain.OauthProvider, providerUserID string) (datastore.UserProvider, error) {
	row, err := s.db.GetUserProviderByProviderUserId(ctx, datastore.GetUserProviderByProviderUserIdParams{
		Provider:       string(provider),
		ProviderUserID: providerUserID,
	})
	if err != nil {
		return datastore.UserProvider{}, fmt.Errorf("failed to get user provider %s: %w", provider, err)
	}
	return row, nil
}

func (s *userStore) DeleteUser(ctx context.Context, guid domain.GUID) error {
	return s.db.DeleteUser(ctx, pgtype.UUID{
		Bytes: guid,
		Valid: uuid.UUID(guid) != uuid.Nil,
	})
}

func (s *userStore) UpdateUserSuspension(ctx context.Context, guid domain.GUID, isSuspended bool) error {
	return s.db.UpdateUserSuspension(ctx, datastore.UpdateUserSuspensionParams{
		Guid: pgtype.UUID{
			Bytes: guid,
			Valid: uuid.UUID(guid) != uuid.Nil,
		},
		IsSuspended: isSuspended,
	})
}
