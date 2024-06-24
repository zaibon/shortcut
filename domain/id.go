package domain

import "github.com/google/uuid"

type ID int32
type GUID uuid.UUID

func (id GUID) IsNil() bool {
	return id == GUID(uuid.Nil)
}
