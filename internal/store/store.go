package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dvl0p/linkctl/internal/db"
	"github.com/dvl0p/linkctl/sql/schema"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

type Store struct {
	db      *sql.DB
	Queries *db.Queries
}

func New(dbPath string) (*Store, error) {

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf(
			"could not create db parent dirs: %w", err,
		)
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf(
			"could not open sqlite db file: %w", err,
		)
	}

	goose.SetBaseFS(schema.FS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, fmt.Errorf(
			"could not set sqlite3 dialect in goose: %w", err,
		)
	}

	if err := goose.Up(conn, "."); err != nil {
		return nil, fmt.Errorf(
			"could not perform goose up migrations: %w", err,
		)
	}

	s := &Store{
		db:      conn,
		Queries: db.New(conn),
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
