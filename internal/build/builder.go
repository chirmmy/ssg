package build

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/chirmmy/ssg/internal/asset"
	"github.com/chirmmy/ssg/internal/config"
	"github.com/chirmmy/ssg/internal/content"
	"github.com/chirmmy/ssg/internal/render"
	"github.com/chirmmy/ssg/internal/site"
	"github.com/chirmmy/ssg/internal/template"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
)

type Builder struct {
	Config *config.Config
	Root   string
	Dev    bool
}

func NewBuilder(cfg *config.Config, root string, dev bool) *Builder {
	return &Builder{
		Config: cfg,
		Root:   root,
		Dev:    dev,
	}
}

func (b *Builder) Build(ctx context.Context) error {
	outDir := filepath.Join(b.Root, b.Config.Build.OutDir)
	if err := os.RemoveAll(outDir); err != nil {
		return err
	}

	// 1. 资源管线
	pipeline := asset.NewPipeline(
		filepath.Join(b.Root, "assets"),
		filepath.Join(outDir, "assets"),
		b.Dev,
		b.Config.Build.Minify,
	)
	manifest, err := pipeline.Build(ctx)
	if err != nil {
		return fmt.Errorf("assets: %w", err)
	}

	// 生产模式下把 manifest 写到 dist/manifest.json
	if !b.Dev {
		if err := manifest.Save(filepath.Join(outDir, "manifest.json")); err != nil {
			return fmt.Errorf("manifest: %w", err)
		}
	}

	// 2. 内容管线
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

	// 3. 读取关键 CSS
	criticalCSS, err := loadCriticalCSS(b.Root, b.Config.Build.Minify, b.Dev)
	if err != nil {
		return fmt.Errorf("critical css: %w", err)
	}

	// 4. 模板引擎（注入 manifest）
	tmplDir := filepath.Join(b.Root, "templates")
	tmplFS := os.DirFS(tmplDir)
	enginge, err := template.NewEngine(template.Options{
		FS:          tmplFS,
		Assets:      manifest,
		CriticalCSS: criticalCSS,
		Dev:         b.Dev,
	})
	if err != nil {
		return fmt.Errorf("load templates: %w", err)
	}

	// 4. 渲染
	renderer := render.NewRender(enginge, outDir, b.Config.Build.Workers)
	routes := render.BuildRoutes(s)
	if err := renderer.RenderAll(ctx, routes); err != nil {
		return fmt.Errorf("render: %w", err)
	}

	// 5. 静态资源
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

// loadCriticalCSS 读取 assets/critical.css。
// dev 模式不返回（避免影响热替换），prod 模式返回压缩后的内容。
func loadCriticalCSS(root string, minifyEnabled, dev bool) (string, error) {
	if dev {
		return "", nil
	}
	path := filepath.Join(root, "assets", "critical.css")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	if minifyEnabled {
		m := minify.New()
		m.AddFunc("text/css", css.Minify)
		out, err := m.Bytes("text/css", data)
		if err == nil {
			return string(out), nil
		}
	}
	return string(data), nil
}

var _ = embed.FS{}
