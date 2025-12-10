package aqm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewAuthzClient(t *testing.T) {
	client := NewAuthzClient("http://localhost:8083")
	if client == nil {
		t.Fatal("NewAuthzClient returned nil")
	}
	if client.client == nil {
		t.Error("service client should not be nil")
	}
}

func TestAuthzClientCheckPermissionAllowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/authz/policy/evaluate" {
			t.Errorf("expected /authz/policy/evaluate, got %s", r.URL.Path)
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["user_id"] != "user-123" {
			t.Errorf("expected user_id user-123, got %v", body["user_id"])
		}
		if body["permission"] != "read" {
			t.Errorf("expected permission read, got %v", body["permission"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SuccessResponse{
			Data: map[string]interface{}{"allowed": true},
		})
	}))
	defer server.Close()

	client := NewAuthzClient(server.URL)
	allowed, err := client.CheckPermission(context.Background(), "user-123", "read", "resource-456")

	if err != nil {
		t.Fatalf("CheckPermission error: %v", err)
	}
	if !allowed {
		t.Error("expected allowed to be true")
	}
}

func TestAuthzClientCheckPermissionDenied(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SuccessResponse{
			Data: map[string]interface{}{"allowed": false},
		})
	}))
	defer server.Close()

	client := NewAuthzClient(server.URL)
	allowed, err := client.CheckPermission(context.Background(), "user-123", "write", "resource-456")

	if err != nil {
		t.Fatalf("CheckPermission error: %v", err)
	}
	if allowed {
		t.Error("expected allowed to be false")
	}
}

func TestAuthzClientCheckPermissionGlobalScope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		scope, ok := body["scope"].(map[string]interface{})
		if !ok {
			t.Error("scope should be a map")
		}
		if scope["type"] != "global" {
			t.Errorf("expected scope type global, got %v", scope["type"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SuccessResponse{
			Data: map[string]interface{}{"allowed": true},
		})
	}))
	defer server.Close()

	client := NewAuthzClient(server.URL)

	// Test with "*" resource
	_, err := client.CheckPermission(context.Background(), "user-123", "admin", "*")
	if err != nil {
		t.Fatalf("CheckPermission error: %v", err)
	}

	// Test with empty resource
	_, err = client.CheckPermission(context.Background(), "user-123", "admin", "")
	if err != nil {
		t.Fatalf("CheckPermission error: %v", err)
	}
}

func TestAuthzClientCheckPermissionInvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SuccessResponse{
			Data: "invalid", // not a map
		})
	}))
	defer server.Close()

	client := NewAuthzClient(server.URL)
	_, err := client.CheckPermission(context.Background(), "user-123", "read", "resource")

	if err == nil {
		t.Fatal("expected error for invalid response format")
	}
}

func TestAuthzClientCheckPermissionMissingAllowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SuccessResponse{
			Data: map[string]interface{}{"other": "value"}, // missing "allowed"
		})
	}))
	defer server.Close()

	client := NewAuthzClient(server.URL)
	_, err := client.CheckPermission(context.Background(), "user-123", "read", "resource")

	if err == nil {
		t.Fatal("expected error for missing allowed field")
	}
}

func TestAuthzClientCheckPermissionHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewAuthzClient(server.URL)
	client.client.http.MaxRetries = 0
	_, err := client.CheckPermission(context.Background(), "user-123", "read", "resource")

	if err == nil {
		t.Fatal("expected error for HTTP error")
	}
}

func TestNewAuthzHelper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SuccessResponse{
			Data: map[string]interface{}{"allowed": true},
		})
	}))
	defer server.Close()

	client := NewAuthzClient(server.URL)
	helper := NewAuthzHelper(client, 5*time.Minute)

	if helper == nil {
		t.Fatal("NewAuthzHelper returned nil")
	}
}
