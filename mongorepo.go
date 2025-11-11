package aqm

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoRepo is a generic aggregate repository backed by a Mongo collection.
// Aggregates must tag their ID field with `bson:"_id"` so the filters defined
// here work as expected.
type MongoRepo[T Identifiable] struct {
	collection *mongo.Collection
	factory    func() T
}

func NewMongoRepo[T Identifiable](collection *mongo.Collection, factory func() T) (*MongoRepo[T], error) {
	if collection == nil {
		return nil, errors.New("mongo collection is required")
	}
	if factory == nil {
		return nil, errors.New("mongo repository factory is required")
	}
	return &MongoRepo[T]{collection: collection, factory: factory}, nil
}

func (r *MongoRepo[T]) Save(ctx context.Context, aggregate T) error {
	if any(aggregate) == nil {
		return errors.New("aggregate cannot be nil")
	}
	filter := bson.M{"_id": aggregate.ID()}
	opts := options.Replace().SetUpsert(true)
	if _, err := r.collection.ReplaceOne(ctx, filter, aggregate, opts); err != nil {
		return fmt.Errorf("mongo save aggregate: %w", err)
	}
	return nil
}

func (r *MongoRepo[T]) FindByID(ctx context.Context, id uuid.UUID) (T, error) {
	var zero T
	res := r.collection.FindOne(ctx, bson.M{"_id": id})
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return zero, ErrRepoNotFound
		}
		return zero, fmt.Errorf("mongo find aggregate: %w", err)
	}
	aggregate := r.factory()
	if err := res.Decode(aggregate); err != nil {
		return zero, fmt.Errorf("mongo decode aggregate: %w", err)
	}
	return aggregate, nil
}

func (r *MongoRepo[T]) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("mongo delete aggregate: %w", err)
	}
	if result.DeletedCount == 0 {
		return ErrRepoNotFound
	}
	return nil
}

func (r *MongoRepo[T]) List(ctx context.Context, filter any) ([]T, error) {
	if filter == nil {
		filter = bson.M{}
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("mongo list aggregates: %w", err)
	}
	defer cursor.Close(ctx)

	var aggregates []T
	for cursor.Next(ctx) {
		aggregate := r.factory()
		if err := cursor.Decode(aggregate); err != nil {
			return nil, fmt.Errorf("mongo decode aggregate: %w", err)
		}
		aggregates = append(aggregates, aggregate)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("mongo cursor: %w", err)
	}
	return aggregates, nil
}
