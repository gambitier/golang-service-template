package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gambitier/go-pkgs/logging"
	"github.com/gambitier/golang-service-template/internal/config"
	itemapp "github.com/gambitier/golang-service-template/internal/item/application"
	itemmongo "github.com/gambitier/golang-service-template/internal/item/infrastructure/mongodb"
	itemhttp "github.com/gambitier/golang-service-template/internal/item/presentation/http"
	"github.com/gambitier/golang-service-template/internal/shared/infrastructure/persistence/mongodb"
	"github.com/gambitier/golang-service-template/internal/shared/infrastructure/persistence/persistopts"
	"github.com/gambitier/golang-service-template/internal/shared/lifecycle"
	"github.com/gambitier/golang-service-template/internal/shared/platform"
	"github.com/gambitier/golang-service-template/internal/shared/server"
)

// Options are CLI inputs for process startup.
type Options struct {
	ConfigPath string
	Env        string
}

// Run loads config, registers lifecycle components, and blocks until ctx is done.
// Graceful shutdown of all components is owned by lifecycle.App.
func Run(ctx context.Context, opts Options) error {
	bootstrapLog := logging.NewDefault("golang-service-template")

	resolvedConfig, err := resolveConfigPath(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("resolve config path: %w", err)
	}

	cfg, err := config.LoadConfig(bootstrapLog, resolvedConfig, opts.Env)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := platform.NewLogger(cfg.Logging)
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}
	logger.Info("config loaded", logging.Fields{
		"path": resolvedConfig,
		"env":  cfg.Server.Env.String(),
	})

	otelCfg := cfg.Opentel
	if strings.TrimSpace(otelCfg.ServiceName) == "" {
		otelCfg.ServiceName = "golang-service-template"
	}

	mongoComp := mongodb.NewComponent(cfg.Mongo.URI, cfg.Mongo.Database)
	otelComp := platform.NewOTelComponent(otelCfg, logger)
	httpComp := &httpComponent{
		cfg:    cfg,
		logger: logger,
		mongo:  mongoComp,
	}

	app := lifecycle.New(cfg.Server.HTTP.WriteTimeout)
	app.Add(mongoComp, otelComp, httpComp)

	logger.Info("starting server", logging.Fields{"port": cfg.Server.HTTP.Port})
	return app.Run(ctx)
}

// httpComponent builds the item stack after Mongo has started, then runs Fiber.
type httpComponent struct {
	cfg    *config.Config
	logger logging.Logger
	mongo  *mongodb.Component
	http   *server.HTTP
}

func (c *httpComponent) Name() string { return "http" }

func (c *httpComponent) Start(ctx context.Context) error {
	itemRepo, err := itemmongo.NewItemRepository(c.mongo.DB(), persistopts.Options{})
	if err != nil {
		return fmt.Errorf("item repository: %w", err)
	}
	itemSvc := itemapp.NewService(itemRepo)
	itemHandler := itemhttp.NewHandler(itemSvc)

	httpSrv, err := server.NewHTTP(c.cfg, c.logger, itemHandler)
	if err != nil {
		return err
	}
	c.http = httpSrv
	return c.http.Start(ctx)
}

func (c *httpComponent) Stop(ctx context.Context) error {
	if c.http == nil {
		return nil
	}
	return c.http.Stop(ctx)
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
