package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// defaultShutdownTimeout は SIGTERM 受信後に進行中リクエストを待つ既定値。
// Kubernetes の terminationGracePeriodSeconds (デフォルト 30s) と揃えてある。
const defaultShutdownTimeout = 30 * time.Second

// Config はアプリ起動時に env から読み取る全設定値の集合。env への直接アクセス
// はこのパッケージ内に閉じ、他のパッケージは Config 経由で値を受け取る。
type Config struct {
	AppEnv          string
	DatabaseURL     string
	Port            string
	ShutdownTimeout time.Duration
}

// Load は環境変数からアプリ設定を読み取り、検証する。development では .env を
// 自動 load する。検証失敗時はそのエラーを返し Config は nil。
//
// 設定項目:
//   - APP_ENV (必須): "" は不正、"development" / "production" / "test" /
//     "staging" のみ許容
//   - DATABASE_URL (必須)
//   - PORT (任意): 既定 "8080"
//   - SHUTDOWN_TIMEOUT_SECONDS (任意): 既定 30
func Load() (*Config, error) {
	appEnv := os.Getenv("APP_ENV")
	switch appEnv {
	case "":
		return nil, errors.New("APP_ENV is not set")
	case "development":
		if err := godotenv.Load(); err != nil {
			return nil, fmt.Errorf("failed to load .env: %w", err)
		}
	case "production", "test", "staging":
		// OK
	default:
		return nil, fmt.Errorf("invalid APP_ENV: %q", appEnv)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, errors.New("DATABASE_URL is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	shutdownTimeout := defaultShutdownTimeout
	if raw := os.Getenv("SHUTDOWN_TIMEOUT_SECONDS"); raw != "" {
		secs, err := strconv.Atoi(raw)
		if err != nil || secs <= 0 {
			return nil, fmt.Errorf("invalid SHUTDOWN_TIMEOUT_SECONDS: %q", raw)
		}
		shutdownTimeout = time.Duration(secs) * time.Second
	}

	return &Config{
		AppEnv:          appEnv,
		DatabaseURL:     dsn,
		Port:            port,
		ShutdownTimeout: shutdownTimeout,
	}, nil
}
