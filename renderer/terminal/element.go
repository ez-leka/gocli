package terminal

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/ez-leka/gocli/renderer"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type _Element struct {
	element_type   renderer.TElementTag
	parent_element renderer.IElement
	my_writer      bytes.Buffer

	textAttributes text.Colors
	textIndent     int
	textPrefix     string
}

func document() *_Element {

	e := _Element{
		element_type:   renderer.TagDocument,
		my_writer:      bytes.Buffer{},
		parent_element: nil,
		textAttributes: nil,
		textIndent:     0,
		textPrefix:     "",
	}

	return &e

}

func (e *_Element) GetType() renderer.TElementTag {
	return e.element_type
}

func (e *_Element) MakeChild(tag_type renderer.TElementTag, parent renderer.IElement) renderer.IElement {

	var new_e renderer.IElement

	switch tag_type {
	case renderer.TagHeading:
		new_e = &_Element{
			element_type:   tag_type,
			parent_element: parent,
			my_writer:      bytes.Buffer{},
			textAttributes: text.Colors{text.Bold},
			textIndent:     0,
			textPrefix:     "",
		}
	case renderer.TagBlockquot:
		new_e = &_Blockquote{
			_Element: _Element{
				element_type:   tag_type,
				parent_element: parent,
				my_writer:      bytes.Buffer{},
				textAttributes: text.Colors{text.ReverseVideo},
				textIndent:     4,
				textPrefix:     "| ",
			},
		}
	case renderer.TagLink:
		link := &_Link{
			_Element: _Element{
				element_type:   tag_type,
				parent_element: parent,
				my_writer:      bytes.Buffer{},
				textAttributes: nil,
				textIndent:     0,
				textPrefix:     "",
			},
		}
		new_e = link

	case renderer.TagList:
		new_e = &_List{
			_Element: _Element{
				element_type:   tag_type,
				parent_element: parent,
				my_writer:      bytes.Buffer{},
				textAttributes: []text.Color{},
				textIndent:     0,
				textPrefix:     "",
			},
			currentItemIndex: 1,
		}
	case renderer.TagTable:
		new_e = &_Table{
			_Element: _Element{
				element_type:   tag_type,
				parent_element: parent,
				my_writer:      bytes.Buffer{},
				textAttributes: []text.Color{},
				textIndent:     0,
				textPrefix:     "",
			},
			tw:     table.NewWriter(),
			header: []string{},
			rows:   [][]string{},
			state:  0,
		}
	case renderer.TagCodeblock:
		new_e = &_Codeblock{
			_Element: _Element{
				element_type:   tag_type,
				parent_element: parent,
				my_writer:      bytes.Buffer{},
				textAttributes: nil,
				textIndent:     4,
				textPrefix:     "",
			},
		}
	case renderer.TagEmph:
		new_e = &_Element{
			element_type:   tag_type,
			parent_element: parent,
			my_writer:      bytes.Buffer{},
			textAttributes: text.Colors{text.Italic},
			textIndent:     0,
			textPrefix:     "",
		}
	case renderer.TagDel:
		new_e = &_Element{
			element_type:   tag_type,
			parent_element: parent,
			my_writer:      bytes.Buffer{},
			textAttributes: text.Colors{text.CrossedOut},
			textIndent:     0,
			textPrefix:     "",
		}
	case renderer.TagStrong:
		new_e = &_Element{
			element_type:   tag_type,
			parent_element: parent,
			my_writer:      bytes.Buffer{},
			textAttributes: text.Colors{text.Bold},
			textIndent:     0,
			textPrefix:     "",
		}
	default:
		new_e = &_Element{
			element_type:   tag_type,
			parent_element: parent,
			my_writer:      bytes.Buffer{},
			textAttributes: nil,
			textIndent:     0,
			textPrefix:     "",
		}
	}
	return new_e
}

func (e *_Element) Writer() io.Writer {
	return &e.my_writer
}

func (e *_Element) Bytes() []byte {
	return e.my_writer.Bytes()
}

func (e *_Element) Close() renderer.IElement {

	data := e.my_writer.String()

	e.Parent().Out(e.textIndent, data)

	return e.Parent()

}
func (e *_Element) CR() {
	e.Writer().Write(renderer.NlBytes)
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

func (e *_Element) Out(indent int, formatter_string string) int {
	indent_str := strings.Repeat(" ", indent)

	s := indent_str + e.Attributes().Sprintf("%s", formatter_string)
	lastOutputLen, _ := e.Writer().Write([]byte(s))
	return lastOutputLen
}

func (e *_Element) HR() {
	e.Out(0, strings.Repeat("─", terminalWidth()))
}

func (e *_Element) Parent() renderer.IElement {
	return e.parent_element
}

func (e *_Element) SetIndent(indent int) {
	e.textIndent = indent
}

// attributes are accumulated as element are nested - so we need to get all current attributes
func (e *_Element) Attributes() text.Colors {
	attrs := text.Colors{}

	if e.textAttributes != nil {
		attrs = append(attrs, e.textAttributes...)
	}
	for pe := e.Parent(); pe != nil; pe = pe.Parent() {
		attrs = append(attrs, pe.Attributes()...)
	}

	return attrs
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
		lines[i] = fmt.Sprintf("%s%s%s%s", indent_str, prefix, l, padding)
	}
	return lines
}

func (e *_Element) setBg(r, g, b int) {
	e.my_writer.Write([]byte(fmt.Sprintf("\x1b[48:2:%d:%d:%dm", r, g, b)))
}

func (e *_Element) resetBg() {
	e.my_writer.Write([]byte("\x1b[0m"))
}
