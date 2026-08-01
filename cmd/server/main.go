package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/gambitier/go-pkgs/logging"
	"github.com/gambitier/golang-service-template/internal/shared/bootstrap"

	_ "github.com/gambitier/golang-service-template/swagger"
)

// @title           Golang Service Template API
// @version         1.0
// @description     Hexagonal Go service template with MongoDB persistence adapter.
// @host            localhost:8080
// @BasePath        /api/v1
// @contact.name    API Support
// @contact.url     https://github.com/gambitier/golang-service-template
func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	env := flag.String("env", "", "environment overlay (e.g. development)")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := bootstrap.Run(ctx, bootstrap.Options{
		ConfigPath: *configPath,
		Env:        *env,
	}); err != nil {
		logging.NewDefault("golang-service-template").Error("server stopped", err, nil)
		os.Exit(1)
	}
}
