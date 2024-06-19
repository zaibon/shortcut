package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"

	_ "github.com/zaibon/shortcut/db/migrations"
)

type Migration struct {
	db *sql.DB
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

func NewMigration(pool *pgxpool.Pool) (*Migration, error) {
	if pool == nil {
		return &Migration{}, errors.New("pool is nil")
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return &Migration{}, err
	}

	cp := pool.Config().ConnConfig.ConnString()
	db, err := sql.Open("pgx/v5", cp)
	if err != nil {
		return &Migration{}, err
	}

	return &Migration{db: db}, nil
}

func (m *Migration) Run(ctx context.Context, command string, args ...string) error {
	if err := goose.RunContext(ctx, command, m.db, "migrations", args...); err != nil {
		return err
	}
	return nil
}

func (m *Migration) Up(ctx context.Context) error {
	if err := goose.UpContext(ctx, m.db, "migrations"); err != nil {
		return err
	}
	return nil
}

func (m *Migration) Down(ctx context.Context) error {
	if err := goose.DownContext(ctx, m.db, "migrations"); err != nil {
		return err
	}
	return nil
}

func MigrateCmd(ctx context.Context, pool *pgxpool.Pool, command string, args ...string) error {
	m, err := NewMigration(pool)
	if err != nil {
		return fmt.Errorf("unable to create migration: %v", err)
	}

	if err := m.Run(ctx, command, args...); err != nil {
		return fmt.Errorf("goose %v: %v", command, err)
	}

	return nil
}
