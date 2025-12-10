package aqm

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsHTMX(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   bool
	}{
		{"true", "true", true},
		{"True", "True", true},
		{"TRUE", "TRUE", true},
		{"false", "false", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set(HXRequest, tt.header)
			}
			got := IsHTMX(req)
			if got != tt.want {
				t.Errorf("IsHTMX() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsBoosted(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   bool
	}{
		{"true", "true", true},
		{"false", "false", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set(HXBoosted, tt.header)
			}
			got := IsBoosted(req)
			if got != tt.want {
				t.Errorf("IsBoosted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsHistoryRestore(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   bool
	}{
		{"true", "true", true},
		{"false", "false", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set(HXHistoryState, tt.header)
			}
			got := IsHistoryRestore(req)
			if got != tt.want {
				t.Errorf("IsHistoryRestore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetHTMXHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HXTrigger, "btn-submit")
	req.Header.Set(HXTriggerName, "submit")
	req.Header.Set(HXTarget, "#content")
	req.Header.Set(HXCurrentURL, "http://example.com/page")
	req.Header.Set(HXPrompt, "user input")

	if GetHTMXTrigger(req) != "btn-submit" {
		t.Error("GetHTMXTrigger failed")
	}
	if GetHTMXTriggerName(req) != "submit" {
		t.Error("GetHTMXTriggerName failed")
	}
	if GetHTMXTarget(req) != "#content" {
		t.Error("GetHTMXTarget failed")
	}
	if GetHTMXCurrentURL(req) != "http://example.com/page" {
		t.Error("GetHTMXCurrentURL failed")
	}
	if GetHTMXPrompt(req) != "user input" {
		t.Error("GetHTMXPrompt failed")
	}
}

func TestSetHTMXRedirect(t *testing.T) {
	rec := httptest.NewRecorder()
	SetHTMXRedirect(rec, "/new-page")

	if rec.Header().Get(HXRedirect) != "/new-page" {
		t.Errorf("expected HX-Redirect header to be /new-page")
	}
}

func TestSetHTMXRefresh(t *testing.T) {
	rec := httptest.NewRecorder()
	SetHTMXRefresh(rec)

	if rec.Header().Get(HXRefresh) != "true" {
		t.Errorf("expected HX-Refresh header to be true")
	}
}

func TestSetHTMXPushURL(t *testing.T) {
	rec := httptest.NewRecorder()
	SetHTMXPushURL(rec, "/pushed")

	if rec.Header().Get(HXPushURL) != "/pushed" {
		t.Errorf("expected HX-Push-Url header")
	}
}

func TestSetHTMXReplaceURL(t *testing.T) {
	rec := httptest.NewRecorder()
	SetHTMXReplaceURL(rec, "/replaced")

	if rec.Header().Get(HXReplaceURL) != "/replaced" {
		t.Errorf("expected HX-Replace-Url header")
	}
}

func TestSetHTMXRetarget(t *testing.T) {
	rec := httptest.NewRecorder()
	SetHTMXRetarget(rec, "#new-target")

	if rec.Header().Get(HXRetarget) != "#new-target" {
		t.Errorf("expected HX-Retarget header")
	}
}

func TestSetHTMXReswap(t *testing.T) {
	rec := httptest.NewRecorder()
	SetHTMXReswap(rec, SwapOuterHTML)

	if rec.Header().Get(HXReswap) != "outerHTML" {
		t.Errorf("expected HX-Reswap header")
	}
}

func TestSetHTMXReselect(t *testing.T) {
	rec := httptest.NewRecorder()
	SetHTMXReselect(rec, ".selected")

	if rec.Header().Get(HXReselect) != ".selected" {
		t.Errorf("expected HX-Reselect header")
	}
}

func TestTriggerEvent(t *testing.T) {
	tests := []struct {
		name    string
		event   interface{}
		wantErr bool
	}{
		{
			name:    "stringEvent",
			event:   "myEvent",
			wantErr: false,
		},
		{
			name:    "mapEvent",
			event:   map[string]interface{}{"event": "data"},
			wantErr: false,
		},
		{
			name:    "structEvent",
			event:   struct{ Name string }{Name: "test"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			err := TriggerEvent(rec, tt.event)

			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if rec.Header().Get(HXTriggerResp) == "" {
				t.Error("expected HX-Trigger header")
			}
		})
	}
}

func TestRedirectOrHeader(t *testing.T) {
	tests := []struct {
		name     string
		isHTMX   bool
		wantCode int
	}{
		{
			name:     "htmxRequest",
			isHTMX:   true,
			wantCode: http.StatusOK,
		},
		{
			name:     "normalRequest",
			isHTMX:   false,
			wantCode: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.isHTMX {
				req.Header.Set(HXRequest, "true")
			}

			RedirectOrHeader(rec, req, "/target")

			if tt.isHTMX {
				if rec.Header().Get(HXRedirect) != "/target" {
					t.Error("expected HX-Redirect header for HTMX request")
				}
			} else {
				if rec.Code != tt.wantCode {
					t.Errorf("expected status %d, got %d", tt.wantCode, rec.Code)
				}
			}
		})
	}
}

func TestHTMXBuilder(t *testing.T) {
	h := H()
	if h == nil {
		t.Fatal("H() returned nil")
	}
}

func TestHTMXBuilderMethods(t *testing.T) {
	h := H().
		Get("/api/data").
		Post("/api/submit").
		Put("/api/update").
		Patch("/api/patch").
		Delete("/api/delete").
		Target("#container").
		Swap(SwapInnerHTML).
		Trigger("click").
		PushURL(true).
		Boost().
		Confirm("Are you sure?").
		Include("#form").
		Vals(`{"key":"value"}`).
		Select(".item").
		Indicator("#spinner")

	attrs := h.Attrs()

	tests := []string{
		`hx-get="/api/data"`,
		`hx-post="/api/submit"`,
		`hx-put="/api/update"`,
		`hx-patch="/api/patch"`,
		`hx-delete="/api/delete"`,
		`hx-target="#container"`,
		`hx-swap="innerHTML"`,
		`hx-trigger="click"`,
		`hx-push-url="true"`,
		`hx-boost="true"`,
		`hx-confirm="Are you sure?"`,
		`hx-include="#form"`,
		`hx-vals='{"key":"value"}'`,
		`hx-select=".item"`,
		`hx-indicator="#spinner"`,
	}

	for _, expected := range tests {
		if !strings.Contains(string(attrs), expected) {
			t.Errorf("expected attrs to contain %q, got %s", expected, attrs)
		}
	}
}

func TestHTMXBuilderTargetID(t *testing.T) {
	h := H().TargetID("myid")
	attrs := h.Attrs()

	if !strings.Contains(string(attrs), `hx-target="#myid"`) {
		t.Errorf("expected hx-target with # prefix, got %s", attrs)
	}
}

func TestHTMXBuilderString(t *testing.T) {
	h := H().Get("/test")
	str := h.String()

	if !strings.Contains(str, "hx-get") {
		t.Errorf("expected String() to contain hx-get, got %s", str)
	}
}

func TestHTMXBuilderEmptyAttrs(t *testing.T) {
	h := H()
	attrs := h.Attrs()

	if string(attrs) != "" {
		t.Errorf("expected empty attrs, got %s", attrs)
	}
}

func TestSwapModeConstants(t *testing.T) {
	tests := []struct {
		mode SwapMode
		want string
	}{
		{SwapInnerHTML, "innerHTML"},
		{SwapOuterHTML, "outerHTML"},
		{SwapBeforeBegin, "beforebegin"},
		{SwapAfterBegin, "afterbegin"},
		{SwapBeforeEnd, "beforeend"},
		{SwapAfterEnd, "afterend"},
		{SwapDelete, "delete"},
		{SwapNone, "none"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if string(tt.mode) != tt.want {
				t.Errorf("SwapMode = %s, want %s", tt.mode, tt.want)
			}
		})
	}
}

func TestHTMXRequestHeaderConstants(t *testing.T) {
	if HXRequest != "HX-Request" {
		t.Errorf("HXRequest = %s", HXRequest)
	}
	if HXTrigger != "HX-Trigger" {
		t.Errorf("HXTrigger = %s", HXTrigger)
	}
	if HXTarget != "HX-Target" {
		t.Errorf("HXTarget = %s", HXTarget)
	}
}

func TestHTMXResponseHeaderConstants(t *testing.T) {
	if HXRedirect != "HX-Redirect" {
		t.Errorf("HXRedirect = %s", HXRedirect)
	}
	if HXRefresh != "HX-Refresh" {
		t.Errorf("HXRefresh = %s", HXRefresh)
	}
	if HXPushURL != "HX-Push-Url" {
		t.Errorf("HXPushURL = %s", HXPushURL)
	}
}
