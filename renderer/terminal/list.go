package terminal

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/ez-leka/gocli/renderer"
	"github.com/jedib0t/go-pretty/v6/text"
)

type _List struct {
	_Element
	currentItemIndex int
}

func (e *_List) MakeChild(tag_type renderer.TElementTag, parent renderer.IElement) renderer.IElement {

	if e.currentItemIndex > 0 {
		e.CR()
	}

	new_e := &_ListItem{
		_Element: _Element{
			element_type:   tag_type,
			parent_element: parent,
			my_writer:      bytes.Buffer{},
			textAttributes: nil,
			textIndent:     e.ListDepth() * 4,
			textPrefix:     "",
		},
	}

	switch tag_type {
	case renderer.TagListItemOrdered:
		e.currentItemIndex++
		new_e.textPrefix = fmt.Sprintf("%d. ", e.currentItemIndex)
	case renderer.TagListItemUnordered:
		new_e.textPrefix = "● "
	case renderer.TagListItemTerm:
		new_e.textAttributes = append(new_e.textAttributes, text.Bold)
	case renderer.TagListItemDefinition:
		new_e.textPrefix = "    "
	}

	return new_e
}

type _ListItem struct {
	_Element
}

func (e *_ListItem) Literal(literal string) int {
	indent := e.textIndent
	prefix := e.textPrefix
	lastOutputLen := 0

	// literal can be multi-line we have to indent and prefix every line
	lines := strings.Split(literal, "\n")
	for i, l := range lines {
		use_indent := indent
		use_prefix := prefix
		if i > 0 {
			e.CR()
			use_indent = indent + utf8.RuneCountInString(prefix)
			use_prefix = ""
		}
		lastOutputLen = e.Out(use_indent, fmt.Sprintf("%s%s", use_prefix, l))

	}
	return lastOutputLen

}
