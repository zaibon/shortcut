package services

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Txer interface {
	Tx(ctx context.Context, fn func(context.Context) error, opts pgx.TxOptions) error
}
