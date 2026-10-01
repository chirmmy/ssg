package template

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chirmmy/ssg/internal/asset"
)

type Options struct {
	FS          fs.FS
	Assets      *asset.Manifest
	CriticalCSS string
	Dev         bool
}

type Engine struct {
	mu   sync.RWMutex
	tmpl *template.Template
}

func (e *Engine) load(opts Options) error {
	root := template.New("").Funcs(funcMap(opts))

	err := fs.WalkDir(opts.FS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, ".html") {
			return nil
		}
		data, err := fs.ReadFile(opts.FS, p)
		if err != nil {
			return err
		}
		name := strings.TrimSuffix(path.Base(p), ".html")
		_, err = root.New(name).Parse(string(data))
		return err
	})

	if err != nil {
		return err
	}

	e.mu.Lock()
	e.tmpl = root
	e.mu.Unlock()

	return nil
}

func (e *Engine) Render(w io.Writer, name string, data interface{}) error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.tmpl == nil {
		return fmt.Errorf("template engine not loaded")
	}
	return e.tmpl.ExecuteTemplate(w, name, data)
}

func NewEngine(opts Options) (*Engine, error) {
	e := &Engine{}
	if err := e.load(opts); err != nil {
		return nil, err
	}
	return e, nil
}

func funcMap(opts Options) template.FuncMap {
	return template.FuncMap{
		"dateFormat": func(t time.Time, layout string) string {
			if t.IsZero() {
				return ""
			}
			return t.Format(layout)
		},
		"year": func() int {
			return time.Now().Year()
		},
		"asset": func(name string) string {
			return opts.Assets.URL(name)
		},
		"dict": func(values ...any) map[string]any {
			m := make(map[string]any)
			for i := 0; i < len(values); i += 2 {
				if i+1 < len(values) {
					if k, ok := values[i].(string); ok {
						m[k] = values[i+1]
					}
				}
			}
			return m
		},
		"criticalCSS": func() template.HTML {
			if opts.CriticalCSS == "" {
				return ""
			}
			return template.HTML("<style>" + opts.CriticalCSS + "</style>")
		},
		"isDev": func() bool {
			return opts.Dev
		},
		// num 给字数加千分位：4266 -> 4,266
		"num": func(n int) string {
			s := strconv.Itoa(n)
			neg := strings.HasPrefix(s, "-")
			if neg {
				s = s[1:]
			}
			var b strings.Builder
			for i, r := range s {
				if i > 0 && (len(s)-i)%3 == 0 {
					b.WriteByte(',')
				}
				b.WriteRune(r)
			}
			if neg {
				return "-" + b.String()
			}
			return b.String()
		},
	}
}
