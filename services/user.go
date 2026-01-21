package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/zaibon/shortcut/db/datastore"
	"github.com/zaibon/shortcut/domain"
)

type userStore interface {
	InsertUserOauth(ctx context.Context, user datastore.InsertUserOauthParams) (domain.GUID, error)
	GetUser(ctx context.Context, email string) (datastore.User, error)

	// oauth
	InsertOauthState(ctx context.Context, state string, provider domain.OauthProvider) error
	GetOauthState(ctx context.Context, state string) (datastore.Oauth2State, error)
	InsertUserProvider(ctx context.Context, userID domain.GUID, provider domain.OauthProvider, providerUserID string) error

	GetUserProvider(ctx context.Context, userID uuid.UUID, provider domain.OauthProvider) (datastore.UserProvider, error)
	GetUserProviderByProviderUserId(ctx context.Context, provider domain.OauthProvider, providerUserID string) (datastore.UserProvider, error)
	DeleteUser(ctx context.Context, guid domain.GUID) error
	UpdateUserSuspension(ctx context.Context, guid domain.GUID, isSuspended bool) error
}

var (
	ErrInvalidCredentials = fmt.Errorf("invalid credentials")
	ErrUserNotFound       = fmt.Errorf("user not found")
)

type userService struct {
	store userStore

	TLS          bool
	oauthConfigs map[domain.OauthProvider]*oauth2.Config
}

func NewUser(store userStore, ownDomain string, tls bool,
	googleClientID, googleClientSecret string,
	githubClientID, githubClientSecret string,
) *userService {

	configs := map[domain.OauthProvider]*oauth2.Config{
		domain.OauthProviderGoogle: {
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  oauthRedirectURL(ownDomain, tls),
			Scopes: []string{
				"openid", "email", "profile",
			},
		},
		domain.OauthProviderGithub: {
			ClientID:     githubClientID,
			ClientSecret: githubClientSecret,
			Endpoint:     github.Endpoint, //FIXME: use github endpoint
			RedirectURL:  oauthRedirectURL(ownDomain, tls),
			Scopes: []string{
				"read:user", "user:email",
			},
		},
	}

	return &userService{
		store:        store,
		oauthConfigs: configs,
	}
}

func (s *userService) ToggleUserSuspension(ctx context.Context, guid domain.GUID, isSuspended bool) error {
	return s.store.UpdateUserSuspension(ctx, guid, isSuspended)
}

func (s *userService) VerifyOauthState(ctx context.Context, state string) (bool, domain.OauthProvider, error) {
	row, err := s.store.GetOauthState(ctx, state)
	if err != nil {
		return false, "", fmt.Errorf("failed to verify oauth state: %w", err)
	}
	// Check if the state has expired
	if time.Now().After(row.ExpireAt.Time) {
		return false, "", fmt.Errorf("oauth state has expired")
	}

	return true, domain.OauthProvider(row.Provider), nil
}

func (s *userService) IdentifyOauthUser(ctx context.Context, code string, provider domain.OauthProvider) (*domain.User, error) {
	oauthConfig, ok := s.oauthConfigs[provider]
	if !ok {
		return nil, fmt.Errorf("unsupported oauth provider: %s", provider)
	}

	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	userInfo, err := getUserInfo(ctx, oauthConfig, token, provider)
	if err != nil {
		return nil, err
	}

	_, errProvider := s.store.GetUserProviderByProviderUserId(ctx, provider, userInfo.ProviderID())
	if errProvider != nil && !errors.Is(errProvider, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to get user provider: %w", err)
	}

	user, errUser := s.store.GetUser(ctx, userInfo.ProviderEmail())
	if errUser != nil && !errors.Is(errUser, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to identify user %s: %w", user.Email, err)
	}

	// user not found, create it
	if errors.Is(errProvider, pgx.ErrNoRows) && errors.Is(errUser, pgx.ErrNoRows) {

		userID, err := s.store.InsertUserOauth(ctx, datastore.InsertUserOauthParams{
			Username: userInfo.ProviderName(),
			Email:    userInfo.ProviderEmail(),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to identify user %s: %w", userInfo.ProviderEmail(), err)
		}

		if err := s.store.InsertUserProvider(ctx, userID, provider, userInfo.ProviderID()); err != nil {
			return nil, fmt.Errorf("failed to insert user provider: %w", err)
		}
	}

	// user with this email exists, but not with this provider, link accounts
	if errUser == nil && errors.Is(errProvider, pgx.ErrNoRows) {
		if err := s.store.InsertUserProvider(ctx, domain.GUID(user.Guid.Bytes), provider, userInfo.ProviderID()); err != nil {
			return nil, fmt.Errorf("failed to insert user provider: %w", err)
		}
	}

	user, err = s.store.GetUser(ctx, userInfo.ProviderEmail())
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to identify user %s: %w", user.Email, err)
	}

	return &domain.User{
		GUID:        domain.GUID(user.Guid.Bytes),
		ID:          domain.ID(user.ID),
		Name:        userInfo.ProviderName(),
		Email:       userInfo.ProviderEmail(),
		Avatar:      userInfo.Avatar(),
		CreatedAt:   user.CreatedAt.Time,
		IsOauth:     true,
		Provider:    provider,
		IsSuspended: user.IsSuspended,
	}, nil
}

