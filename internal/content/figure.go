package content

import (
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindFigureBlock 自定义节点：独占一段的图片，渲染为 <figure class="figure">。
var KindFigureBlock = ast.NewNodeKind("FigureBlock")

// FigureBlock 包装一张独占一行的图片。
type FigureBlock struct {
	ast.BaseBlock
}

func (n *FigureBlock) Kind() ast.NodeKind { return KindFigureBlock }

func (n *FigureBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// figureTransformer 把「仅包含一张图片的段落」转换为 FigureBlock，
// 使独占一行的图片渲染为带图注的 <figure>，行内混排图片不受影响。
type figureTransformer struct{}

func (t *figureTransformer) Transform(node *ast.Document, reader text.Reader, _ parser.Context) {
	source := reader.Source()
	var targets []*ast.Paragraph

	// 先收集再替换，避免遍历过程中修改树结构
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if para, ok := n.(*ast.Paragraph); ok && isImageOnlyParagraph(para, source) {
			targets = append(targets, para)
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})

	for _, para := range targets {
		img := para.FirstChild().(*ast.Image)
		fig := &FigureBlock{}
		fig.AppendChild(fig, img)
		parent := para.Parent()
		parent.ReplaceChild(parent, para, fig)
	}
}

// isImageOnlyParagraph 判断段落是否只包含一张图片（允许夹带纯空白文本节点）。
func isImageOnlyParagraph(para *ast.Paragraph, source []byte) bool {
	imgCount := 0
	for child := para.FirstChild(); child != nil; child = child.NextSibling() {
		switch c := child.(type) {
		case *ast.Image:
			imgCount++
		case *ast.Text:
			if strings.TrimSpace(string(c.Value(source))) != "" {
				return false
			}
		default:
			return false
		}
	}
	return imgCount == 1
}

// figureHTMLRenderer 渲染 FigureBlock：<figure class="figure"><img>…<figcaption>。
type figureHTMLRenderer struct{}

func (r *figureHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindFigureBlock, r.render)
	reg.Register(KindMermaidBlock, r.renderMermaid)
}

// renderMermaid 输出 mermaid 占位块，源码转义后由前端 mermaid 库渲染。
func (r *figureHTMLRenderer) renderMermaid(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	blk := n.(*MermaidBlock)
	w.WriteString(`<div class="mermaid-block"><pre class="mermaid">`)
	w.Write(util.EscapeHTML(blk.Code))
	w.WriteString(`</pre></div>`)
	return ast.WalkSkipChildren, nil
}

func (r *figureHTMLRenderer) render(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString(`<figure class="figure">`)
		// <img> 由 goldmark 默认的 html.Image 渲染器输出（处理转义、title 等）
		return ast.WalkContinue, nil
	}

	// 退出时用图片 alt 文本输出图注。
	// alt 常常是自动生成的文件名（image-20250809…），那不是图注 —— 这种时候
	// 输出空图注，让 CSS 的计数器只显示「图 N」徽标。
	if img, ok := n.FirstChild().(*ast.Image); ok {
		alt := extractAlt(img, source)
		if isFilenameAlt(alt) {
			alt = ""
		}
		// 有 <img> 就一定给 figcaption：编号徽标由 CSS 的 ::before 生成
		w.WriteString(`<figcaption class="figure__caption">`)
		w.Write(util.EscapeHTML([]byte(alt)))
		w.WriteString(`</figcaption>`)
	}
	w.WriteString(`</figure>`)
	return ast.WalkContinue, nil
}

// isFilenameAlt 判断 alt 是否只是自动生成的文件名（image-20250809130637620 之类）。
func isFilenameAlt(alt string) bool {
	if alt == "" {
		return false
	}
	if !strings.HasPrefix(alt, "image-") && !strings.HasPrefix(alt, "image_") {
		return false
	}
	// image-<数字> 或 image_<数字> 形式（数字之间允许 - 与 _）
	for _, r := range alt[6:] {
		if r != '-' && r != '_' && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

// extractAlt 收集节点的全部文本作为 alt（goldmark 用 Text 子节点承载 alt 内容）。
func extractAlt(n ast.Node, source []byte) string {
	var b strings.Builder
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if t, ok := c.(*ast.Text); ok {
				b.Write(t.Value(source))
			}
		}
		return ast.WalkContinue, nil
	})
	return strings.TrimSpace(b.String())
}
