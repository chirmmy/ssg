package asset

import (
	"encoding/json"
	"os"
)

// Manifest 记录源资源名到最终 URL 的映射
// e.g. "style.css" -> "/assets/style.a1b2c3d4.css"
type Manifest struct {
	Entries map[string]string `json:"entries"`
}

func NewManifest() *Manifest {
	return &Manifest{Entries: make(map[string]string)}
}

func (m *Manifest) Set(name, url string) {
	if m.Entries == nil {
		m.Entries = make(map[string]string)
	}
	m.Entries[name] = url
}

// URL 返回资源的最终 URL。不存在时回退到 /assets/<name>
func (m *Manifest) URL(name string) string {
	if m == nil {
		return "/assets/" + name
	}
	if u, ok := m.Entries[name]; ok {
		return u
	}
	return "/assets/" + name
}

func (m *Manifest) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, m)
}

func (m *Manifest) Save(path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
