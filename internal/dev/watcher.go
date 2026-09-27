package dev

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Event struct {
	Path string
	Op   fsnotify.Op
}

type Watcher struct {
	root     string
	debounce time.Duration
	watcher  *fsnotify.Watcher
	events   chan Event
}

func NewWatcher(root string) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		root:     root,
		debounce: 150 * time.Millisecond,
		watcher:  w,
		events:   make(chan Event, 32),
	}, nil
}

func (w *Watcher) Events() <-chan Event { return w.events }
func (w *Watcher) Close() error         { return w.watcher.Close() }

func (w *Watcher) Start(ctx context.Context) error {
	if err := w.addRecursive(w.root); err != nil {
		return err
	}

	go func() {
		var pending Event
		var timer *time.Timer
		var timerC <-chan time.Time

		emit := func() {
			if pending.Path == "" {
				return
			}
			select {
			case w.events <- pending:
			case <-ctx.Done():
			}
			pending = Event{}
		}

		for {
			select {
			case <-ctx.Done():
				if timer != nil {
					timer.Stop()
				}
				return

			case ev, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if w.ignore(ev.Name) {
					continue
				}
				// 新建目录需要递归加监听
				if ev.Op&fsnotify.Create != 0 {
					if info, err := os.Stat(ev.Name); err == nil && info.IsDir() {
						_ = w.addRecursive(ev.Name)
					}
				}
				pending = Event{Path: ev.Name, Op: ev.Op}
				if timer == nil {
					timer = time.NewTimer(w.debounce)
					timerC = timer.C
				} else {
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer.Reset(w.debounce)
				}

			case <-timerC:
				emit()
				timerC = nil

			case _, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				// 监听错误忽略，避免噪音
			}
		}
	}()

	return nil
}

func (w *Watcher) addRecursive(root string) error {
	return filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if ignoreDir(p) {
			return filepath.SkipDir
		}
		return w.watcher.Add(p)
	})
}

func (w *Watcher) ignore(path string) bool {
	base := filepath.Base(path)
	if strings.HasSuffix(base, "~") || strings.HasSuffix(base, ".swp") {
		return true
	}
	if strings.HasPrefix(base, ".") && base != "." {
		return true
	}
	for _, seg := range strings.Split(filepath.ToSlash(path), "/") {
		if ignoreDir(seg) {
			return true
		}
	}
	return false
}

func ignoreDir(name string) bool {
	base := filepath.Base(name)
	switch base {
	case ".git", "dist", "node_modules", ".cache", ".idea", ".vscode":
		return true
	}
	return false
}
