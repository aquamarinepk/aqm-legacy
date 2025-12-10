package aqm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestRespondSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"key": "value"}
	links := []Link{{Rel: RelSelf, Href: "/test"}}

	RespondSuccess(rec, data, links...)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var resp SuccessResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Links) != 1 {
		t.Errorf("expected 1 link, got %d", len(resp.Links))
	}
}

func TestRespondError(t *testing.T) {
	rec := httptest.NewRecorder()

	RespondError(rec, http.StatusBadRequest, "test error")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error.Message != "test error" {
		t.Errorf("expected message 'test error', got %q", resp.Error.Message)
	}
}

func TestRespond(t *testing.T) {
	tests := []struct {
		name         string
		code         int
		data         interface{}
		meta         interface{}
		expectBody   bool
	}{
		{
			name:       "okWithData",
			code:       http.StatusOK,
			data:       map[string]string{"key": "value"},
			meta:       nil,
			expectBody: true,
		},
		{
			name:       "noContent",
			code:       http.StatusNoContent,
			data:       nil,
			meta:       nil,
			expectBody: false,
		},
		{
			name:       "createdWithMeta",
			code:       http.StatusCreated,
			data:       "created",
			meta:       map[string]int{"total": 1},
			expectBody: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			Respond(rec, tt.code, tt.data, tt.meta)

			if rec.Code != tt.code {
				t.Errorf("expected status %d, got %d", tt.code, rec.Code)
			}

			if tt.expectBody && rec.Body.Len() == 0 {
				t.Error("expected body but got empty")
			}
			if !tt.expectBody && rec.Body.Len() > 0 {
				t.Error("expected no body but got content")
			}
		})
	}
}

func TestError(t *testing.T) {
	rec := httptest.NewRecorder()
	details := []ValidationError{{Field: "name", Message: "required"}}

	Error(rec, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", details...)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code 'VALIDATION_ERROR', got %q", resp.Error.Code)
	}
	if len(resp.Error.Details) != 1 {
		t.Errorf("expected 1 detail, got %d", len(resp.Error.Details))
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		singular string
		want     string
	}{
		{"user", "users"},
		{"order", "orders"},
		{"child", "children"},
	}

	for _, tt := range tests {
		t.Run(tt.singular, func(t *testing.T) {
			got := Pluralize(tt.singular)
			if got != tt.want {
				t.Errorf("Pluralize(%q) = %q, want %q", tt.singular, got, tt.want)
			}
		})
	}
}

func TestSingularize(t *testing.T) {
	tests := []struct {
		plural string
		want   string
	}{
		{"users", "user"},
		{"orders", "order"},
		{"children", "child"},
	}

	for _, tt := range tests {
		t.Run(tt.plural, func(t *testing.T) {
			got := Singularize(tt.plural)
			if got != tt.want {
				t.Errorf("Singularize(%q) = %q, want %q", tt.plural, got, tt.want)
			}
		})
	}
}

func TestIsPlural(t *testing.T) {
	tests := []struct {
		word string
		want bool
	}{
		{"users", true},
		{"user", false},
		{"orders", true},
		{"order", false},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := IsPlural(tt.word)
			if got != tt.want {
				t.Errorf("IsPlural(%q) = %v, want %v", tt.word, got, tt.want)
			}
		})
	}
}

type testResource struct {
	id   uuid.UUID
	typ  string
}

func (r testResource) GetID() uuid.UUID    { return r.id }
func (r testResource) ResourceType() string { return r.typ }

func TestRESTfulLinksFor(t *testing.T) {
	id := uuid.New()
	obj := testResource{id: id, typ: "user"}

	links := RESTfulLinksFor(obj)

	if len(links) != 4 {
		t.Errorf("expected 4 links, got %d", len(links))
	}

	// Check self link exists
	var selfFound bool
	for _, link := range links {
		if link.Rel == RelSelf {
			selfFound = true
			if link.Href != "/users/"+id.String() {
				t.Errorf("unexpected self href: %s", link.Href)
			}
		}
	}
	if !selfFound {
		t.Error("expected self link")
	}
}

func TestRESTfulLinksForWithBasePath(t *testing.T) {
	id := uuid.New()
	obj := testResource{id: id, typ: "order"}

	links := RESTfulLinksFor(obj, "/api/v1")

	var selfFound bool
	for _, link := range links {
		if link.Rel == RelSelf {
			selfFound = true
			if link.Href != "/api/v1/orders/"+id.String() {
				t.Errorf("unexpected self href: %s", link.Href)
			}
		}
	}
	if !selfFound {
		t.Error("expected self link")
	}
}

func TestCollectionLinksFor(t *testing.T) {
	links := CollectionLinksFor("user")

	if len(links) != 2 {
		t.Errorf("expected 2 links, got %d", len(links))
	}
}

func TestCollectionLinksForWithBasePath(t *testing.T) {
	links := CollectionLinksFor("order", "/api")

	var selfFound bool
	for _, link := range links {
		if link.Rel == RelSelf {
			selfFound = true
			if link.Href != "/api/orders" {
				t.Errorf("unexpected self href: %s", link.Href)
			}
		}
	}
	if !selfFound {
		t.Error("expected self link")
	}
}

func TestChildLinksFor(t *testing.T) {
	parentID := uuid.New()
	childID := uuid.New()
	parent := testResource{id: parentID, typ: "order"}
	child := testResource{id: childID, typ: "item"}

	links := ChildLinksFor(parent, child)

	if len(links) != 5 {
		t.Errorf("expected 5 links, got %d", len(links))
	}

	var parentFound bool
	for _, link := range links {
		if link.Rel == RelParent {
			parentFound = true
		}
	}
	if !parentFound {
		t.Error("expected parent link")
	}
}

func TestLinkBuilder(t *testing.T) {
	id := uuid.New()
	obj := testResource{id: id, typ: "user"}

	builder := NewLinkBuilder()
	links := builder.
		AddRESTfulLinks(obj).
		Custom("related", "/related").
		Build()

	if len(links) < 5 {
		t.Errorf("expected at least 5 links, got %d", len(links))
	}
}

func TestLinkBuilderAddChildLinks(t *testing.T) {
	parentID := uuid.New()
	childID := uuid.New()
	parent := testResource{id: parentID, typ: "order"}
	child := testResource{id: childID, typ: "item"}

	builder := NewLinkBuilder()
	links := builder.AddChildLinks(parent, child).Build()

	if len(links) != 5 {
		t.Errorf("expected 5 links, got %d", len(links))
	}
}

func TestLinkBuilderAdd(t *testing.T) {
	builder := NewLinkBuilder()
	links := builder.
		Add(Link{Rel: "custom1", Href: "/c1"}, Link{Rel: "custom2", Href: "/c2"}).
		Build()

	if len(links) != 2 {
		t.Errorf("expected 2 links, got %d", len(links))
	}
}

func TestRespondWithLinks(t *testing.T) {
	id := uuid.New()
	obj := testResource{id: id, typ: "user"}

	rec := httptest.NewRecorder()
	RespondWithLinks(rec, obj)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRespondCollection(t *testing.T) {
	rec := httptest.NewRecorder()
	data := []string{"item1", "item2"}

	RespondCollection(rec, data, "item")

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRespondChild(t *testing.T) {
	parentID := uuid.New()
	childID := uuid.New()
	parent := testResource{id: parentID, typ: "order"}
	child := testResource{id: childID, typ: "item"}

	rec := httptest.NewRecorder()
	RespondChild(rec, parent, child)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
