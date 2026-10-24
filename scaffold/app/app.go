// Copyright (c) 2026, Example Org.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package app holds startup shared by every deployment: logging, migrations,
// DB, dependency graph, HTTP server, and signal-driven shutdown.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"example.com/widgetsvc/api"
	"example.com/widgetsvc/config"
	"example.com/widgetsvc/db"
	dbmigrations "example.com/widgetsvc/db_migrations"
	"example.com/widgetsvc/middleware"
	"example.com/widgetsvc/wiring"
)

// Options are the deployment-specific choices main makes.
type Options struct {
	// Server starts the HTTP server.
	Server bool
	// Migrate applies database migrations before anything else.
	Migrate bool
	// TokenVerifier authenticates API callers. Required when Server is set.
	TokenVerifier middleware.TokenVerifier
}

func Run(cfg *config.Config, opts Options) {
	setupLogger(cfg)

	database, err := db.Open(cfg.POSTGRESQL)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	if opts.Migrate {
		if err := dbmigrations.Migrate(database); err != nil {
			slog.Error("Failed to migrate database", "error", err)
			os.Exit(1)
		}
	}
	if !opts.Server {
		return
	}
	if opts.TokenVerifier == nil {
		slog.Error("Refusing to start: no token verifier configured")
		os.Exit(1)
	}

	params := wiring.Build(database, opts.TokenVerifier)
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:        api.MakeHTTPHandler(cfg, params),
		ReadTimeout:    time.Duration(cfg.ReadTimeoutSeconds) * time.Second,
		WriteTimeout:   time.Duration(cfg.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:    time.Duration(cfg.IdleTimeoutSeconds) * time.Second,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
	}

	go func() {
		slog.Info("Starting HTTP server", "port", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	slog.Info("Shutdown signal received, stopping server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownSeconds)*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Graceful shutdown failed", "error", err)
	}
}

func setupLogger(cfg *config.Config) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(strings.ToUpper(cfg.LogLevel))); err != nil {
		level = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))
}
