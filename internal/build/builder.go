package build

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/chirmmy/ssg/internal/config"
	"github.com/chirmmy/ssg/internal/content"
	"github.com/chirmmy/ssg/internal/render"
	"github.com/chirmmy/ssg/internal/site"
	"github.com/chirmmy/ssg/internal/template"
)

type Builder struct {
	Config *config.Config
	Root   string
}

func NewBuilder(cfg *config.Config, root string) *Builder {
	return &Builder{
		Config: cfg,
		Root:   root,
	}
}

func (b *Builder) Build(ctx context.Context) error {

	md := content.NewMarkdown(b.Config.Markdown.Highlight)
	loader := content.NewLoader(
		filepath.Join(b.Root, "content"),
		b.Config.Build.Workers,
		md,
	)

	items, err := loader.Load(ctx)
	if err != nil {
		return fmt.Errorf("load content: %w", err)
	}

	s := site.NewSite(b.Config, items)

	tmplDir := filepath.Join(b.Root, "templates")
	tmplFS := os.DirFS(tmplDir)
	enginge, err := template.NewEngine(tmplFS)
	if err != nil {
		return fmt.Errorf("load templates: %w", err)
	}

	outDir := filepath.Join(b.Root, b.Config.Build.OutDir)
	if err := os.RemoveAll(outDir); err != nil {
		return err
	}

	renderer := render.NewRender(enginge, outDir, b.Config.Build.Workers)
	routes := render.BuildRoutes(s)
	if err := renderer.RenderAll(ctx, routes); err != nil {
		return fmt.Errorf("render: %w", err)
	}

	if err := copyStatic(filepath.Join(b.Root, "static"), outDir); err != nil {
		return fmt.Errorf("static: %w", err)
	}

	return nil
}

func copyStatic(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

var _ = embed.FS{}
