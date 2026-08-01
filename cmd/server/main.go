package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gambitier/go-pkgs/logging"
	appitem "github.com/gambitier/golang-service-template/internal/application/item"
	"github.com/gambitier/golang-service-template/internal/config"
	"github.com/gambitier/golang-service-template/internal/infrastructure/persistence/mongodb"
	"github.com/gambitier/golang-service-template/internal/platform"
	"github.com/gambitier/golang-service-template/internal/server"

	_ "github.com/gambitier/golang-service-template/swagger"
)

// @title           Golang Service Template API
// @version         1.0
// @description     Hexagonal Go service template with MongoDB persistence adapter.
// @host            localhost:8080
// @BasePath        /items/api/v1
// @contact.name    API Support
// @contact.url     https://github.com/gambitier/golang-service-template
func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	env := flag.String("env", "", "environment overlay (e.g. development)")
	flag.Parse()

	bootstrap := logging.NewDefault("golang-service-template")

	resolvedConfig, err := resolveConfigPath(*configPath)
	if err != nil {
		bootstrap.Error("resolve config path", err, nil)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(bootstrap, resolvedConfig, *env)
	if err != nil {
		bootstrap.Error("load config", err, nil)
		os.Exit(1)
	}

	logger, err := platform.NewLogger(cfg.Logging)
	if err != nil {
		bootstrap.Error("create logger", err, nil)
		os.Exit(1)
	}
	logger.Info("config loaded", logging.Fields{
		"path": resolvedConfig,
		"env":  cfg.Server.Env.String(),
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	otelCfg := cfg.Opentel
	if strings.TrimSpace(otelCfg.ServiceName) == "" {
		otelCfg.ServiceName = "golang-service-template"
	}
	otelShutdown, err := platform.InitObservability(ctx, otelCfg, logger)
	if err != nil {
		logger.Error("init observability", err, nil)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.HTTP.WriteTimeout)
		defer cancel()
		if err := otelShutdown(shutdownCtx); err != nil {
			logger.Error("observability shutdown", err, nil)
		}
	}()

	// Persistence adapter: swap Mongo for Postgres (or another driver) here only.
	// Domain and application layers depend on domain/item.Repository, not Mongo types.
	client, db, err := mongodb.Connect(ctx, cfg.Mongo.URI, cfg.Mongo.Database)
	if err != nil {
		logger.Error("connect mongo", err, nil)
		os.Exit(1)
	}
	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.HTTP.WriteTimeout)
		defer cancel()
		if err := client.Disconnect(disconnectCtx); err != nil {
			logger.Error("mongo disconnect", err, nil)
		}
	}()

	persistence, err := mongodb.InitializePersistence(ctx, db)
	if err != nil {
		logger.Error("initialize persistence", err, nil)
		os.Exit(1)
	}

	itemSvc := appitem.NewService(persistence.Repositories.Item)

	srv, err := server.New(cfg, logger, itemSvc)
	if err != nil {
		logger.Error("create server", err, nil)
		os.Exit(1)
	}

	logger.Info("starting server", logging.Fields{"port": cfg.Server.HTTP.Port})
	if err := srv.Start(ctx); err != nil {
		logger.Error("server stopped", err, nil)
		os.Exit(1)
	}
}

func resolveConfigPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("config file %q: %w", path, err)
		}
		return path, nil
	}

	candidates := []string{
		path,
		filepath.Join(".", path),
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, path))
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			abs, err := filepath.Abs(candidate)
			if err != nil {
				return candidate, nil
			}
			return abs, nil
		}
	}
	return "", fmt.Errorf("config file %q not found", path)
}
