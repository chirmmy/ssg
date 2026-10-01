package content

import (
	"bytes"
	"strings"
	"sync"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark-emoji"
	emjidef "github.com/yuin/goldmark-emoji/definition"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// codeBlockSVG 图标常量，构建期直接内联到代码块头部。
const (
	codeBlockChevron = `<path d="M6 9l6 6 6-6"/>`
	codeBlockCopy    = `<rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15V5a2 2 0 0 1 2-2h10"/>`
	codeBlockSVGAttr = ` fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"`
)

// sanitizeLang 保留语言标识中的安全字符，防止注入。
func sanitizeLang(lang string) string {
	var b strings.Builder
	for _, r := range lang {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '+', r == '-', r == '#':
			b.WriteRune(r)
		}
	}
	s := b.String()
	if s == "" {
		return "code"
	}
	return strings.ToLower(s)
}

// renderCodeBlockHeader 输出 macOS 风格的代码块头部：三色圆点 + 语言标签 + 折叠/复制按钮。
func renderCodeBlockHeader(w util.BufWriter, lang string) {
	safe := sanitizeLang(lang)
	w.WriteString(`<div class="code-block" data-lang="` + safe + `">`)
	w.WriteString(`<div class="code-block__header">`)
	w.WriteString(`<span class="code-block__dots" aria-hidden="true"><i></i><i></i><i></i></span>`)
	w.WriteString(`<span class="code-block__lang">` + strings.ToUpper(safe) + `</span>`)
	w.WriteString(`<div class="code-block__actions">`)
	w.WriteString(`<button type="button" class="code-block__collapse" aria-label="折叠代码" aria-expanded="true">`)
	w.WriteString(`<svg viewBox="0 0 24 24"` + codeBlockSVGAttr + `>` + codeBlockChevron + `</svg></button>`)
	w.WriteString(`<button type="button" class="code-block__copy" aria-label="复制代码">`)
	w.WriteString(`<svg viewBox="0 0 24 24"` + codeBlockSVGAttr + `>` + codeBlockCopy + `</svg></button>`)
	w.WriteString(`</div>`)
	w.WriteString(`</div>`)
	w.WriteString(`<div class="code-block__body">`)
}

// Markdown processor.
type Markdown struct {
	md      goldmark.Markdown
	bufPool sync.Pool
}

func NewMarkdown(highlightStyle string) *Markdown {
	if _, ok := styles.Registry[highlightStyle]; !ok {
		highlightStyle = "github" // Default style if the specified one is not found
	}
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,         // GitHub Flavored Markdown
			extension.Footnote,    // Footnotes
			extension.Typographer, // Typographic replacements
			emoji.New(emoji.WithEmojis(emjidef.Github())), // :shortcode: 形式的 emoji（如 :one: :tada:）
			highlighting.NewHighlighting(
				highlighting.WithStyle(highlightStyle),
				// 用 class 而非内联样式渲染，配合 chroma.css 的明暗双主题
				// （内联样式会锁死浅色配色，无法跟随站点主题切换）。
				highlighting.WithFormatOptions(chromahtml.WithClasses(true)),
				highlighting.WithWrapperRenderer(func(w util.BufWriter, ctx highlighting.CodeBlockContext, entering bool) {
					if entering {
						lang := "code"
						if l, ok := ctx.Language(); ok && len(l) > 0 {
							lang = string(l)
						}
						renderCodeBlockHeader(w, lang)
					} else {
						w.WriteString(`</div></div>`)
					}
				}),
			), // Use the specified style for syntax highlighting

		),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(), // Automatically generate heading IDs
		// 独占一段的图片 → <figure class="figure"> + figcaption 图注
		parser.WithASTTransformers(
			util.Prioritized(&figureTransformer{}, 100),
			util.Prioritized(&mermaidTransformer{}, 90), // ```mermaid → 前端渲染占位块
		),
	),
	goldmark.WithRendererOptions(
		html.WithUnsafe(), // Allow raw HTML in Markdown
		renderer.WithNodeRenderers(util.Prioritized(&figureHTMLRenderer{}, 100)),
	),
	)

	return &Markdown{
		md: md,
		bufPool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer) // Create a new buffer when needed
			},
		},
	}
}

// Render converts Markdown input to HTML output.
func (m *Markdown) Render(input []byte) ([]byte, error) {
	buf := m.bufPool.Get().(*bytes.Buffer)
	buf.Reset() // Reset the buffer before use
	defer m.bufPool.Put(buf)

	if err := m.md.Convert(input, buf); err != nil {
		return nil, err
	}

	output := make([]byte, buf.Len())
	copy(output, buf.Bytes()) // Copy the buffer content to output

	return output, nil
}
