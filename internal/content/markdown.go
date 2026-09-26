package content

import (
	"bytes"
	"sync"

	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

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
			highlighting.NewHighlighting(
				highlighting.WithStyle(highlightStyle),
			), // Use the specified style for syntax highlighting

		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // Automatically generate heading IDs
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(), // Allow raw HTML in Markdown
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
