package template

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/aquamarinepk/aqm"
	"github.com/gertd/go-pluralize"
)

const (
	defaultBasePath  = "assets/templates"
	defaultSharedDir = "shared"
	defaultExtension = ".html"
)

// Manager loads HTML templates from an embedded filesystem and keeps them in-memory
// for fast lookups at runtime. It implements aqm.Startable so it can be wired into
// service boot sequences directly.
type Manager struct {
	fs         fs.FS
	log        aqm.Logger
	basePath   string
	sharedDir  string
	extension  string
	pluralizer *pluralize.Client

	mu        sync.RWMutex
	templates map[string]*template.Template
}

// Option configures a Manager instance.
type Option func(*Manager)

// NewManager returns a Manager configured to read templates from the provided
// filesystem. When no options are supplied it defaults to the Appetite layout
// of assets/templates with a shared/ folder and .html files.
func NewManager(assets fs.FS, opts ...Option) *Manager {
	mgr := &Manager{
		fs:         assets,
		log:        aqm.NewNoopLogger(),
		basePath:   defaultBasePath,
		sharedDir:  defaultSharedDir,
		extension:  defaultExtension,
		pluralizer: pluralize.NewClient(),
		templates:  make(map[string]*template.Template),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(mgr)
		}
	}
	return mgr
}

// WithLogger wires a custom logger. It falls back to a noop logger when nil.
func WithLogger(logger aqm.Logger) Option {
	return func(m *Manager) {
		if logger != nil {
			m.log = logger
		}
	}
}

// WithBasePath overrides the root folder where templates live.
func WithBasePath(base string) Option {
	return func(m *Manager) {
		if base != "" {
			m.basePath = strings.Trim(base, "/")
		}
	}
}

// WithSharedDir overrides the shared template directory name.
func WithSharedDir(name string) Option {
	return func(m *Manager) {
		if name != "" {
			m.sharedDir = strings.Trim(name, "/")
		}
	}
}

// WithExtension changes the file extension filter (defaults to .html).
func WithExtension(ext string) Option {
	return func(m *Manager) {
		if ext != "" {
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			m.extension = ext
		}
	}
}

// WithPluralizer allows callers to bring their own pluralization rules.
func WithPluralizer(client *pluralize.Client) Option {
	return func(m *Manager) {
		if client != nil {
			m.pluralizer = client
		}
	}
}

// Start loads all templates into memory. It satisfies aqm.Startable.
func (m *Manager) Start(context.Context) error {
	if err := m.parseTemplates(); err != nil {
		return fmt.Errorf("parse templates: %w", err)
	}
	m.log.Info("template manager ready", "count", len(m.templates))
	return nil
}

// Reload reparses all templates from disk/FS.
func (m *Manager) Reload() error {
	m.log.Info("Reloading templates")
	return m.parseTemplates()
}

// Get retrieves a parsed template by filename.
func (m *Manager) Get(name string) (*template.Template, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tmpl, ok := m.templates[name]
	if !ok {
		return nil, fmt.Errorf("template %s not found", name)
	}

	return tmpl, nil
}

// GetByPath resolves a template name using the handler/action convention
// used by the Appetite admin UI.
func (m *Manager) GetByPath(handler, action string) (*template.Template, error) {
	handler = strings.TrimSpace(handler)
	action = strings.TrimSpace(action)
	if handler == "" || action == "" {
		return nil, errors.New("handler and action required")
	}

	var name string
	switch action {
	case "list":
		name = m.pluralizer.Plural(handler) + m.extension
	case "new", "edit", "show":
		name = fmt.Sprintf("%s-%s%s", action, handler, m.extension)
	default:
		name = fmt.Sprintf("%s-%s%s", action, handler, m.extension)
	}

	return m.Get(name)
}

func (m *Manager) parseTemplates() error {
	if m.fs == nil {
		return errors.New("template filesystem not configured")
	}

	baseEntries, err := fs.ReadDir(m.fs, m.basePath)
	if err != nil {
		return fmt.Errorf("reading template base path %s: %w", m.basePath, err)
	}

	sharedEntries, err := m.readShared()
	if err != nil {
		return err
	}

	handlerDirs := collectHandlerDirs(baseEntries, m.sharedDir)
	allPaths, err := m.collectAllPaths(handlerDirs, sharedEntries)
	if err != nil {
		return err
	}
	if len(allPaths) == 0 {
		return errors.New("no templates found")
	}

	templates := make(map[string]*template.Template)
	if err := m.buildHandlerTemplates(handlerDirs, allPaths, templates); err != nil {
		return err
	}
	if err := m.buildSharedTemplates(sharedEntries, allPaths, templates); err != nil {
		return err
	}

	m.mu.Lock()
	m.templates = templates
	m.mu.Unlock()
	return nil
}

func (m *Manager) readShared() ([]fs.DirEntry, error) {
	sharedPath := path.Join(m.basePath, m.sharedDir)
	entries, err := fs.ReadDir(m.fs, sharedPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading shared templates: %w", err)
	}
	return entries, nil
}

func (m *Manager) collectAllPaths(handlerDirs []string, sharedEntries []fs.DirEntry) ([]string, error) {
	var paths []string

	for _, entry := range sharedEntries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), m.extension) {
			paths = append(paths, path.Join(m.basePath, m.sharedDir, entry.Name()))
		}
	}

	for _, handlerDir := range handlerDirs {
		entries, err := fs.ReadDir(m.fs, path.Join(m.basePath, handlerDir))
		if err != nil {
			m.log.Error("error reading handler templates", "handler", handlerDir, "error", err)
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), m.extension) {
				continue
			}
			paths = append(paths, path.Join(m.basePath, handlerDir, entry.Name()))
		}
	}

	// Ensure deterministic ordering so template parsing remains stable.
	sort.Strings(paths)
	return paths, nil
}

func (m *Manager) buildHandlerTemplates(handlerDirs []string, allPaths []string, templates map[string]*template.Template) error {
	for _, handlerDir := range handlerDirs {
		entries, err := fs.ReadDir(m.fs, path.Join(m.basePath, handlerDir))
		if err != nil {
			m.log.Error("error reading handler templates", "handler", handlerDir, "error", err)
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), m.extension) {
				continue
			}
			name := entry.Name()
			tmpl := template.New(name)
			parsed, err := tmpl.ParseFS(m.fs, allPaths...)
			if err != nil {
				return fmt.Errorf("parsing template %s: %w", name, err)
			}
			templates[name] = parsed
			m.log.Debug("loaded template", "name", name)
		}
	}
	return nil
}

func (m *Manager) buildSharedTemplates(sharedEntries []fs.DirEntry, allPaths []string, templates map[string]*template.Template) error {
	for _, entry := range sharedEntries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), m.extension) {
			continue
		}
		name := entry.Name()
		if strings.EqualFold(name, "base"+m.extension) {
			// base layout is included everywhere but not exposed via Get.
			continue
		}
		tmpl := template.New(name)
		parsed, err := tmpl.ParseFS(m.fs, allPaths...)
		if err != nil {
			return fmt.Errorf("parsing shared template %s: %w", name, err)
		}
		templates[name] = parsed
		m.log.Debug("loaded shared template", "name", name)
	}
	return nil
}

func collectHandlerDirs(entries []fs.DirEntry, sharedDir string) []string {
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != sharedDir {
			dirs = append(dirs, entry.Name())
		}
	}
	sort.Strings(dirs)
	return dirs
}
