package dev

import (
	"context"
	"fmt"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BuildFunc 是一次完整构建的抽象，便于 dev 与 build 解耦
type BuildFunc func(ctx context.Context) error

// Server 是开发服务器，负责托管产物、SSE、脚本注入
type Server struct {
	root      string
	outDir    string
	assetsDir string // 源 assets 目录，dev 下直接从这里提供
	addr      string
	autoOpen  bool
	build     BuildFunc
	hub       *Hub
	srv       *http.Server
	lastErr   error // lastErr 保存最近一次构建错误，用于在页面请求时展示
}

func NewServer(root, outDir, assetsDir, addr string, build BuildFunc, autoOpen bool) *Server {
	return &Server{
		root:      root,
		outDir:    outDir,
		assetsDir: assetsDir,
		addr:      addr,
		autoOpen:  autoOpen,
		build:     build,
		hub:       NewHub(),
	}
}

func (s *Server) Hub() *Hub {
	return s.hub
}

// SetError 由 Reloader 调用，记录最近一次构建结果
func (s *Server) SetError(err error) {
	s.lastErr = err
}
func (s *Server) LastError() error {
	return s.lastErr
}

func (s *Server) Start(ctx context.Context) error {
	// HTTP 请求多路复用器
	mux := http.NewServeMux()
	mux.HandleFunc("/_dev/sse", s.handleSSE)
	mux.HandleFunc("/_dev/client.js", s.handleClientJS)
	mux.HandleFunc("/", s.handleStatic) // 匹配所有路径（兜底）

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.addr, err)
	}

	s.srv = &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0, // SSE 需要长连接
	}

	url := "http://" + ln.Addr().String() + "/"
	fmt.Printf("\n  ssg dev\n")
	fmt.Printf("  ➜  Local:   %s\n", url)
	if ip := lanIp(); ip != "" {
		fmt.Printf("  ➜  Network: http://%s/\n", net.JoinHostPort(ip, portOf(ln.Addr())))
	}
	fmt.Printf("  ➜  watching: content/, templates/, static/, site.toml\n\n")

	if s.autoOpen {
		go func() {
			time.Sleep(200 * time.Millisecond)
			OpenBrowser(url)
		}()
	}

	// 优雅关闭
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.srv.Shutdown(shutdownCtx)
	}()

	err = s.srv.Serve(ln)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	p := r.URL.Path
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}

	// dev 模式：直接从源 assets/ 目录提供，跳过 dist 缓存
	if strings.HasPrefix(p, "/assets/") && s.assetsDir != "" {
		if s.serveFromAssets(w, r, p) {
			return
		}
	}

	// 目录重定向：/about -> /about/
	if p != "/" && !strings.HasSuffix(p, "/") && !strings.Contains(filepath.Base(p), ".") {
		full := filepath.Join(s.outDir, filepath.FromSlash(strings.TrimPrefix(p, "/")))
		if info, err := os.Stat(full); err == nil && info.IsDir() {
			http.Redirect(w, r, p+"/", http.StatusMovedPermanently)
			return
		}
	}

	rel := resolvePath(p)
	full := filepath.Join(s.outDir, filepath.FromSlash(rel))

	if _, err := os.Stat(full); err != nil {
		// 尝试目录回退
		fallback := filepath.Join(s.outDir, filepath.FromSlash(rel), "index.html")
		if _, err2 := os.Stat(fallback); err2 == nil {
			full = fallback
		} else {
			s.serveNotFound(w, r)
			return
		}
	}

	// 有构建错误时，如果请求的是 HTML，优先展示错误页
	if s.lastErr != nil && strings.HasSuffix(full, ".html") {
		s.serveErrorPage(w, s.lastErr)
		return
	}

	s.serveFile(w, r, full)
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch := s.hub.Subscribe()
	defer s.hub.Unsubscribe(ch)

	// 首帧：告知客户端已连接
	_, _ = fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	ping := time.NewTicker(20 * time.Second)
	defer ping.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			_, _ = fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-ping.C:
			_, _ = fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		}
	}

}

func (s *Server) handleClientJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(DevClientJS))
}

func (s *Server) serveErrorPage(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write(DefaultErrorPage(err))
}

func (s *Server) serveNotFound(w http.ResponseWriter, r *http.Request) {
	p404 := filepath.Join(s.outDir, "404.html")
	if data, err := os.ReadFile(p404); err == nil {
		data = InjectDevScript(data)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(data)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) serveFile(w http.ResponseWriter, r *http.Request, full string) {
	data, err := os.ReadFile(full)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctype := contentType(full)
	if strings.HasPrefix(ctype, "text/html") {
		data = InjectDevScript(data)
	}

	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(data)
}

// serveFromAssets 从源 assets/ 目录提供文件
// 返回 true 表示已处理（成功或 404），false 表示继续走 dist 逻辑
func (s *Server) serveFromAssets(w http.ResponseWriter, r *http.Request, p string) bool {
	rel := strings.TrimPrefix(p, "/assets/")
	if rel == "" || strings.Contains(rel, "..") {
		return false
	}
	full := filepath.Join(s.assetsDir, filepath.FromSlash(rel))
	if info, err := os.Stat(full); err != nil || info.IsDir() {
		return false
	}
	data, err := os.ReadFile(full)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	w.Header().Set("Content-Type", contentType(full))
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(data)
	return true
}

func resolvePath(p string) string {
	rel := strings.TrimPrefix(p, "/")
	if rel == "" || strings.HasSuffix(rel, "/") {
		return filepath.Join(rel, "index.html")
	}
	return rel
}

func contentType(p string) string {
	ext := strings.ToLower(filepath.Ext(p))
	switch ext {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js", ".mjs":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".xml":
		return "application/xml; charset=utf-8"
	case ".txt":
		return "text/plain; charset=utf-8"
	}
	if ct := mime.TypeByExtension(ext); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

// lanIP 返回本机第一个非回环 IPv4 地址，用于打印 Network 地址
func lanIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				return ip4.String()
			}
		}
	}
	return ""
}

func portOf(addr net.Addr) string {
	if tcp, ok := addr.(*net.TCPAddr); ok {
		return fmt.Sprintf("%d", tcp.Port)
	}
	_, port, _ := net.SplitHostPort(addr.String())
	return port
}
