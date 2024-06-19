package domain

import (
	"encoding/gob"
	"time"
)

type User struct {
	ID        ID
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
}

func init() {
	// register into gob for redis session store
	gob.Register(&User{})
}
