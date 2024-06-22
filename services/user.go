package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/zaibon/shortcut/db/datastore"
	"github.com/zaibon/shortcut/domain"
	"github.com/zaibon/shortcut/services/password"
)

type userStore interface {
	InsertUser(ctx context.Context, user datastore.InsertUserParams) error
	InsertUserOauth(ctx context.Context, user datastore.InsertUserOauthParams) error
	GetUser(ctx context.Context, email string) (datastore.User, error)
	UpdateUser(ctx context.Context, id domain.ID, user *domain.User) (*domain.User, error)
	UpdatePassword(ctx context.Context, id domain.ID, password, salt []byte) error

	// oauth
	InsertOauthState(ctx context.Context, state string) error
	GetOauthState(ctx context.Context, state string) (datastore.Oauth2State, error)
}

var (
	ErrInvalidCredentials = fmt.Errorf("invalid credentials")
	ErrUserNotFound       = fmt.Errorf("user not found")
)

type userService struct {
	store userStore

	TLS               bool
	ownDomain         string
	googleOauthConfig *oauth2.Config
}

func NewUser(store userStore, ownDomain string, tls bool, googleClientID, googleClientSecret string) *userService {

	return &userService{
		store:     store,
		ownDomain: ownDomain,
		googleOauthConfig: &oauth2.Config{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  oauthRedirectURL(ownDomain, tls),
			Scopes: []string{
				"openid", "email", "profile",
			},
		},
	}
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

func (s *userService) VerifyOauthState(ctx context.Context, state string) (bool, error) {
	row, err := s.store.GetOauthState(ctx, state)
	if err != nil {
		return false, fmt.Errorf("failed to verify oauth state: %w", err)
	}
	// Check if the state has expired
	if time.Now().After(row.ExpireAt.Time) {
		return false, fmt.Errorf("oauth state has expired")
	}

	return true, nil
}

func (s *userService) IdentifyOauthUser(ctx context.Context, code string) (*domain.User, error) {
	token, err := s.googleOauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	client := s.googleOauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	userInfo := domain.GoogleUserInfo{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	user, err := s.store.GetUser(ctx, userInfo.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to identify user %s: %w", user.Email, err)
	}
	if errors.Is(err, sql.ErrNoRows) {
		if err := s.store.InsertUserOauth(ctx, datastore.InsertUserOauthParams{
			Username: userInfo.Name,
			Email:    user.Email,
		}); err != nil {
			return nil, fmt.Errorf("failed to identify user %s: %w", userInfo.Email, err)
		}
	}

	return &domain.User{
		ID:        domain.ID(user.ID),
		Name:      user.Username,
		Email:     user.Email,
		Password:  "",
		CreatedAt: user.CreatedAt.Time,
		IsOauth:   true,
	}, nil
}

func (s *userService) InitiateOauthFlow(ctx context.Context) (string, error) {
	state := uuid.New().String()
	if err := s.store.InsertOauthState(ctx, state); err != nil {
		return "", fmt.Errorf("failed to initiate oauth flow: %w", err)
	}

	return s.googleOauthConfig.AuthCodeURL(state), nil
}

func oauthRedirectURL(domain string, tls bool) string {
	url := fmt.Sprintf("%s/oauth/callback", domain)
	if tls {
		return fmt.Sprintf("https://%s", url)
	} else {
		return fmt.Sprintf("http://%s", url)
	}
}
