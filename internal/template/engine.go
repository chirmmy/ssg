package template

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path"
	"strings"
	"sync"
	"time"
)

type Engine struct {
	mu   sync.RWMutex
	tmpl *template.Template
}

func (e *Engine) load(fsys fs.FS) error {
	root := template.New("").Funcs(funcMap())

	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, ".html") {
			return nil
		}
		data, err := fs.ReadFile(fsys, p)
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

func NewEngine(fysy fs.FS) (*Engine, error) {
	e := &Engine{}
	if err := e.load(fysy); err != nil {
		return nil, err
	}
	return e, nil
}

func funcMap() template.FuncMap {
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
	}
}
