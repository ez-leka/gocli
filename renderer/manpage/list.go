package manpage

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ez-leka/gocli/renderer"
)

type _List struct {
	_Element
	numItems         int
	currentItemIndex int

	indent int
}

type _ListItem struct {
	_Element
}

func (e *_List) caculateIndents(idx int) (padding string) {
	if e.indent == 0 {
		// this is done once for teh first item
		// count digits in largest index
		i := e.numItems
		for i != 0 {
			i /= 10
			e.indent++
		}
		// add 1 position for . and 1 for space after
		e.indent += 2
	}
	// calculate padding as diff in max num of digits and digins in index
	p := 0
	for idx != 0 {
		idx /= 10
		p++
	}
	p = e.indent - 2 - p
	for i := 0; i < p; i++ {
		padding += "\\ "
	}

	return
}

func (e *_List) MakeChild(tag_type renderer.TElementTag, parent renderer.IElement) renderer.IElement {

	new_e := &_ListItem{
		_Element: _Element{
			element_type:   tag_type,
			parent_element: e,
			my_writer:      bytes.Buffer{},
			textIndent:     e.ListDepth() * 4,
			textPrefix:     "",
		},
	}

	switch tag_type {
	case renderer.TagListItemOrdered:
		new_e.openMacro = fmt.Sprintf("\n.IP %s%d. %d\n", e.caculateIndents(e.currentItemIndex), e.currentItemIndex, e.indent)
		e.currentItemIndex++
	case renderer.TagListItemUnordered:
		new_e.openMacro = "\n.IP \\(bu 2\n"
		new_e.closeMacro = "\n"
		new_e.textIndent = 0
	case renderer.TagListItemTerm:
		new_e.openMacro = "\n.TP\n"
		new_e.textIndent = 0
	case renderer.TagListItemDefinition:
		new_e.openMacro = "\n.TQ\n.RS 1\n"
		new_e.closeMacro = "\n.RE\n"
		new_e.textIndent = 0
	}

	return new_e
}

func (e *_ListItem) Literal(literal string) int {
	lastOutputLen := 0

	// literal can be multi-line we have to indent and prefix every line
	lines := strings.Split(literal, "\n")
	for i, l := range lines {
		if i > 0 {
			e.CR()
		}
		lastOutputLen = e.Out(0, l)

	}
	return lastOutputLen

}
