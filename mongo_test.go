package aqm

import (
	"context"
	"testing"
	"time"
)

func TestMongoConfigFields(t *testing.T) {
	cfg := MongoConfig{
		URI:            "mongodb://localhost:27017",
		Database:       "testdb",
		ConnectTimeout: 5 * time.Second,
	}

	if cfg.URI != "mongodb://localhost:27017" {
		t.Errorf("URI = %s, want mongodb://localhost:27017", cfg.URI)
	}
	if cfg.Database != "testdb" {
		t.Errorf("Database = %s, want testdb", cfg.Database)
	}
	if cfg.ConnectTimeout != 5*time.Second {
		t.Errorf("ConnectTimeout = %v, want 5s", cfg.ConnectTimeout)
	}
}

func TestNewMongoClientEmptyURI(t *testing.T) {
	cfg := MongoConfig{
		URI:      "",
		Database: "testdb",
	}

	_, err := NewMongoClient(context.Background(), cfg)
	if err == nil {
		t.Error("NewMongoClient should return error for empty URI")
	}
}

func TestNewMongoClientEmptyDatabase(t *testing.T) {
	cfg := MongoConfig{
		URI:      "mongodb://localhost:27017",
		Database: "",
	}

	_, err := NewMongoClient(context.Background(), cfg)
	if err == nil {
		t.Error("NewMongoClient should return error for empty database")
	}
}

func TestNewMongoClientInvalidURI(t *testing.T) {
	cfg := MongoConfig{
		URI:            "invalid-uri",
		Database:       "testdb",
		ConnectTimeout: 100 * time.Millisecond,
	}

	_, err := NewMongoClient(context.Background(), cfg)
	if err == nil {
		t.Error("NewMongoClient should return error for invalid URI")
	}
}

func TestMongoClientDisconnectNil(t *testing.T) {
	var client *MongoClient

	err := client.Disconnect(context.Background())
	if err != nil {
		t.Errorf("Disconnect on nil client should return nil, got %v", err)
	}
}

func TestMongoClientDisconnectNilInternalClient(t *testing.T) {
	client := &MongoClient{
		client:   nil,
		database: "testdb",
	}

	err := client.Disconnect(context.Background())
	if err != nil {
		t.Errorf("Disconnect on nil internal client should return nil, got %v", err)
	}
}

func TestMongoClientStructFields(t *testing.T) {
	client := &MongoClient{
		client:   nil, // We can't create a real client without mongo
		database: "testdb",
	}

	if client.database != "testdb" {
		t.Errorf("database = %s, want testdb", client.database)
	}
}
