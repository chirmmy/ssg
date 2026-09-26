package content

import (
	"html/template"
	"time"
)

type Kind string

const (
	KindPage Kind = "page"
	KindPost Kind = "post"
)

type Content struct {
	Kind        Kind
	Slug        string
	Title       string
	Description string
	PubDate     time.Time
	Updated     time.Time
	Tags        []string
	Draft       bool
	Body        template.HTML
	Raw         string
	SourcePath  string
	Hash        string
	Extra       map[string]any
}

type Frontmatter struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	PubDate     string   `yaml:"pubDate"`
	Updated     string   `yaml:"updated"`
	Tags        []string `yaml:"tags"`
	Draft       bool     `yaml:"draft"`
}
