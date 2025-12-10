package template

import (
	"context"
	"html/template"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/aquamarinepk/aqm"
	"github.com/gertd/go-pluralize"
)

func TestNewManager(t *testing.T) {
	assets := fstest.MapFS{}
	mgr := NewManager(assets)

	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}
	if mgr.fs == nil {
		t.Error("fs should not be nil")
	}
	if mgr.basePath != defaultBasePath {
		t.Errorf("basePath = %s, want %s", mgr.basePath, defaultBasePath)
	}
	if mgr.sharedDir != defaultSharedDir {
		t.Errorf("sharedDir = %s, want %s", mgr.sharedDir, defaultSharedDir)
	}
	if mgr.extension != defaultExtension {
		t.Errorf("extension = %s, want %s", mgr.extension, defaultExtension)
	}
	if mgr.templates == nil {
		t.Error("templates map should be initialized")
	}
}

func TestNewManagerWithOptions(t *testing.T) {
	assets := fstest.MapFS{}
	logger := aqm.NewNoopLogger()
	pluralizer := pluralize.NewClient()

	mgr := NewManager(assets,
		WithLogger(logger),
		WithBasePath("custom/templates"),
		WithSharedDir("common"),
		WithExtension(".tmpl"),
		WithPluralizer(pluralizer),
	)

	if mgr.basePath != "custom/templates" {
		t.Errorf("basePath = %s, want custom/templates", mgr.basePath)
	}
	if mgr.sharedDir != "common" {
		t.Errorf("sharedDir = %s, want common", mgr.sharedDir)
	}
	if mgr.extension != ".tmpl" {
		t.Errorf("extension = %s, want .tmpl", mgr.extension)
	}
}

func TestNewManagerWithNilOption(t *testing.T) {
	assets := fstest.MapFS{}

	// Should not panic
	mgr := NewManager(assets, nil)

	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}
}

func TestWithLogger(t *testing.T) {
	assets := fstest.MapFS{}
	logger := aqm.NewNoopLogger()

	mgr := NewManager(assets, WithLogger(logger))

	if mgr.log == nil {
		t.Error("log should not be nil")
	}
}

func TestWithLoggerNil(t *testing.T) {
	assets := fstest.MapFS{}

	mgr := NewManager(assets, WithLogger(nil))

	if mgr.log == nil {
		t.Error("log should not be nil (should use default)")
	}
}

func TestWithBasePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"normal", "custom/path", "custom/path"},
		{"with slashes", "/custom/path/", "custom/path"},
		{"empty", "", defaultBasePath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(fstest.MapFS{}, WithBasePath(tt.path))
			if mgr.basePath != tt.want {
				t.Errorf("basePath = %s, want %s", mgr.basePath, tt.want)
			}
		})
	}
}

func TestWithSharedDir(t *testing.T) {
	tests := []struct {
		name string
		dir  string
		want string
	}{
		{"normal", "common", "common"},
		{"with slashes", "/common/", "common"},
		{"empty", "", defaultSharedDir},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(fstest.MapFS{}, WithSharedDir(tt.dir))
			if mgr.sharedDir != tt.want {
				t.Errorf("sharedDir = %s, want %s", mgr.sharedDir, tt.want)
			}
		})
	}
}

func TestWithExtension(t *testing.T) {
	tests := []struct {
		name string
		ext  string
		want string
	}{
		{"with dot", ".tmpl", ".tmpl"},
		{"without dot", "tmpl", ".tmpl"},
		{"empty", "", defaultExtension},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(fstest.MapFS{}, WithExtension(tt.ext))
			if mgr.extension != tt.want {
				t.Errorf("extension = %s, want %s", mgr.extension, tt.want)
			}
		})
	}
}

func TestWithPluralizer(t *testing.T) {
	assets := fstest.MapFS{}
	pluralizer := pluralize.NewClient()

	mgr := NewManager(assets, WithPluralizer(pluralizer))

	if mgr.pluralizer == nil {
		t.Error("pluralizer should not be nil")
	}
}

func TestWithPluralizerNil(t *testing.T) {
	assets := fstest.MapFS{}

	mgr := NewManager(assets, WithPluralizer(nil))

	if mgr.pluralizer == nil {
		t.Error("pluralizer should not be nil (should use default)")
	}
}

