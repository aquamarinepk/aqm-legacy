package fileserver

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/aquamarinepk/aqm"
	"github.com/go-chi/chi/v5"
)

const (
	defaultStaticDir = "assets/static"
	defaultURLPrefix = "/static"
	rootPrefix       = "/"
)

// Server mounts static assets from an embedded filesystem under a given URL prefix.
type Server struct {
	fs        fs.FS
	log       aqm.Logger
	dir       string
	urlPrefix string
}

// Option configures a static file server.
type Option func(*Server)

// New returns a Server using Appetite's default static folder and prefix.
func New(assets fs.FS, opts ...Option) *Server {
	srv := &Server{
		fs:        assets,
		log:       aqm.NewNoopLogger(),
		dir:       defaultStaticDir,
		urlPrefix: defaultURLPrefix,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(srv)
		}
	}
	return srv
}

// WithLogger wires a custom logger for route registration output.
func WithLogger(logger aqm.Logger) Option {
	return func(s *Server) {
		if logger != nil {
			s.log = logger
		}
	}
}

// WithDirectory overrides the static asset directory.
func WithDirectory(dir string) Option {
	return func(s *Server) {
		if dir != "" {
			s.dir = strings.Trim(dir, "/")
		}
	}
}

// WithURLPrefix overrides the HTTP mount point (defaults to /static).
func WithURLPrefix(prefix string) Option {
	return func(s *Server) {
		if prefix == "" {
			return
		}
		if !strings.HasPrefix(prefix, "/") {
			prefix = "/" + prefix
		}
		s.urlPrefix = strings.TrimRight(prefix, "/")
		if s.urlPrefix == "" {
			s.urlPrefix = rootPrefix
		}
	}
}

// RegisterRoutes implements aqm.HTTPModule.
func (s *Server) RegisterRoutes(r chi.Router) {
	if r == nil || s.fs == nil {
		return
	}

	staticFS, err := fs.Sub(s.fs, s.dir)
	if err != nil {
		s.log.Error("fileserver: cannot create sub filesystem", "dir", s.dir, "error", err)
		return
	}

	prefix := s.urlPrefix
	pattern := prefix + "/*"
	strip := prefix + "/"
	if prefix == rootPrefix {
		pattern = "/*"
		strip = rootPrefix
	}

	s.log.Info("Registering static file server", "prefix", prefix, "dir", path.Join(rootPrefix, s.dir))
	handler := http.StripPrefix(strip, http.FileServer(http.FS(staticFS)))
	r.Handle(pattern, handler)
}
