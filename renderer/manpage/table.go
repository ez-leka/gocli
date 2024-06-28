package manpage

import (
	"bytes"

	"github.com/ez-leka/gocli/renderer"
	"github.com/jedib0t/go-pretty/v6/text"
)

type _Table struct {
	_Element
	format [][]string
	data   [][]string
}

type _TableRow struct {
	_Element
	row int
}

type _TableCell struct {
	_Element
	row  int
	cell int
}

func makeTable(parent _ITroffElement) _ITroffElement {
	new_e := &_Table{
		_Element: _Element{
			element_type:   renderer.TagTable,
			parent_element: parent,
			my_writer:      bytes.Buffer{},
			textIndent:     0,
			textPrefix:     "",
			openMacro:      ".TS\nallbox tab (@);\n",
			closeMacro:     ".TE",
		},
		format: make([][]string, 0),
		data:   make([][]string, 0),
	}

	return new_e
}

func (td *_TableCell) setAlignment(align text.Align) {
	format := "l"
	switch align {
	case text.AlignCenter:
		format = "c"
	case text.AlignLeft:
		format = "l"
	case text.AlignRight:
		format = "r"
	}
	td.parent_element.Parent().(*_Table).format[td.row][td.cell] = format
}

func (td *_TableCell) Close() renderer.IElement {

	data := td.my_writer.String()

	td.parent_element.Parent().(*_Table).data[td.row][td.cell] = data
	return td.parent_element
}

func (t *_Table) MakeChild(tag_type renderer.TElementTag, parent renderer.IElement) renderer.IElement {

	tc := &_TableRow{
		_Element: _Element{
			element_type:   renderer.TagTableRow,
			parent_element: t,
			my_writer:      bytes.Buffer{},
			textIndent:     0,
			textPrefix:     "",
			openMacro:      "",
			closeMacro:     "",
		},
		row: len(t.format),
	}
	t.format = append(t.format, make([]string, 0))
	t.data = append(t.data, make([]string, 0))

	return tc
}

func (tr *_TableRow) MakeChild(tag_type renderer.TElementTag, parent renderer.IElement) renderer.IElement {
	t := tr.parent_element.(*_Table)

	tc := &_TableCell{
		_Element: _Element{
			element_type:   renderer.TagTableCell,
			parent_element: tr,
			my_writer:      bytes.Buffer{},
			textIndent:     0,
			textPrefix:     "",
			openMacro:      "",
			closeMacro:     "",
		},
		row:  tr.row,
		cell: len(t.data[tr.row]),
	}

	t.format[tr.row] = append(t.format[tr.row], "")
	t.data[tr.row] = append(t.data[tr.row], "")

	return tc
}

func (t *_Table) Close() renderer.IElement {

	t.my_writer.WriteString(t.openMacro)

	for i, fr := range t.format {
		for _, f := range fr {
			t.my_writer.WriteString(f + " ")
		}
		if i == len(t.format)-1 {
			t.my_writer.WriteRune('.')
		}
		t.my_writer.WriteString(renderer.Nl)
	}
	for _, dr := range t.data {
		for _, d := range dr {
			t.my_writer.WriteString(d + "@")
		}
		t.my_writer.WriteString(renderer.Nl)
	}

	return t._Element.Close()
}