func (s *userService) InitiateOauthFlow(ctx context.Context, provider domain.OauthProvider) (string, error) {
	oauthConfig, ok := s.oauthConfigs[provider]
	if !ok {
		return "", fmt.Errorf("unsupported oauth provider: %s", provider)
	}
	state := uuid.New().String()

	if err := s.store.InsertOauthState(ctx, state, provider); err != nil {
		return "", fmt.Errorf("failed to initiate oauth flow: %w", err)
	}

	return oauthConfig.AuthCodeURL(state), nil
}

func getUserInfo(ctx context.Context, oauthConfig *oauth2.Config, token *oauth2.Token, provider domain.OauthProvider) (UserProvider, error) {
	client := oauthConfig.Client(ctx, token)
	var resp *http.Response
	var err error
	var user UserProvider

	switch provider {
	case domain.OauthProviderGoogle:
		resp, err = client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
		if err != nil {
			return nil, fmt.Errorf("failed to get user info: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		var googleUserInfo domain.GoogleUserInfo
		if err := json.NewDecoder(resp.Body).Decode(&googleUserInfo); err != nil {
			return nil, fmt.Errorf("failed to decode user info: %w", err)
		}
		user = googleUserInfo

	case domain.OauthProviderGithub:
		resp, err = client.Get("https://api.github.com/user")
		if err != nil {
			return nil, fmt.Errorf("failed to get user info: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		var githubUserInfo domain.GithubUserInfo
		if err := json.NewDecoder(resp.Body).Decode(&githubUserInfo); err != nil {
			return nil, fmt.Errorf("failed to decode user info: %w", err)
		}

		if githubUserInfo.Email == "" {
			resp, err = client.Get("https://api.github.com/user/emails")
			if err != nil {
				return nil, fmt.Errorf("failed to get user emails: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			var githubEmails []domain.GithubEmail
			if err := json.NewDecoder(resp.Body).Decode(&githubEmails); err != nil {
				return nil, fmt.Errorf("failed to decode user emails: %w", err)
			}
			for _, email := range githubEmails {
				if email.Primary {
					githubUserInfo.Email = email.Email
					break
				}
			}
		}
		user = githubUserInfo

	default:
		return nil, fmt.Errorf("unsupported oauth provider: %s", provider)
	}

	return user, nil
}

func (s *userService) ListConnectedProvider(ctx context.Context, userID domain.GUID) ([]domain.AccountProvider, error) {
	ap := make([]domain.AccountProvider, 2)

	for i, p := range []domain.OauthProvider{
		domain.OauthProviderGithub,
		domain.OauthProviderGoogle,
	} {
		_, err := s.store.GetUserProvider(ctx, uuid.UUID(userID), p)
		ap[i] = domain.AccountProvider{
			Provider:  p,
			Connected: err == nil,
		}
	}

	return ap, nil
}

func (s *userService) Delete(ctx context.Context, guid domain.GUID) error {
	return s.store.DeleteUser(ctx, guid)
}

func oauthRedirectURL(domain string, tls bool) string {
	url := fmt.Sprintf("%s/oauth/callback", domain)
	if tls {
		return fmt.Sprintf("https://%s", url)
	} else {
		return fmt.Sprintf("http://%s", url)
	}
}

type UserProvider interface {
	ProviderID() string
	ProviderName() string
	ProviderEmail() string
	Avatar() string
}
