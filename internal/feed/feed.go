// Package feed 生成 RSS 2.0 订阅源。
// 页脚、列表页页头与侧栏的「订阅」入口都指向 /rss.xml，此前没有任何生成逻辑。
package feed

import (
	"bytes"
	"encoding/xml"
	"html"
	"strings"
	"time"

	"github.com/chirmmy/ssg/internal/config"
	"github.com/chirmmy/ssg/internal/content"
)

type rss struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title         string    `xml:"title"`
	Link          string    `xml:"link"`
	Description   string    `xml:"description"`
	Language      string    `xml:"language,omitempty"`
	LastBuildDate string    `xml:"lastBuildDate,omitempty"`
	Items         []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description,omitempty"`
}

// Render 输出整份 RSS 文档。
func Render(cfg *config.Config, posts []*content.Content) ([]byte, error) {
	base := strings.TrimSuffix(cfg.Site.BaseURL, "/")
	abs := func(p string) string {
		if base == "" {
			return p
		}
		return base + p
	}

	ch := rssChannel{
		Title:       cfg.Site.Title,
		Link:        abs("/blog/"),
		Description: cfg.Site.Description,
		Language:    cfg.Site.Language,
	}

	if len(posts) > 0 && !posts[0].PubDate.IsZero() {
		ch.LastBuildDate = posts[0].PubDate.Format(time.RFC1123Z)
	}

	for _, p := range posts {
		link := abs("/blog/" + p.Slug + "/")
		item := rssItem{
			Title:       p.Title,
			Link:        link,
			GUID:        link,
			Description: p.Description,
		}
		if !p.PubDate.IsZero() {
			item.PubDate = p.PubDate.Format(time.RFC1123Z)
		}
		if item.Description == "" {
			// 没有摘要时退化为纯文本正文的开头
			item.Description = excerpt(p.Raw, 160)
		}
		ch.Items = append(ch.Items, item)
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(rss{Version: "2.0", Channel: ch}); err != nil {
		return nil, err
	}
	enc.Flush()
	buf.WriteString("\n")
	return buf.Bytes(), nil
}

// excerpt 从 Markdown 正文里摘一段纯文本：去掉代码块与标记符号。
func excerpt(raw string, limit int) string {
	s := raw
	if i := strings.Index(s, "```"); i >= 0 {
		s = s[:i]
	}
	replacer := strings.NewReplacer(
		"#", " ", "*", " ", "`", " ", ">", " ", "[", " ", "]", " ",
		"(", " ", ")", " ", "-", " ", "\n", " ", "\t", " ",
	)
	s = replacer.Replace(s)
	s = strings.Join(strings.Fields(s), " ")
	runes := []rune(s)
	if len(runes) > limit {
		s = string(runes[:limit]) + "…"
	}
	return html.EscapeString(s)
}
