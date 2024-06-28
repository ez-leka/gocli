package terminal

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/ez-leka/gocli/renderer"
)

type _Blockquote struct {
	_Element
}

func (e *_Blockquote) Literal(literal string) int {

	lines := e.paddLines(false, literal)

	lastOutputLen := 0
	for _, l := range lines {
		lastOutputLen = e.Out(4, l)
	}

	return lastOutputLen
}

type _Link struct {
	_Element
}

func (e *_Element) startLink(url string) {
	s := fmt.Sprintf("\x1b]8;;%s \x1b\\", url)
	e.Out(0, s)
}

func (l *_Link) Close() renderer.IElement {
	l.Out(0, "\x1b]8;;\x1b\\")
	return l._Element.Close()
}

type _Codeblock struct {
	_Element
	lang string
}

func (e *_Codeblock) Literal(literal string) int {

	// trim new lines
	literal = strings.Trim(literal, "\n")
	lines := e.paddLines(false, literal)

	for i, l := range lines {
		temp_writer := bytes.Buffer{}
		indent := 0
		if i > 0 {
			e.Out(indent, "\n")
			// for all lines except first one add parent indent since
			// becouse for the rirst line it will be added when element is added to the parent
			indent = e.textIndent
		}
		// for all lines add padding to the left and right to create clear box
		l = " " + l + " "
		quick.Highlight(&temp_writer, l, e.lang, "terminal16m", "code-block")

		e.Out(indent, temp_writer.String())
	}

	return len(literal)
}
