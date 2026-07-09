package dbtest

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/minty1202/go-ddd-onion-template/sql/migrations"
)

var (
	pool     *pgxpool.Pool
	once     sync.Once
	setupErr error
)

// Setup は PostgreSQL コンテナを起動して（初回のみ）マイグレーションを適用し、
// テーブルを TRUNCATE してクリーンな状態の pool を返す。
// t.Parallel() は使えない（共有コンテナ + TRUNCATE のため）。
func Setup(t *testing.T) *pgxpool.Pool {
	t.Helper()
	once.Do(initContainer)
	if setupErr != nil {
		t.Fatalf("dbtest: setup failed: %v", setupErr)
	}
	truncateAll(t)
	return pool
}

func initContainer() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pgContainer, err := postgres.Run(ctx,
		image(),
		postgres.WithDatabase(dbName()),
		postgres.WithUsername(user()),
		postgres.WithPassword(password()),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		setupErr = fmt.Errorf("start container: %w", err)
		return
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		setupErr = fmt.Errorf("connection string: %w", err)
		return
	}

	pool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		setupErr = fmt.Errorf("create pool: %w", err)
		return
	}

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer func() { _ = sqlDB.Close() }()

	provider, err := goose.NewProvider(goose.DialectPostgres, sqlDB, migrations.FS)
	if err != nil {
		setupErr = fmt.Errorf("new goose provider: %w", err)
		return
	}
	if _, err := provider.Up(ctx); err != nil {
		setupErr = fmt.Errorf("goose up: %w", err)
	}
}

func truncateAll(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	rows, err := pool.Query(ctx, `
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public' AND tablename != 'goose_db_version'
	`)
	require.NoError(t, err, "dbtest: list tables")

	var tables []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		tables = append(tables, name)
	}
	rows.Close()

	if len(tables) == 0 {
		return
	}

	stmt := fmt.Sprintf("TRUNCATE %s RESTART IDENTITY CASCADE", strings.Join(tables, ", "))
	_, err = pool.Exec(ctx, stmt)
	require.NoError(t, err, "dbtest: truncate")
}

func image() string {
	if v := os.Getenv("POSTGRES_IMAGE"); v != "" {
		return v
	}
	return "postgres:17"
}

func user() string {
	base := os.Getenv("POSTGRES_USER")
	if base == "" {
		base = "postgres"
	}
	return base + "_test"
}

func password() string {
	base := os.Getenv("POSTGRES_PASSWORD")
	if base == "" {
		base = "postgres"
	}
	return base + "_test"
}

func dbName() string {
	base := os.Getenv("POSTGRES_DB")
	if base == "" {
		base = "app"
	}
	return base + "_test"
}
