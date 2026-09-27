package asset

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

type Pipeline struct {
	SrcDir string // 源目录，例如 assets/
	OutDir string // 输出目录，例如 dist/assets/
	Dev    bool   // dev 模式：不压缩、不哈希
	Minify bool   // 生产模式下是否压缩
	minify *minify.M
}

func NewPipeline(srcDir, outDir string, dev, minifyEnabled bool) *Pipeline {
	p := &Pipeline{
		SrcDir: srcDir,
		OutDir: outDir,
		Dev:    dev,
		Minify: minifyEnabled,
	}
	if minifyEnabled && !dev {
		p.minify = minify.New()
		p.minify.AddFunc("text/css", css.Minify)
		p.minify.AddFunc("application/javascript", js.Minify)
		p.minify.AddFunc("text/javascript", js.Minify)
	}
	return p
}

// Build 处理所有资源，返回 manifest
// dev 模式：直接返回空 manifest，源文件由 dev server 直接从 SrcDir 提供
func (p *Pipeline) Build(ctx context.Context) (*Manifest, error) {
	if p.Dev {
		return NewManifest(), nil
	}

	m := NewManifest()

	if _, err := os.Stat(p.SrcDir); os.IsNotExist(err) {
		return m, nil
	}

	if err := os.MkdirAll(p.OutDir, 0o755); err != nil {
		return nil, err
	}

	err := filepath.WalkDir(p.SrcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		rel, _ := filepath.Rel(p.SrcDir, path)
		rel = filepath.ToSlash(rel)

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", rel, err)
		}

		processed, err := p.process(rel, data)
		if err != nil {
			return fmt.Errorf("process %s: %w", rel, err)
		}

		outName := fingerprintName(rel, processed)
		outPath := filepath.Join(p.OutDir, filepath.FromSlash(outName))
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(outPath, processed, 0o644); err != nil {
			return err
		}

		m.Set(rel, "/assets/"+outName)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (p *Pipeline) process(rel string, data []byte) ([]byte, error) {
	if p.minify == nil {
		return data, nil
	}
	mimeType := mimeByExt(rel)
	if mimeType == "" {
		return data, nil
	}
	out, err := p.minify.Bytes(mimeType, data)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// fingerprintName 给文件名加内容哈希：style.css -> style.a1b2c3d4.css
// 保留子目录结构：css/main.css -> css/main.a1b2c3d4.css
func fingerprintName(name string, data []byte) string {
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:4])
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	return fmt.Sprintf("%s.%s%s", base, hash, ext)
}

func mimeByExt(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".css":
		return "text/css"
	case ".js", ".mjs":
		return "application/javascript"
	}
	return ""
}
