package migrations

import (
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"log"
	"pg-test-task-2024/internal/config"
	"strings"
)

func Apply() {
	log.Println("applying migrations...")
	after, _ := strings.CutPrefix(config.GetDbConnStr(), "postgres")
	m, err := migrate.New("file://scripts/migrations", "pgx"+after)
	if err != nil {
		log.Fatalf("failed to prepare migrations: %v", err)
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("failed to apply migrations: %v", err)
	}
}
