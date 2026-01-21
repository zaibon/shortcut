package domain

import (
	"encoding/gob"
	"fmt"
	"time"
)

func init() {
	// register into gob for redis session store
	gob.Register(&User{})
	gob.Register(&GoogleUserInfo{})
	gob.Register(&GithubUserInfo{})
}

type User struct {
	ID          ID
	GUID        GUID
	Name        string
	Email       string
	Avatar      string
	CreatedAt   time.Time
	IsOauth     bool
	Provider    OauthProvider
	IsSuspended bool
}

type GoogleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	EmailVerified bool   `json:"email_verified"`
}

func (g GoogleUserInfo) ProviderID() string {
	return g.Email
}

func (g GoogleUserInfo) ProviderName() string {
	return g.Name
}

func (g GoogleUserInfo) ProviderEmail() string {
	return g.Email
}
func (g GoogleUserInfo) Avatar() string {
	return g.Picture
}

type GithubUserInfo struct {
	Login                   string    `json:"login"`
	ID                      int       `json:"id"`
	NodeID                  string    `json:"node_id"`
	AvatarURL               string    `json:"avatar_url"`
	GravatarID              string    `json:"gravatar_id"`
	URL                     string    `json:"url"`
	HTMLURL                 string    `json:"html_url"`
	FollowersURL            string    `json:"followers_url"`
	FollowingURL            string    `json:"following_url"`
	GistsURL                string    `json:"gists_url"`
	StarredURL              string    `json:"starred_url"`
	SubscriptionsURL        string    `json:"subscriptions_url"`
	OrganizationsURL        string    `json:"organizations_url"`
	ReposURL                string    `json:"repos_url"`
	EventsURL               string    `json:"events_url"`
	ReceivedEventsURL       string    `json:"received_events_url"`
	Type                    string    `json:"type"`
	SiteAdmin               bool      `json:"site_admin"`
	Name                    string    `json:"name"`
	Company                 string    `json:"company"`
	Blog                    string    `json:"blog"`
	Location                string    `json:"location"`
	Email                   string    `json:"email"`
	Hireable                bool      `json:"hireable"`
	Bio                     string    `json:"bio"`
	TwitterUsername         string    `json:"twitter_username"`
	PublicRepos             int       `json:"public_repos"`
	PublicGists             int       `json:"public_gists"`
	Followers               int       `json:"followers"`
	Following               int       `json:"following"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
	PrivateGists            int       `json:"private_gists"`
	TotalPrivateRepos       int       `json:"total_private_repos"`
	OwnedPrivateRepos       int       `json:"owned_private_repos"`
	DiskUsage               int       `json:"disk_usage"`
	Collaborators           int       `json:"collaborators"`
	TwoFactorAuthentication bool      `json:"two_factor_authentication"`
	Plan                    struct {
		Name          string `json:"name"`
		Space         int    `json:"space"`
		PrivateRepos  int    `json:"private_repos"`
		Collaborators int    `json:"collaborators"`
	} `json:"plan"`
}

func (g GithubUserInfo) ProviderID() string {
	return fmt.Sprintf("%d", g.ID)
}

func (g GithubUserInfo) ProviderName() string {
	return g.Login
}

func (g GithubUserInfo) ProviderEmail() string {
	return g.Email
}

func (g GithubUserInfo) Avatar() string {
	return g.AvatarURL
}

type GithubEmail struct {
	Email      string `json:"email"`
	Verified   bool   `json:"verified"`
	Primary    bool   `json:"primary"`
	Visibility string `json:"visibility"`
}

type SubscriptionStats struct {
	PlanName        string
	URLUsage        int
	URLLimit        int
	Remaining       int
	UsagePercentage int
}
