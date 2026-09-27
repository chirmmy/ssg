package dev

import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type changeKind int

const (
	kindOther changeKind = iota
	kindCSS
	kindContent
	kindTemplate
	kindConfig
)

type Reloader struct {
	server  *Server
	watcher *Watcher
	build   BuildFunc

	mu       sync.Mutex
	building bool
	pending  bool
}

func NewReloader(s *Server, w *Watcher, b BuildFunc) *Reloader {
	return &Reloader{server: s, watcher: w, build: b}
}

func (r *Reloader) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-r.watcher.Events():
			if !ok {
				return
			}
			r.handle(ctx, ev)
		}
	}
}

func (r *Reloader) handle(ctx context.Context, ev Event) {
	r.mu.Lock()
	if r.building {
		r.pending = true
		r.mu.Unlock()
		return
	}
	r.building = true
	r.mu.Unlock()

	defer func() {
		r.mu.Lock()
		r.building = false
		wasPending := r.pending
		r.pending = false
		r.mu.Unlock()
		if wasPending {
			go r.handle(ctx, ev)
		}
	}()

	kind := classify(ev.Path)
	slog.Info("change", "path", shortPath(ev.Path), "kind", kindName(kind))

	start := time.Now()
	err := r.build(ctx)
	elapsed := time.Since(start).Round(time.Millisecond)

	if err != nil {
		slog.Error("build failed", "err", err, "took", elapsed)
		r.server.SetError(err)
		r.server.Hub().Broadcast("reload")
		return
	}

	r.server.SetError(nil)
	slog.Info("build ok", "took", elapsed)

	switch kind {
	case kindCSS:
		r.server.Hub().Broadcast("css")
	default:
		r.server.Hub().Broadcast("reload")
	}
}

func classify(path string) changeKind {
	ext := strings.ToLower(filepath.Ext(path))
	base := filepath.Base(path)
	switch {
	case ext == ".css":
		return kindCSS
	case ext == ".md":
		return kindContent
	case ext == ".html":
		return kindTemplate
	case base == "site.toml":
		return kindConfig
	}
	return kindOther
}

func kindName(k changeKind) string {
	switch k {
	case kindCSS:
		return "css"
	case kindContent:
		return "content"
	case kindTemplate:
		return "template"
	case kindConfig:
		return "config"
	}
	return "other"
}

func shortPath(p string) string {
	if i := strings.LastIndex(filepath.ToSlash(p), "/"); i >= 0 {
		return filepath.ToSlash(p)[i+1:]
	}
	return p
}
