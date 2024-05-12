package dbtest

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"pg-test-task-2024/internal/db/migrations"
	"testing"
	"time"
)

const (
	testDbUserName = "test_user"
	testDbPassword = "test_password"
	testDbName     = "test_db"
)

func CreateTestContainer(ctx context.Context, t *testing.T) *postgres.PostgresContainer {
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithUsername(testDbUserName),
		postgres.WithPassword(testDbPassword),
		postgres.WithDatabase(testDbName),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).WithStartupTimeout(5*time.Second)))
	if err != nil {
		t.Fatalf("failed to init test container: %s", err)
	}
	t.Cleanup(func() {
		err := pgContainer.Terminate(ctx)
		if err != nil {
			t.Fatalf("failed to terminate test container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("")
	}
	t.Setenv("EXECUTOR_DB_CONN_STR", connStr)
	t.Setenv("EXECUTOR_MIGRATIONS_SOURCE", "file://../../scripts/migrations")

	migrations.Apply()
	return pgContainer
}
