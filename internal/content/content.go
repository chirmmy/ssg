package content

import (
	"html/template"
	"strings"
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
	ReadingTime int // 分钟
	WordCount   int // 正文字符数（去掉所有空白）
	Cover       string
	Featured    bool
}

type Frontmatter struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	PubDate     string   `yaml:"pubDate"`
	Updated     string   `yaml:"updated"`
	Tags        []string `yaml:"tags"`
	Draft       bool     `yaml:"draft"`
	Cover       string   `yaml:"cover"`
	Featured    bool     `yaml:"featured"`
}

// UpdatedLater 判断这篇文章是否被更新过（loader 里 updated 缺省等于 pubDate）。
// 模板用它决定要不要渲染「更新于……」。
func (c *Content) UpdatedLater() bool {
	if c == nil || c.Updated.IsZero() || c.PubDate.IsZero() {
		return false
	}
	return c.Updated.After(c.PubDate)
}

// TagAttr 供卡片模板写 data-tags（列表页的客户端筛选用它匹配）。
func (c *Content) TagAttr() string {
	if c == nil {
		return ""
	}
	return strings.Join(c.Tags, "|")
}

// SearchKey 供卡片模板写 data-search：标题 + 摘要 + 标签 + slug，统一小写。
func (c *Content) SearchKey() string {
	if c == nil {
		return ""
	}
	parts := []string{c.Title, c.Description, strings.Join(c.Tags, " "), c.Slug}
	return strings.ToLower(strings.Join(parts, " "))
}