func TestManagerStartNoFS(t *testing.T) {
	mgr := &Manager{
		fs:        nil,
		log:       aqm.NewNoopLogger(),
		templates: make(map[string]*template.Template),
	}

	err := mgr.Start(context.Background())

	if err == nil {
		t.Error("Start should return error when fs is nil")
	}
}

func TestManagerStartInvalidBasePath(t *testing.T) {
	assets := fstest.MapFS{
		"other/file.html": &fstest.MapFile{Data: []byte("content")},
	}
	mgr := NewManager(assets)

	err := mgr.Start(context.Background())

	if err == nil {
		t.Error("Start should return error for invalid base path")
	}
}

func TestManagerStartNoTemplates(t *testing.T) {
	assets := fstest.MapFS{
		"assets/templates/shared/base.html": &fstest.MapFile{Data: []byte("{{define \"base\"}}base{{end}}")},
	}
	mgr := NewManager(assets)

	err := mgr.Start(context.Background())

	// Start succeeds with only shared templates (allPaths is not empty)
	// The templates map will be empty since base.html is excluded from direct access
	if err != nil {
		t.Errorf("Start error: %v", err)
	}
}

func TestManagerStartEmptyFS(t *testing.T) {
	assets := fstest.MapFS{}
	mgr := NewManager(assets)

	err := mgr.Start(context.Background())

	if err == nil {
		t.Error("Start should return error for empty filesystem")
	}
}

func TestManagerStartSuccess(t *testing.T) {
	assets := fstest.MapFS{
		"assets/templates/shared/base.html":   &fstest.MapFile{Data: []byte("{{define \"base\"}}base{{end}}")},
		"assets/templates/shared/header.html": &fstest.MapFile{Data: []byte("{{define \"header\"}}header{{end}}")},
		"assets/templates/user/users.html":    &fstest.MapFile{Data: []byte("{{template \"base\" .}}")},
		"assets/templates/user/new-user.html": &fstest.MapFile{Data: []byte("{{template \"base\" .}}")},
	}
	mgr := NewManager(assets)

	err := mgr.Start(context.Background())

	if err != nil {
		t.Errorf("Start error: %v", err)
	}
}

func TestManagerReload(t *testing.T) {
	assets := fstest.MapFS{
		"assets/templates/shared/base.html":  &fstest.MapFile{Data: []byte("{{define \"base\"}}base{{end}}")},
		"assets/templates/user/users.html":   &fstest.MapFile{Data: []byte("{{template \"base\" .}}")},
	}
	mgr := NewManager(assets)

	err := mgr.Start(context.Background())
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	err = mgr.Reload()
	if err != nil {
		t.Errorf("Reload error: %v", err)
	}
}

func TestManagerGet(t *testing.T) {
	assets := fstest.MapFS{
		"assets/templates/shared/base.html": &fstest.MapFile{Data: []byte("{{define \"base\"}}base{{end}}")},
		"assets/templates/user/users.html":  &fstest.MapFile{Data: []byte("{{template \"base\" .}}")},
	}
	mgr := NewManager(assets)

	err := mgr.Start(context.Background())
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	tmpl, err := mgr.Get("users.html")
	if err != nil {
		t.Errorf("Get error: %v", err)
	}
	if tmpl == nil {
		t.Error("template should not be nil")
	}
}

func TestManagerGetNotFound(t *testing.T) {
	assets := fstest.MapFS{
		"assets/templates/shared/base.html": &fstest.MapFile{Data: []byte("{{define \"base\"}}base{{end}}")},
		"assets/templates/user/users.html":  &fstest.MapFile{Data: []byte("{{template \"base\" .}}")},
	}
	mgr := NewManager(assets)

	err := mgr.Start(context.Background())
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	_, err = mgr.Get("nonexistent.html")
	if err == nil {
		t.Error("Get should return error for nonexistent template")
	}
}

