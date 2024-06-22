package domain

import (
	"encoding/gob"
	"time"
)

func init() {
	// register into gob for redis session store
	gob.Register(&User{})
	gob.Register(&GoogleUserInfo{})
}

type User struct {
	ID        ID
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
	IsOauth   bool
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
