package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/minty1202/go-ddd-onion-template/internal/app"
	"github.com/minty1202/go-ddd-onion-template/internal/config"
)

func main() {
	// APP_ENV を先読みして slog の handler を選ぶ (development なら tint、
	// それ以外は JSON)。validation の本体は config.Load 側で行うため、ここで
	// は値の正当性は判定しない。空 / 不正値の場合は JSON ハンドラに
	// フォールバックし、続く config.Load がエラーを返して構造化ログとして
	// 失敗が出る。
	slog.SetDefault(newLogger(os.Getenv("APP_ENV")))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	a, err := app.New(ctx, cfg.DatabaseURL, cfg.Port, cfg.ShutdownTimeout)
	if err != nil {
		slog.Error("app build failed", "error", err)
		os.Exit(1)
	}
	defer a.Close()

	if err := a.Server.Run(ctx); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func newLogger(appEnv string) *slog.Logger {
	var handler slog.Handler
	if appEnv == "development" {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelInfo,
			TimeFormat: time.TimeOnly,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}
	return slog.New(handler)
}
