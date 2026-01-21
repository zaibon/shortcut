package domain

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ID int32
type GUID uuid.UUID

func (id GUID) IsNil() bool {
	return id == GUID(uuid.Nil)
}
func (id GUID) String() string {
	return uuid.UUID(id).String()
}
func (id GUID) PgType() pgtype.UUID {
	return pgtype.UUID{
		Bytes: uuid.UUID(id),
		Valid: !id.IsNil(),
	}
}

func ParseGUID(s string) (GUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return GUID(uuid.Nil), err
	}
	return GUID(id), nil
}
