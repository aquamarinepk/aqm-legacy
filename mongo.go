package aqm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoConfig encapsulates the parameters required to connect to MongoDB.
type MongoConfig struct {
	URI            string
	Database       string
	ConnectTimeout time.Duration
}

// MongoClient is a thin wrapper over the official driver that implements a
// simple lifecycle and exposes typed helpers friendly to services.
type MongoClient struct {
	client   *mongo.Client
	database string
}

// NewMongoClient establishes a new MongoDB connection.
func NewMongoClient(ctx context.Context, cfg MongoConfig) (*MongoClient, error) {
	if cfg.URI == "" {
		return nil, errors.New("mongo uri is required")
	}
	if cfg.Database == "" {
		return nil, errors.New("mongo database is required")
	}
	connectTimeout := cfg.ConnectTimeout
	if connectTimeout <= 0 {
		connectTimeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	return &MongoClient{client: client, database: cfg.Database}, nil
}

// Collection returns a typed collection handle.
func (m *MongoClient) Collection(name string) *mongo.Collection {
	return m.client.Database(m.database).Collection(name)
}

// Disconnect closes the underlying client.
func (m *MongoClient) Disconnect(ctx context.Context) error {
	if m == nil || m.client == nil {
		return nil
	}
	return m.client.Disconnect(ctx)
}
