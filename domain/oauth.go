package domain

type OauthProvider string

var (
	OauthProviderGithub OauthProvider = "github"
	OauthProviderGoogle OauthProvider = "google"
)

func IsValidProvider(p string) bool {
	switch OauthProvider(p) {
	case OauthProviderGithub, OauthProviderGoogle:
		return true
	}
	return false
}
