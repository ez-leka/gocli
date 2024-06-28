package manpage

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/ez-leka/gocli/renderer"
	"github.com/jedib0t/go-pretty/v6/text"
)

type _Font struct {
	_Element
}

func makeFontElement(parent _ITroffElement, tag_type renderer.TElementTag) _ITroffElement {
	open_macro := "R"
	switch tag_type {
	case renderer.TagEmph:
		open_macro = "\\f[I]"
	case renderer.TagDel:
		open_macro = "\\[R]" // strike font not supported - use R as macro
	case renderer.TagStrong:
		open_macro = "\\f[B]"
	}
	font := &_Font{
		_Element: _Element{
			element_type:   tag_type,
			parent_element: parent,
			my_writer:      bytes.Buffer{},
			textIndent:     0,
			textPrefix:     "",
			openMacro:      open_macro,
			closeMacro:     "P",
		},
	}

	font.openMacro = font.calculateOpenMacro()
	font.closeMacro = font.calculateCloseMacro()

	return font
}

func (e *_Font) calculateOpenMacro() string {
	var pe renderer.IElement
	var rgx = regexp.MustCompile(`\[(.*?)\]`)
	open_macro := rgx.FindStringSubmatch(e.openMacro)[1]
	// here we start with parent
	for pe = e.parent_element; pe != nil; pe = pe.Parent() {
		if pe.GetType() == renderer.TagEmph || pe.GetType() == renderer.TagStrong {
			pm := rgx.FindStringSubmatch(pe.(_ITroffElement).GetOpenMacro())[1]
			tmp := pm + open_macro
			open_macro = tmp
		}
	}
	return fmt.Sprintf("\\f[%s]", open_macro)
}

func (e *_Font) calculateCloseMacro() string {
	var pe renderer.IElement
	// here we start with parent
	for pe = e.parent_element; pe != nil; pe = pe.Parent() {
		if pe.GetType() == renderer.TagEmph || pe.GetType() == renderer.TagStrong {
			return "\\f[P]"
		}
	}
	return "\\f[R]"
}

type _Heading struct {
	_Element
	level int
}

func (e *_Heading) setLevel(level int) {
	e.level = level
	if level == 1 {
		e.openMacro = ".SH "
	} else {
		e.openMacro = ".SS "
	}
}
func (e *_Heading) Literal(literal string) int {
	literal = string(e.escapeSpecialChars([]byte(literal)))
	literal = text.FormatUpper.Apply(literal)

	return e.Out(0, literal)
}

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

func (e *_Link) setUrl(url string) {
	e.Out(0, " "+url)
}

func (e *_Link) Literal(literal string) int {
	e.Out(0, "\n.UE \" "+literal+"\"")
	return 1
}

func (l *_Link) Close() renderer.IElement {
	return l._Element.Close()
}

type _Codeblock struct {
	_Element
	lang string
}

func (e *_Codeblock) Literal(literal string) int {

	lines := e.paddLines(false, literal)

	for _, l := range lines {
		temp_writer := bytes.Buffer{}

		quick.Highlight(&temp_writer, l, e.lang, "noop", "code-block")
		e.Out(e.textIndent, temp_writer.String())
	}
	// e.my_writer.Write(temp_writer.Bytes())

	return len(literal)
}
