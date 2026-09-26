package render

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/chirmmy/ssg/internal/template"
	"golang.org/x/sync/errgroup"
)

type Render struct {
	Templates *template.Engine
	OutDir    string
	Workers   int
}

func NewRender(tmpl *template.Engine, outDir string, workers int) *Render {
	return &Render{
		Templates: tmpl,
		OutDir:    outDir,
		Workers:   workers,
	}
}

func (e *Render) RenderAll(ctx context.Context, routes []Route) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(e.Workers)

	bufPool := sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	for _, r := range routes {
		r := r
		g.Go(func() error {
			buf := bufPool.Get().(*bytes.Buffer)
			buf.Reset()
			defer bufPool.Put(buf)

			if err := e.Templates.Render(buf, r.Template, r.Data); err != nil {
				return fmt.Errorf("render %s: %w", r.Pattern, err)
			}
			return e.write(r.Pattern, buf.Bytes())
		})
	}
	return g.Wait()
}

func (e *Render) write(pattern string, data []byte) error {
	rel := strings.TrimPrefix(pattern, "/")
	if rel == "" || strings.HasSuffix(pattern, "/") {
		rel = filepath.Join(rel, "index.html")
	}
	if !strings.HasSuffix(rel, ".html") {
		rel += ".html"
	}
	out := filepath.Join(e.OutDir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	return os.WriteFile(out, data, 0o644)
}
