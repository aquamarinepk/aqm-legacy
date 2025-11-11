package seed

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Seed represents a versioned, idempotent mutation that should run once per environment.
type Seed struct {
	ID          string
	Description string
	Run         func(ctx context.Context) error
}

// Record tracks the execution metadata for a seed.
type Record struct {
	ID          string    `bson:"_id"`
	Application string    `bson:"application"`
	Description string    `bson:"description"`
	AppliedAt   time.Time `bson:"applied_at"`
}

// Tracker persists which seeds have executed.
type Tracker interface {
	HasRun(ctx context.Context, id string) (bool, error)
	MarkRun(ctx context.Context, record Record) error
}

// Apply executes the provided seeds exactly once per tracker.
func Apply(ctx context.Context, tracker Tracker, seeds []Seed, application string) error {
	if tracker == nil {
		return errors.New("seed tracker is required")
	}

	for i, s := range seeds {
		if s.ID == "" {
			return fmt.Errorf("seed at index %d missing ID", i)
		}
		if s.Run == nil {
			return fmt.Errorf("seed %s missing Run function", s.ID)
		}

		ran, err := tracker.HasRun(ctx, s.ID)
		if err != nil {
			return fmt.Errorf("check seed %s status: %w", s.ID, err)
		}
		if ran {
			continue
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if err := s.Run(ctx); err != nil {
			return fmt.Errorf("seed %s failed: %w", s.ID, err)
		}

		record := Record{
			ID:          s.ID,
			Application: application,
			Description: s.Description,
			AppliedAt:   time.Now().UTC(),
		}
		if err := tracker.MarkRun(ctx, record); err != nil {
			return fmt.Errorf("mark seed %s as complete: %w", s.ID, err)
		}
	}

	return nil
}

const defaultCollectionName = "_seeds"

// MongoTracker stores seed records inside a MongoDB collection.
type MongoTracker struct {
	collection *mongo.Collection
}

// MongoTrackerOption configures a MongoTracker.
type MongoTrackerOption func(*mongoTrackerConfig)

type mongoTrackerConfig struct {
	collectionName string
}

// WithCollectionName overrides the default collection name used by MongoTracker.
func WithCollectionName(name string) MongoTrackerOption {
	return func(cfg *mongoTrackerConfig) {
		if trimmed := strings.TrimSpace(name); trimmed != "" {
			cfg.collectionName = trimmed
		}
	}
}

// NewMongoTracker creates a tracker that records seed executions in Mongo.
func NewMongoTracker(db *mongo.Database, opts ...MongoTrackerOption) *MongoTracker {
	cfg := mongoTrackerConfig{collectionName: defaultCollectionName}
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.collectionName == "" {
		cfg.collectionName = defaultCollectionName
	}

	return &MongoTracker{collection: db.Collection(cfg.collectionName)}
}

// HasRun reports whether a seed with the provided ID is already recorded.
func (t *MongoTracker) HasRun(ctx context.Context, id string) (bool, error) {
	if t == nil || t.collection == nil {
		return false, errors.New("mongo tracker is not initialized")
	}

	err := t.collection.FindOne(ctx, bson.M{"_id": id}).Err()
	if err == nil {
		return true, nil
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	return false, fmt.Errorf("query seed %s: %w", id, err)
}

// MarkRun inserts the provided record into the backing collection.
func (t *MongoTracker) MarkRun(ctx context.Context, record Record) error {
	if t == nil || t.collection == nil {
		return errors.New("mongo tracker is not initialized")
	}
	if record.ID == "" {
		return errors.New("seed record ID is required")
	}

	_, err := t.collection.InsertOne(ctx, record)
	if err != nil {
		return fmt.Errorf("insert seed record %s: %w", record.ID, err)
	}
	return nil
}

// UpsertOnce ensures a document exists by inserting it exactly once.
func UpsertOnce(ctx context.Context, collection *mongo.Collection, filter, document any) error {
	if collection == nil {
		return errors.New("collection is required")
	}
	if filter == nil {
		return errors.New("filter is required")
	}
	if document == nil {
		return errors.New("document is required")
	}

	_, err := collection.UpdateOne(
		ctx,
		filter,
		bson.M{"$setOnInsert": document},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("upsert document: %w", err)
	}
	return nil
}
