package content

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// KindMermaidBlock 自定义节点：mermaid 围栏代码块，运行时由前端 mermaid 库渲染成图。
var KindMermaidBlock = ast.NewNodeKind("MermaidBlock")

// MermaidBlock 持有 mermaid 源码文本。
type MermaidBlock struct {
	ast.BaseBlock
	Code []byte
}

func (n *MermaidBlock) Kind() ast.NodeKind { return KindMermaidBlock }

func (n *MermaidBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// mermaidTransformer 把 ```mermaid 围栏代码块替换为 MermaidBlock，
// 使其脱离 chroma 语法高亮与 mac 代码块包装（复制/折叠按钮对图表无意义），
// 输出 <pre class="mermaid"> 交给前端渲染。
type mermaidTransformer struct{}

func (t *mermaidTransformer) Transform(node *ast.Document, reader text.Reader, _ parser.Context) {
	source := reader.Source()
	var targets []*ast.FencedCodeBlock

	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if cb, ok := n.(*ast.FencedCodeBlock); ok && isMermaidFence(cb, source) {
			targets = append(targets, cb)
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})

	for _, cb := range targets {
		blk := &MermaidBlock{Code: fencedCodeText(cb, source)}
		parent := cb.Parent()
		parent.ReplaceChild(parent, cb, blk)
	}
}

// isMermaidFence 判断围栏代码块的语言是否为 mermaid（忽略大小写）。
func isMermaidFence(cb *ast.FencedCodeBlock, source []byte) bool {
	lang := cb.Language(source)
	return strings.EqualFold(string(lang), "mermaid")
}

// fencedCodeText 提取围栏代码块的全部行文本。
func fencedCodeText(cb *ast.FencedCodeBlock, source []byte) []byte {
	var buf bytes.Buffer
	lines := cb.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		buf.Write(seg.Value(source))
	}
	return buf.Bytes()
}
