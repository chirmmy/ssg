package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Site     SiteConfig     `toml:"site"`
	Build    BuildConfig    `toml:"build"`
	Markdown MarkdownConfig `toml:"markdown"`
}

// SiteConfig represents the configuration for the static site generator.
type SiteConfig struct {
	Title       string `toml:"title"`
	Description string `toml:"description"`
	BaseURL     string `toml:"base_url"`
	Author      string `toml:"author"`
	Language    string `toml:"language"`
}

type BuildConfig struct {
	Workers int    `toml:"workers"`
	Minify  bool   `toml:"minify"`
	Drafts  bool   `toml:"drafts"`
	OutDir  string `toml:"out_dir"`
}

type MarkdownConfig struct {
	Highlight string `toml:"highlight"`
}

func (c *Config) applyDefaults() {
	if c.Build.OutDir == "" {
		c.Build.OutDir = "dist"
	}
	if c.Build.Workers == 0 {
		c.Build.Workers = 8
	}
	if c.Markdown.Highlight == "" {
		c.Markdown.Highlight = "catppuccin-mocha"
	}
	if c.Site.Language == "" {
		c.Site.Language = "zh-CN"
	}
}

func LoadConfig(cfgPath string) (*Config, error) {
	// Implement the logic to load the configuration from a file (e.g., config.toml)
	// For now, return a default configuration
	cfgBytes, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(cfgBytes, &cfg); err != nil {
		return nil, err
	}

	cfg.applyDefaults()

	return &cfg, nil
}
