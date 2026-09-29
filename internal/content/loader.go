package content

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

type Loader struct {
	Root     string
	Workers  int
	Markdown *Markdown
}

func NewLoader(root string, workers int, md *Markdown) *Loader {
	return &Loader{
		Root:     root,
		Workers:  workers,
		Markdown: md,
	}
}

func (l *Loader) Load(ctx context.Context) ([]*Content, error) {

	var files []struct {
		path string
		kind Kind
	}

	const blogdir string = "blog" + string(filepath.Separator)
	err := filepath.WalkDir(l.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		kind := KindPage
		rel, _ := filepath.Rel(l.Root, path)
		if strings.HasPrefix(rel, blogdir) { // md file
			kind = KindPost
		}
		files = append(files, struct {
			path string
			kind Kind
		}{path, kind})

		return nil
	})

	if err != nil {
		return nil, err
	}

	results := make([]*Content, len(files))
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(l.Workers)

	for index, file := range files {
		index, file := index, file
		g.Go(func() error {
			content, err := l.parse(ctx, file.path, file.kind)
			if err != nil {
				return fmt.Errorf("%s: %w", file.path, err)
			}
			results[index] = content
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].PubDate.After(results[j].PubDate)
	})

	return results, nil
}

func (l *Loader) parse(_ context.Context, path string, kind Kind) (*Content, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	frontmatter, body, err := splitFrontmatter(raw)
	if err != nil {
		return nil, err
	}

	var meta Frontmatter
	if len(frontmatter) > 0 {
		if err := yaml.Unmarshal(frontmatter, &meta); err != nil {
			return nil, fmt.Errorf("Frontmatter: %w", err)
		}
	}

	htmlBody, err := l.Markdown.Render(body)
	if err != nil {
		return nil, err
	}

	pubDate := parseDate(meta.PubDate)
	updated := parseDate(meta.Updated)
	if updated.IsZero() {
		updated = pubDate
	}

	sum := sha256.Sum256(raw)
	readingTime := estimateReadingTime(string(body))
	return &Content{
		Kind:        kind,
		Slug:        slugFromPath(l.Root, path, kind),
		Title:       meta.Title,
		Description: meta.Description,
		PubDate:     pubDate,
		Updated:     updated,
		Tags:        meta.Tags,
		Draft:       meta.Draft,
		Body:        template.HTML(htmlBody),
		Raw:         string(body),
		SourcePath:  path,
		Hash:        hex.EncodeToString(sum[:8]),
		ReadingTime: readingTime,
	}, nil
}

func splitFrontmatter(raw []byte) (frontmatter, body []byte, err error) {
	// 归一化 Windows CRLF 换行：否则 "---\r\n" 无法匹配 "---\n"，
	// 整个 front matter 会被当成正文渲染
	rawstr := strings.ReplaceAll(string(raw), "\r\n", "\n")
	if !strings.HasPrefix(rawstr, "---\n") {
		return nil, raw, nil
	}
	end := strings.Index(rawstr[4:], "\n---")
	if end < 0 {
		return nil, nil, fmt.Errorf("unclosed frontmatter")
	}
	frontmatter = []byte(rawstr[4 : 4+end])
	rest := rawstr[4+end+4:]
	rest = strings.TrimPrefix(rest, "\n")
	return frontmatter, []byte(rest), nil
}

func parseDate(date string) time.Time {
	if date == "" { // default now
		return time.Time{}
	}
	for _, layout := range []string{"2006-01-02", time.RFC3339} {
		if t, err := time.Parse(layout, date); err == nil {
			return t
		}
	}
	return time.Time{}
}

func slugFromPath(root, path string, kind Kind) string {
	rel, _ := filepath.Rel(root, path)
	rel = strings.TrimSuffix(rel, ".md")
	if kind == KindPage {
		rel = strings.TrimPrefix(rel, "pages"+string(filepath.Separator))
		rel = strings.TrimPrefix(rel, "pages/")
	} else {
		rel = strings.TrimPrefix(rel, "blog"+string(filepath.Separator))
		rel = strings.TrimPrefix(rel, "blog/")
	}
	return filepath.ToSlash(rel)
}

// estimateReadingTime 按中文 400 字/分、英文 200 词/分估算。
func estimateReadingTime(text string) int {
	runes := 0
	for _, r := range text {
		if r > 0x2E80 { // 中日韩
			runes++
		}
	}
	words := len(strings.Fields(text))
	minutes := float64(runes)/400 + float64(words)/200
	if minutes < 1 {
		return 1
	}
	return int(minutes + 0.5)
}
