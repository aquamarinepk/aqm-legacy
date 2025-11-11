package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/aquamarinepk/aqm"
	"github.com/aquamarinepk/aqm/examples/orchestration/pkg/shared/runtime"
	"github.com/aquamarinepk/aqm/examples/orchestration/services/tasks/internal/task"
)

const (
	namespace  = "TASKS"
	appName    = "Tasks"
	appVersion = "v0.0.1"
)

func main() {
	cfg, err := aqm.LoadConfig(namespace, os.Args[1:])
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	logLevel, _ := cfg.GetString("log.level")
	if logLevel == "" {
		logLevel = "info"
	}
	logger := aqm.NewLogger(logLevel)

	stack := runtime.MiddlewareStack(logger, runtime.MiddlewareConfig{})

	mongoURI, _ := cfg.GetString("mongo.uri")
	if mongoURI == "" {
		panic("mongo.uri is required")
	}
	mongoDB, _ := cfg.GetString("mongo.database")
	if mongoDB == "" {
		panic("mongo.database is required")
	}
	mongoCollection, ok := cfg.GetString("mongo.collection")
	if !ok || mongoCollection == "" {
		mongoCollection = "tasks"
	}

	mongoClient, err := aqm.NewMongoClient(context.Background(), aqm.MongoConfig{
		URI:      mongoURI,
		Database: mongoDB,
	})
	if err != nil {
		panic(fmt.Errorf("new mongo client: %w", err))
	}

	repo, err := task.NewMongoRepo(mongoClient.Collection(mongoCollection))
	if err != nil {
		panic(fmt.Errorf("new task repository: %w", err))
	}

	service := task.NewService(repo, logger, cfg)
	handler := task.NewHandler(service, logger, cfg)

	ms := aqm.NewMs(
		aqm.WithConfig(cfg),
		aqm.WithLogger(logger),
		aqm.WithHTTPMiddleware(stack...),
		aqm.WithHealthChecks("tasks"),
		aqm.WithDebugRoutes(),
		aqm.WithLifecycle(service),
		aqm.WithHTTPServerModules("http.port", handler),
		aqm.WithShutdown(func(ctx context.Context) error {
			return mongoClient.Disconnect(ctx)
		}),
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