func TestManagerGetByPath(t *testing.T) {
	assets := fstest.MapFS{
		"assets/templates/shared/base.html":   &fstest.MapFile{Data: []byte("{{define \"base\"}}base{{end}}")},
		"assets/templates/user/users.html":    &fstest.MapFile{Data: []byte("{{template \"base\" .}}")},
		"assets/templates/user/new-user.html": &fstest.MapFile{Data: []byte("{{template \"base\" .}}")},
		"assets/templates/user/edit-user.html": &fstest.MapFile{Data: []byte("{{template \"base\" .}}")},
		"assets/templates/user/show-user.html": &fstest.MapFile{Data: []byte("{{template \"base\" .}}")},
	}
	mgr := NewManager(assets)

	err := mgr.Start(context.Background())
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	tests := []struct {
		handler  string
		action   string
		wantName string
	}{
		{"user", "list", "users.html"},
		{"user", "new", "new-user.html"},
		{"user", "edit", "edit-user.html"},
		{"user", "show", "show-user.html"},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			tmpl, err := mgr.GetByPath(tt.handler, tt.action)
			if err != nil {
				t.Errorf("GetByPath(%s, %s) error: %v", tt.handler, tt.action, err)
			}
			if tmpl == nil {
				t.Error("template should not be nil")
			}
		})
	}
}

func TestManagerGetByPathEmptyParams(t *testing.T) {
	assets := fstest.MapFS{}
	mgr := NewManager(assets)

	tests := []struct {
		name    string
		handler string
		action  string
	}{
		{"empty handler", "", "list"},
		{"empty action", "user", ""},
		{"both empty", "", ""},
		{"whitespace handler", "  ", "list"},
		{"whitespace action", "user", "  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mgr.GetByPath(tt.handler, tt.action)
			if err == nil {
				t.Error("GetByPath should return error for empty params")
			}
		})
	}
}

func TestManagerGetByPathCustomAction(t *testing.T) {
	assets := fstest.MapFS{
		"assets/templates/shared/base.html":      &fstest.MapFile{Data: []byte("{{define \"base\"}}base{{end}}")},
		"assets/templates/user/custom-user.html": &fstest.MapFile{Data: []byte("{{template \"base\" .}}")},
	}
	mgr := NewManager(assets)

	err := mgr.Start(context.Background())
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	tmpl, err := mgr.GetByPath("user", "custom")
	if err != nil {
		t.Errorf("GetByPath error: %v", err)
	}
	if tmpl == nil {
		t.Error("template should not be nil")
	}
}

func TestCollectHandlerDirs(t *testing.T) {
	entries := []fs.DirEntry{
		&fakeEntry{name: "shared", isDir: true},
		&fakeEntry{name: "user", isDir: true},
		&fakeEntry{name: "admin", isDir: true},
		&fakeEntry{name: "file.html", isDir: false},
	}

	dirs := collectHandlerDirs(entries, "shared")

	if len(dirs) != 2 {
		t.Errorf("len(dirs) = %d, want 2", len(dirs))
	}
	// Should be sorted
	if dirs[0] != "admin" {
		t.Errorf("dirs[0] = %s, want admin", dirs[0])
	}
	if dirs[1] != "user" {
		t.Errorf("dirs[1] = %s, want user", dirs[1])
	}
}

func TestManagerReadSharedNotExist(t *testing.T) {
	assets := fstest.MapFS{
		"assets/templates/user/users.html": &fstest.MapFile{Data: []byte("content")},
	}
	mgr := NewManager(assets)

	entries, err := mgr.readShared()
	if err != nil {
		t.Errorf("readShared error: %v", err)
	}
	if entries != nil {
		t.Error("entries should be nil when shared dir doesn't exist")
	}
}

func TestConstants(t *testing.T) {
	if defaultBasePath != "assets/templates" {
		t.Errorf("defaultBasePath = %s, want assets/templates", defaultBasePath)
	}
	if defaultSharedDir != "shared" {
		t.Errorf("defaultSharedDir = %s, want shared", defaultSharedDir)
	}
	if defaultExtension != ".html" {
		t.Errorf("defaultExtension = %s, want .html", defaultExtension)
	}
}

func TestOptionType(t *testing.T) {
	var opt Option = func(m *Manager) {
		m.basePath = "custom"
	}

	mgr := &Manager{basePath: "original"}
	opt(mgr)

	if mgr.basePath != "custom" {
		t.Errorf("basePath = %s, want custom", mgr.basePath)
	}
}

// Helper types for testing
type fakeEntry struct {
	name  string
	isDir bool
}

func (e *fakeEntry) Name() string               { return e.name }
func (e *fakeEntry) IsDir() bool                { return e.isDir }
func (e *fakeEntry) Type() fs.FileMode          { return 0 }
func (e *fakeEntry) Info() (fs.FileInfo, error) { return nil, nil }
