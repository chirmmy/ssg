package content

import (
	"html"
	"html/template"
	"net/url"
	"regexp"
	"strings"
)

// 正文里使用的两种短代码：
//
//	{{<externalLinkCard title="Starter Workflows" link="https://…" cover="auto">}}
//	{{<postLinkCard path="PicGo-Typora" cover="https://…">}}
//
// 它们在 Markdown 渲染阶段不会被展开（goldmark 会把 <externalLinkCard …> 当成
// 内联 HTML 标签原样保留），因此在渲染完成后用下面的正则把它们换成链接卡。
// 注意：这两条正则都接受 " 与 &quot; 两种引号形态 —— 后者在 goldmark 转义后可能出现。
var (
	blockShortcodeRe  = regexp.MustCompile(`(?s)<p>\s*\{\{<\s*([a-zA-Z]\w*)((?:\s+[a-zA-Z]\w*=(?:"[^"]*"|&quot;[^&]*&quot;))*)\s*>\}\}\s*</p>`)
	inlineShortcodeRe = regexp.MustCompile(`\{\{<\s*([a-zA-Z]\w*)((?:\s+[a-zA-Z]\w*=(?:"[^"]*"|&quot;[^&]*&quot;))*)\s*>\}\}`)
	attrRe            = regexp.MustCompile(`([a-zA-Z]\w*)=(?:"([^"]*)"|&quot;([^&]*)&quot;)`)
)

// PostLookup 由 slug 查一篇文章的标题，用于 postLinkCard。
type PostLookup func(slug string) (title string, ok bool)

const linkCardIcon = `<svg class="link-card__icon-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M10 13a5 5 0 0 0 7.5.5l3-3a5 5 0 0 0-7-7l-1.5 1.5"/><path d="M14 11a5 5 0 0 0-7.5-.5l-3 3a5 5 0 0 0 7 7L12 19"/></svg>`
const linkCardGo = `<svg class="link-card__go" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M5 12h14M12 5l7 7-7 7"/></svg>`

// ExpandShortcodes 把正文里的短代码替换成链接卡。
// 认不出来的短代码原样保留，避免把内容改坏。
func ExpandShortcodes(body template.HTML, lookup PostLookup) template.HTML {
	s := string(body)
	s = blockShortcodeRe.ReplaceAllStringFunc(s, func(m string) string {
		return renderShortcode(blockShortcodeRe, m, lookup)
	})
	s = inlineShortcodeRe.ReplaceAllStringFunc(s, func(m string) string {
		return renderShortcode(inlineShortcodeRe, m, lookup)
	})
	return template.HTML(s)
}

func renderShortcode(re *regexp.Regexp, match string, lookup PostLookup) string {
	g := re.FindStringSubmatch(match)
	if len(g) < 3 {
		return match
	}
	name := g[1]
	attrs := parseAttrs(g[2])

	switch name {
	case "externalLinkCard":
		href := strings.TrimSpace(attrs["link"])
		title := strings.TrimSpace(attrs["title"])
		if !safeHref(href) || title == "" {
			return match
		}
		return linkCard(href, title, hrefLabel(href), true)

	case "postLinkCard":
		slug := strings.TrimSpace(attrs["path"])
		if slug == "" || lookup == nil {
			return match
		}
		title, ok := lookup(slug)
		if !ok || title == "" {
			return match
		}
		href := "/blog/" + slug + "/"
		return linkCard(href, title, "站内文章 · "+slug, false)
	}
	return match
}

func linkCard(href, title, sub string, external bool) string {
	target := ""
	if external {
		target = ` target="_blank" rel="noopener"`
	}
	return `<a class="link-card" href="` + html.EscapeString(href) + `"` + target + `>` +
		`<span class="link-card__icon">` + linkCardIcon + `</span>` +
		`<span class="link-card__body">` +
		`<span class="link-card__title">` + html.EscapeString(title) + `</span>` +
		`<span class="link-card__url">` + html.EscapeString(sub) + `</span>` +
		`</span>` + linkCardGo + `</a>`
}

// hrefLabel 把 https://github.com/actions/starter-workflows 缩写为 github.com/actions/starter-workflows
func hrefLabel(href string) string {
	u, err := url.Parse(href)
	if err != nil || u.Host == "" {
		return href
	}
	s := u.Host + u.Path
	return strings.TrimSuffix(s, "/")
}

// safeHref 只允许 http(s) 与站内绝对路径，避免 jinja 式的伪协议注入。
func safeHref(href string) bool {
	if href == "" {
		return false
	}
	if strings.HasPrefix(href, "/") && !strings.HasPrefix(href, "//") {
		return true
	}
	return strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://")
}

func parseAttrs(raw string) map[string]string {
	out := map[string]string{}
	for _, m := range attrRe.FindAllStringSubmatch(raw, -1) {
		val := m[2]
		if val == "" && m[3] != "" {
			val = html.UnescapeString("&quot;" + m[3] + "&quot;")
		}
		out[m[1]] = val
	}
	return out
}
