package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zaibon/shortcut/db/datastore"
	"github.com/zaibon/shortcut/domain"
	"github.com/zaibon/shortcut/log"
)

type txKey struct{}

type repoSubscriptions struct {
	pool *pgxpool.Pool
	db   *datastore.Queries
}

func NewRepoSubscription(db *pgxpool.Pool) *repoSubscriptions {
	return &repoSubscriptions{
		pool: db,
		db:   datastore.New(db),
	}
}

// q returns the transaction-bound querier when called inside Tx, otherwise the
// pool-bound one. This is what makes Tx an actual transaction.
func (r *repoSubscriptions) q(ctx context.Context) *datastore.Queries {
	if q, ok := ctx.Value(txKey{}).(*datastore.Queries); ok {
		return q
	}
	return r.db
}

func (r *repoSubscriptions) Tx(ctx context.Context, fn func(context.Context) error, opts pgx.TxOptions) error {
	tx, err := r.pool.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		// no-op if the tx already committed
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Error("error rolling back transaction", "err", err)
		}
	}()

	ctx = context.WithValue(ctx, txKey{}, r.db.WithTx(tx))
	if err := fn(ctx); err != nil {
		return fmt.Errorf("error in transaction: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}
	return nil
}

func (r *repoSubscriptions) GetCustomer(ctx context.Context, user *domain.User) (datastore.Customer, error) {
	row, err := r.q(ctx).GetCustomer(ctx, user.GUID.PgType())
	if err != nil {
		return datastore.Customer{}, fmt.Errorf("error getting customer with id %s: %w", user.GUID, err)
	}
	return row, nil
}

func (r *repoSubscriptions) GetCustomerByStripeId(ctx context.Context, id string) (datastore.Customer, error) {
	row, err := r.q(ctx).GetCustomerByStripeId(ctx, id)
	if err != nil {
		return datastore.Customer{}, fmt.Errorf("error getting customer with id %s: %w", id, err)
	}
	return row, nil
}

func (r *repoSubscriptions) InsertCustomer(ctx context.Context, customer datastore.InsertCustomerParams) (datastore.Customer, error) {
	row, err := r.q(ctx).InsertCustomer(ctx, customer)
	if err != nil {
		return datastore.Customer{}, fmt.Errorf("error inserting customer with id %s: %w", customer.StripeID, err)
	}
	return row, nil
}

func (r *repoSubscriptions) InsertSubscription(ctx context.Context, subscription datastore.InsertSubscriptionParams) error {
	_, err := r.q(ctx).InsertSubscription(ctx, subscription)
	if err != nil {
		return fmt.Errorf("error inserting subscription with id %s: %w", subscription.StripeID, err)
	}
	return nil
}

func (r repoSubscriptions) UpdateSubscription(ctx context.Context, subscription datastore.UpdateSubscriptionParams) error {
	err := r.q(ctx).UpdateSubscription(ctx, subscription)
	if err != nil {
		return fmt.Errorf("error updating subscription with id %s: %w", subscription.StripeID, err)
	}
	return nil
}

func (r *repoSubscriptions) ListSubscriptions(ctx context.Context, user *domain.User, status string) ([]datastore.Subscription, error) {
	row, err := r.q(ctx).ListCustomerSubscription(ctx, datastore.ListCustomerSubscriptionParams{
		CustomerID: user.GUID.PgType(),
		Status: pgtype.Text{
			String: status,
			Valid:  status != "",
		}})
	if err != nil {
		return nil, fmt.Errorf("error listing customer subscription: %w", err)
	}
	return row, nil
}
