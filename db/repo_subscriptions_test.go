package db

import (
	"context"
	"testing"

	"github.com/zaibon/shortcut/db/datastore"
)

// TestRepoSubscriptions_q guards the transaction fix: repo methods must use the
// tx-bound querier when one is stashed in the context, and the pool-bound one
// otherwise. If this routing breaks, Tx() silently runs queries outside the
// transaction (the original bug).
func TestRepoSubscriptions_q(t *testing.T) {
	pool := datastore.New(nil)
	r := &repoSubscriptions{db: pool}

	if got := r.q(context.Background()); got != pool {
		t.Fatalf("without tx in context: got %p, want pool %p", got, pool)
	}

	txQ := pool.WithTx(nil)
	ctx := context.WithValue(context.Background(), txKey{}, txQ)
	if got := r.q(ctx); got != txQ {
		t.Fatalf("with tx in context: got %p, want tx querier %p", got, txQ)
	}
}
