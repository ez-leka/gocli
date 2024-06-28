package terminal

import (
	"bytes"

	"github.com/ez-leka/gocli/renderer"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type _Table struct {
	_Element

	tw     table.Writer
	header []string
	rows   [][]string
	state  int
}

type _TableSection struct {
	_Element
	isHeader bool
}

type _TableRow struct {
	_Element
	table.Row
	isHeader bool
	columns  []table.ColumnConfig
}

type _TableCell struct {
	_Element
	align text.Align
}

func (th *_TableSection) MakeChild(tag_type renderer.TElementTag, parent renderer.IElement) renderer.IElement {

	return &_TableRow{
		_Element: _Element{
			element_type:   renderer.TagTableRow,
			parent_element: th,
			my_writer:      bytes.Buffer{},
			textAttributes: nil,
			textIndent:     0,
			textPrefix:     "",
		},
		Row:      []interface{}{},
		isHeader: th.isHeader,
		columns:  []table.ColumnConfig{},
	}
}
func (th *_TableSection) Close() renderer.IElement {
	return th.parent_element
}

func (tr *_TableRow) MakeChild(tag_type renderer.TElementTag, parent renderer.IElement) renderer.IElement {
	return &_TableCell{
		_Element: _Element{
			element_type:   renderer.TagTableCell,
			parent_element: tr,
			my_writer:      bytes.Buffer{},
			textAttributes: nil,
			textIndent:     0,
			textPrefix:     "",
		},
	}

}

func (tr *_TableRow) Close() renderer.IElement {
	table := tr.parent_element.(*_TableSection).parent_element.(*_Table)
	table.tw.SetColumnConfigs(tr.columns)
	if tr.isHeader {
		table.tw.AppendHeader(tr.Row)
	} else {
		table.tw.AppendRow(tr.Row)
	}

	return tr.parent_element
}

func (td *_TableCell) setAlignment(align text.Align) {
	td.align = align

}
func (td *_TableCell) Close() renderer.IElement {

	data := td.my_writer.Bytes()

	row := td.parent_element.(*_TableRow)
	row.columns = append(row.columns, table.ColumnConfig{
		Number: len(row.columns) + 1,
		Align:  td.align,
	})
	row.Row = append(row.Row, string(data))
	return td.Parent()
}

func (t *_Table) MakeChild(tag_type renderer.TElementTag, parent renderer.IElement) renderer.IElement {

	tr := &_TableSection{
		_Element: _Element{
			element_type:   renderer.TagTableHead,
			parent_element: t, my_writer: bytes.Buffer{},
			textAttributes: nil,
			textIndent:     0,
			textPrefix:     "",
		},
	}
	switch tag_type {
	case renderer.TagTableHead:
		tr.isHeader = true
	case renderer.TagTableRow:
		tr.isHeader = false
	}

	return tr
}

func (t *_Table) Close() renderer.IElement {

	s := t.tw.Render()

	t.Parent().Out(0, s)

	return t.Parent()

}
