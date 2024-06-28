package manpage

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/ez-leka/gocli/renderer"
	"github.com/jedib0t/go-pretty/v6/text"
)

type _ITroffElement interface {
	renderer.IElement
	GetOpenMacro() string
}

type _Element struct {
	element_type   renderer.TElementTag
	parent_element _ITroffElement
	my_writer      bytes.Buffer

	textIndent int
	textPrefix string

	openMacro  string
	closeMacro string
}

func document() *_Element {

	e := _Element{
		element_type:   renderer.TagDocument,
		my_writer:      bytes.Buffer{},
		parent_element: nil,
		textIndent:     0,
		textPrefix:     "",
	}

	return &e

}

func (e *_Element) GetType() renderer.TElementTag {
	return e.element_type
}

func (e *_Element) MakeChild(tag_type renderer.TElementTag, parent renderer.IElement) renderer.IElement {

	var new_e _ITroffElement

	switch tag_type {
	case renderer.TagHeading:
		new_e = &_Heading{
			_Element: _Element{
				element_type:   tag_type,
				parent_element: parent.(_ITroffElement),
				my_writer:      bytes.Buffer{},
				textIndent:     0,
				textPrefix:     "",
				openMacro:      ".SH ",
				closeMacro:     renderer.Nl,
			},
		}
	case renderer.TagParagraph:
		new_e = &_Element{
			element_type:   tag_type,
			parent_element: parent.(_ITroffElement),
			my_writer:      bytes.Buffer{},
			textIndent:     0,
			textPrefix:     "",
			openMacro:      ".PP\n",
			closeMacro:     renderer.Nl,
		}
	case renderer.TagBlockquot:
		new_e = &_Blockquote{
			_Element: _Element{
				element_type:   tag_type,
				parent_element: parent.(_ITroffElement),
				my_writer:      bytes.Buffer{},
				textIndent:     4,
				textPrefix:     "| ",
			},
		}
	case renderer.TagLink:
		link := &_Link{
			_Element: _Element{
				element_type:   tag_type,
				parent_element: parent.(_ITroffElement),
				my_writer:      bytes.Buffer{},
				textIndent:     0,
				textPrefix:     "",
				openMacro:      "\n.UR",
				closeMacro:     "\n",
			},
		}
		new_e = link

	case renderer.TagList:
		open_macro := ""
		close_macro := ""

		if e.ListDepth() >= 1 {
			open_macro = "\n.RS\n"
			close_macro = "\n.RE\n"
		}

		new_e = &_List{
			_Element: _Element{
				element_type:   tag_type,
				parent_element: parent.(_ITroffElement),
				my_writer:      bytes.Buffer{},
				textIndent:     0,
				textPrefix:     "",
				openMacro:      open_macro,
				closeMacro:     close_macro,
			},
			currentItemIndex: 1,
		}

	case renderer.TagTable:
		new_e = makeTable(e)
	case renderer.TagCodeblock:
		new_e = &_Codeblock{
			_Element: _Element{
				element_type:   tag_type,
				parent_element: parent.(_ITroffElement),
				my_writer:      bytes.Buffer{},
				textIndent:     0,
				textPrefix:     "",
				openMacro:      "\n.EX\n",
				closeMacro:     "\n.EE\n",
			},
		}
	case renderer.TagEmph:
		new_e = makeFontElement(parent.(_ITroffElement), tag_type)
	case renderer.TagDel:
		new_e = makeFontElement(parent.(_ITroffElement), tag_type)
	case renderer.TagStrong:
		new_e = makeFontElement(parent.(_ITroffElement), tag_type)
	default:
		new_e = &_Element{
			element_type:   tag_type,
			parent_element: parent.(_ITroffElement),
			my_writer:      bytes.Buffer{},
			textIndent:     0,
			textPrefix:     "",
		}
	}

	return new_e
}

func (e *_Element) Open() {

}

func (e *_Element) Writer() io.Writer {
	return &e.my_writer
}

func (e *_Element) Bytes() []byte {
	return e.my_writer.Bytes()
}

func (e *_Element) Close() renderer.IElement {

	e.Out(0, e.closeMacro)

	data := e.my_writer.String()

	e.Parent().Out(e.textIndent, data)

	return e.Parent()

}
func (e *_Element) CR() {
	e.my_writer.WriteString("\n.br\n")
}

func (e *_Element) escapeSpecialChars(text []byte) []byte {
	escaped := []byte{}

	escape_seq := []byte("\\&")
	for i := 0; i < len(text); i++ {
		// escape initial apostrophe or period
		if len(text) >= 1 && (text[0] == '\'' || text[0] == '.') {
			escaped = append(escaped, escape_seq...)
		}
		escaped = append(escaped, text[i])
	}
	return escaped
}

func (e *_Element) Literal(literal string) int {
	prefix := e.textPrefix
	lastOutputLen := 0

	// literal can be multi-line we have to indent and prefix every line
	lines := strings.Split(literal, "\n")
	for _, l := range lines {
		lastOutputLen = e.Out(e.textIndent, fmt.Sprintf("%s%s", prefix, l))
	}
	return lastOutputLen
}

func (e *_Element) Out(indent int, formatted_string string) int {
	if e.my_writer.Len() == 0 {
		// make sure to start with open tag
		e.Writer().Write([]byte(e.GetOpenMacro()))
	}
	indent_str := strings.Repeat(" ", indent)

	s := indent_str + formatted_string
	lastOutputLen, _ := e.Writer().Write([]byte(s))
	return lastOutputLen
}

func (e *_Element) HR() {
	e.Out(0, strings.Repeat("─", terminalWidth()-8))
}

func (e *_Element) Parent() renderer.IElement {
	return e.parent_element
}

func (e *_Element) Attributes() text.Colors {
	return nil
}

func (e *_Element) SetIndent(indent int) {
	e.textIndent = indent
}

func (e *_Element) Indent() int {
	return e.textIndent
}
func (e *_Element) Prefix() string {
	return e.textPrefix
}

func (e *_Element) ListDepth() int {
	depth := 0

	var pe renderer.IElement

	for pe = e; pe != nil; pe = pe.Parent() {
		if pe.GetType() == renderer.TagList {
			depth++
		}
	}
	return depth
}

func (e *_Element) GetOpenMacro() string {
	return e.openMacro
}
func (e *_Element) GetCloseMacro() string {
	return e.closeMacro
}

func (e *_Element) paddLines(with_indent bool, literal string) []string {

	prefix := e.Prefix()
	indent_str := ""

	if with_indent {
		indent_str = strings.Repeat(" ", e.textIndent)
	}

	// first, replace tabs with  spaces
	literal = strings.ReplaceAll(literal, "\t", "    ")

	// split into lines and calculate length includign tabs and indent
	lines := strings.Split(literal, "\n")
	// calculate longest line
	max_len := e.textIndent
	for _, l := range lines {
		if len(l) > max_len {
			max_len = len(l) + len(indent_str) + len(prefix)
		}
	}

	for i, l := range lines {
		padding := strings.Repeat(" ", max_len-len(l))
		// only apply attributes after indent
		lines[i] = fmt.Sprintf("%s%s%s%s\n", indent_str, prefix, l, padding)

	}
	return lines
}
