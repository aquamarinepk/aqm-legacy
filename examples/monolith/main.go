package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aquamarinepk/aqm"
	"github.com/aquamarinepk/aqm/examples/monolith/internal/todo"
	"github.com/aquamarinepk/aqm/middleware"
)

const (
	namespace  = "TODO"
	appName    = "Todo"
	appVersion = "v0.0.1"
)

func main() {
	cfg, err := aqm.LoadConfig(namespace, os.Args[1:])
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}
	// Port is overwrited her for demo purposes;
	// the same key could be provided via YAML, env, or flags.
	cfg.Set("http.port", ":8080")

	log := aqm.NewLogger("debug")

	repo := todo.NewInMemoryRepo(log, cfg)
	service := todo.NewService(repo, log, cfg)
	handler := todo.NewHandler(service, log, cfg)

	stack := middleware.DefaultStack(middleware.StackOptions{
		Logger:        log,
		Timeout:       30 * time.Second,
		CompressLevel: 5,
	})

	ms := aqm.NewMs(
		aqm.WithConfig(cfg),
		aqm.WithLogger(log),
		aqm.WithHTTPMiddleware(stack...),
		aqm.WithHealthChecks("todo"),
		aqm.WithDebugRoutes(),
		aqm.WithLifecycle(repo, service),
		aqm.WithHTTPServerModules("http.port", handler),
	)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGTSTP,
	)
	defer stop()

	if err := ms.Run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s (%s) stopped with error: %v\n", appName, appVersion, err)
		os.Exit(1)
	}
}
