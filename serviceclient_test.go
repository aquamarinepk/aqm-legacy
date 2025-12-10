package aqm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewServiceClient(t *testing.T) {
	client := NewServiceClient("http://localhost:8080")
	if client == nil {
		t.Fatal("NewServiceClient returned nil")
	}
	if client.baseURL != "http://localhost:8080" {
		t.Errorf("baseURL = %s, want http://localhost:8080", client.baseURL)
	}
	if client.http == nil {
		t.Error("http client should not be nil")
	}
}

func TestServiceClientList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/users" {
			t.Errorf("expected /users, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SuccessResponse{Data: []string{"user1", "user2"}})
	}))
	defer server.Close()

	client := NewServiceClient(server.URL)
	resp, err := client.List(context.Background(), "users")

	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if resp == nil {
		t.Fatal("response should not be nil")
	}
}

func TestServiceClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/users/123" {
			t.Errorf("expected /users/123, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SuccessResponse{Data: map[string]string{"id": "123"}})
	}))
	defer server.Close()

	client := NewServiceClient(server.URL)
	resp, err := client.Get(context.Background(), "users", "123")

	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if resp == nil {
		t.Fatal("response should not be nil")
	}
}

func TestServiceClientCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/users" {
			t.Errorf("expected /users, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SuccessResponse{Data: map[string]string{"id": "new"}})
	}))
	defer server.Close()

	client := NewServiceClient(server.URL)
	body := map[string]string{"name": "test"}
	resp, err := client.Create(context.Background(), "users", body)

	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if resp == nil {
		t.Fatal("response should not be nil")
	}
}

func TestServiceClientUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/users/123" {
			t.Errorf("expected /users/123, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SuccessResponse{Data: map[string]string{"id": "123"}})
	}))
	defer server.Close()

	client := NewServiceClient(server.URL)
	body := map[string]string{"name": "updated"}
	resp, err := client.Update(context.Background(), "users", "123", body)

	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if resp == nil {
		t.Fatal("response should not be nil")
	}
}

func TestServiceClientDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/users/123" {
			t.Errorf("expected /users/123, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewServiceClient(server.URL)
	err := client.Delete(context.Background(), "users", "123")

	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestServiceClientRequest(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"GET", "GET", "/test"},
		{"POST", "POST", "/test"},
		{"PUT", "PUT", "/test"},
		{"PATCH", "PATCH", "/test"},
		{"DELETE", "DELETE", "/test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.method {
					t.Errorf("expected %s, got %s", tt.method, r.Method)
				}
				if tt.method == "DELETE" {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(SuccessResponse{Data: "ok"})
			}))
			defer server.Close()

			client := NewServiceClient(server.URL)
			resp, err := client.Request(context.Background(), tt.method, tt.path, nil)

			if err != nil {
				t.Fatalf("Request error: %v", err)
			}
			if resp == nil {
				t.Fatal("response should not be nil")
			}
		})
	}
}

func TestServiceClientRequestUnsupportedMethod(t *testing.T) {
	client := NewServiceClient("http://localhost")
	_, err := client.Request(context.Background(), "INVALID", "/test", nil)

	if err == nil {
		t.Fatal("expected error for unsupported method")
	}
}

func TestServiceClientPing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" {
			t.Errorf("expected /healthz, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewServiceClient(server.URL)
	err := client.Ping(context.Background())

	if err != nil {
		t.Fatalf("Ping error: %v", err)
	}
}

func TestServiceClientError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	client := NewServiceClient(server.URL)
	client.http.MaxRetries = 0
	_, err := client.List(context.Background(), "nonexistent")

	if err == nil {
		t.Fatal("expected error")
	}
}
