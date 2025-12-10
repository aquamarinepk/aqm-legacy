package fileserver

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/aquamarinepk/aqm"
	"github.com/go-chi/chi/v5"
)

func TestNew(t *testing.T) {
	assets := fstest.MapFS{
		"assets/static/test.js": &fstest.MapFile{Data: []byte("content")},
	}

	srv := New(assets)

	if srv == nil {
		t.Fatal("New returned nil")
	}
	if srv.fs == nil {
		t.Error("fs should not be nil")
	}
	if srv.dir != defaultStaticDir {
		t.Errorf("dir = %s, want %s", srv.dir, defaultStaticDir)
	}
	if srv.urlPrefix != defaultURLPrefix {
		t.Errorf("urlPrefix = %s, want %s", srv.urlPrefix, defaultURLPrefix)
	}
}

func TestNewWithOptions(t *testing.T) {
	assets := fstest.MapFS{}
	logger := aqm.NewNoopLogger()

	srv := New(assets,
		WithLogger(logger),
		WithDirectory("custom/static"),
		WithURLPrefix("/assets"),
	)

	if srv.dir != "custom/static" {
		t.Errorf("dir = %s, want custom/static", srv.dir)
	}
	if srv.urlPrefix != "/assets" {
		t.Errorf("urlPrefix = %s, want /assets", srv.urlPrefix)
	}
}

func TestWithLogger(t *testing.T) {
	assets := fstest.MapFS{}
	logger := aqm.NewNoopLogger()

	srv := New(assets, WithLogger(logger))

	if srv.log == nil {
		t.Error("log should not be nil")
	}
}

func TestWithLoggerNil(t *testing.T) {
	assets := fstest.MapFS{}

	srv := New(assets, WithLogger(nil))

	if srv.log == nil {
		t.Error("log should not be nil (should use default)")
	}
}

func TestWithDirectory(t *testing.T) {
	tests := []struct {
		name string
		dir  string
		want string
	}{
		{"normal", "custom/dir", "custom/dir"},
		{"with slashes", "/custom/dir/", "custom/dir"},
		{"empty", "", defaultStaticDir},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := New(fstest.MapFS{}, WithDirectory(tt.dir))
			if srv.dir != tt.want {
				t.Errorf("dir = %s, want %s", srv.dir, tt.want)
			}
		})
	}
}

func TestWithURLPrefix(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		want   string
	}{
		{"normal", "/assets", "/assets"},
		{"without slash", "assets", "/assets"},
		{"with trailing slash", "/assets/", "/assets"},
		{"empty", "", defaultURLPrefix},
		{"root only slash", "/", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := New(fstest.MapFS{}, WithURLPrefix(tt.prefix))
			if srv.urlPrefix != tt.want {
				t.Errorf("urlPrefix = %s, want %s", srv.urlPrefix, tt.want)
			}
		})
	}
}

func TestWithNilOption(t *testing.T) {
	assets := fstest.MapFS{}

	// Should not panic
	srv := New(assets, nil)

	if srv == nil {
		t.Error("New should return server even with nil option")
	}
}

func TestRegisterRoutesNilRouter(t *testing.T) {
	assets := fstest.MapFS{
		"assets/static/test.js": &fstest.MapFile{Data: []byte("content")},
	}
	srv := New(assets)

	// Should not panic
	srv.RegisterRoutes(nil)
}

func TestRegisterRoutesNilFS(t *testing.T) {
	srv := &Server{
		fs:        nil,
		log:       aqm.NewNoopLogger(),
		dir:       defaultStaticDir,
		urlPrefix: defaultURLPrefix,
	}
	r := chi.NewRouter()

	// Should not panic
	srv.RegisterRoutes(r)
}

func TestRegisterRoutesInvalidSubDir(t *testing.T) {
	assets := fstest.MapFS{
		"other/test.js": &fstest.MapFile{Data: []byte("content")},
	}
	srv := New(assets, WithDirectory("nonexistent"))
	r := chi.NewRouter()

	// Should not panic, just log error
	srv.RegisterRoutes(r)
}

func TestRegisterRoutesServesFiles(t *testing.T) {
	assets := fstest.MapFS{
		"assets/static/test.js":  &fstest.MapFile{Data: []byte("console.log('test');")},
		"assets/static/style.css": &fstest.MapFile{Data: []byte("body { color: red; }")},
	}
	srv := New(assets)
	r := chi.NewRouter()
	srv.RegisterRoutes(r)

	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
	}{
		{"/static/test.js", http.StatusOK, "console.log('test');"},
		{"/static/style.css", http.StatusOK, "body { color: red; }"},
		{"/static/notfound.js", http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("Body = %s, want %s", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestRegisterRoutesRootPrefix(t *testing.T) {
	assets := fstest.MapFS{
		"assets/static/test.js": &fstest.MapFile{Data: []byte("content")},
	}
	srv := New(assets, WithURLPrefix("/"))
	r := chi.NewRouter()
	srv.RegisterRoutes(r)

	req := httptest.NewRequest("GET", "/test.js", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRegisterRoutesCustomPrefix(t *testing.T) {
	assets := fstest.MapFS{
		"public/files/doc.txt": &fstest.MapFile{Data: []byte("document")},
	}
	srv := New(assets,
		WithDirectory("public/files"),
		WithURLPrefix("/files"),
	)
	r := chi.NewRouter()
	srv.RegisterRoutes(r)

	req := httptest.NewRequest("GET", "/files/doc.txt", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "document" {
		t.Errorf("Body = %s, want document", rec.Body.String())
	}
}

func TestServerImplementsHTTPModule(t *testing.T) {
	assets := fstest.MapFS{}
	srv := New(assets)

	// Verify Server has RegisterRoutes method matching HTTPModule interface
	r := chi.NewRouter()
	srv.RegisterRoutes(r)
}

func TestOptionType(t *testing.T) {
	var opt Option = func(s *Server) {
		s.dir = "custom"
	}

	srv := &Server{dir: "original"}
	opt(srv)

	if srv.dir != "custom" {
		t.Errorf("dir = %s, want custom", srv.dir)
	}
}

func TestConstants(t *testing.T) {
	if defaultStaticDir != "assets/static" {
		t.Errorf("defaultStaticDir = %s, want assets/static", defaultStaticDir)
	}
	if defaultURLPrefix != "/static" {
		t.Errorf("defaultURLPrefix = %s, want /static", defaultURLPrefix)
	}
	if rootPrefix != "/" {
		t.Errorf("rootPrefix = %s, want /", rootPrefix)
	}
}

// Test with real fs.FS interface
type mockFS struct {
	err error
}

func (m mockFS) Open(name string) (fs.File, error) {
	if m.err != nil {
		return nil, m.err
	}
	return nil, fs.ErrNotExist
}

func TestRegisterRoutesWithMockFS(t *testing.T) {
	srv := &Server{
		fs:        mockFS{err: fs.ErrNotExist},
		log:       aqm.NewNoopLogger(),
		dir:       "nonexistent",
		urlPrefix: "/static",
	}
	r := chi.NewRouter()

	// Should not panic
	srv.RegisterRoutes(r)
}
