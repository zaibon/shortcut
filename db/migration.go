package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/pressly/goose/v3"

	_ "github.com/zaibon/shortcut/db/migrations"
)

type Migration struct {
	db *sql.DB
}

func NewMigration(db *sql.DB) (*Migration, error) {
	goose.SetBaseFS(migrationsFS)

	return &Migration{db: db}, nil
}

func (m *Migration) Run(ctx context.Context, command string, args ...string) error {
	log.Printf("Running migrations")
	if err := goose.RunContext(ctx, command, m.db, "migrations", args...); err != nil {
		return err
	}
	log.Printf("Migrations done")
	return nil
}

func MigrateCmd(ctx context.Context, db *sql.DB, command string, args ...string) error {
	m, err := NewMigration(db)
	if err != nil {
		return fmt.Errorf("unable to create migration: %v", err)
	}

	if err := m.Run(ctx, command, args...); err != nil {
		return fmt.Errorf("goose %v: %v", command, err)
	}

	return nil
}
