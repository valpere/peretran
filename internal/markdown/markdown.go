package markdown

import (
	"bytes"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func ToHTML(md []byte) string {
	opts := html.RendererOptions{
		Flags: html.CommonFlags | html.HrefTargetBlank,
	}
	renderer := html.NewRenderer(opts)
	ext := parser.CommonExtensions | parser.Attributes
	p := parser.NewWithExtensions(ext)
	doc := p.Parse(md)
	return string(markdown.Render(doc, renderer))
}

func ToPlainText(md []byte) string {
	htmlContent := ToHTML(md)
	return StripHTMLTags(htmlContent)
}

func StripHTMLTags(htmlContent string) string {
	var result bytes.Buffer
	inTag := false

	for _, ch := range htmlContent {
		switch ch {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				result.WriteRune(ch)
			}
		}
	}

	return result.String()
}
