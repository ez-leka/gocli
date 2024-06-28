package renderer

import (
	"github.com/jedib0t/go-pretty/v6/text"
)

type TElementTag int

const (
	TagDocument TElementTag = iota

	TagParagraph
	TagBlockquot

	TagList
	TagListItemOrdered
	TagListItemUnordered
	TagListItemTerm
	TagListItemDefinition

	TagEmph
	TagStrong
	TagDel
	TagRoman

	TagHeading
	TagSubHeading

	TagLink
	TagCodeblock
	TagCode

	TagTable
	TagTableHead
	TagTableBody
	TagTableRow
	TagTableCell
)

var (
	NlBytes = []byte{'\n'}
	Nl      = "\n"
)

type IElement interface {
	GetType() TElementTag
	Parent() IElement
	MakeChild(tag_type TElementTag, parent IElement) IElement
	Close() IElement
	ListDepth() int
	Literal(string) int
	HR()
	CR()
	Out(indent int, formatted_string string) int
	Attributes() text.Colors
	SetIndent(int)
	Indent() int
	Prefix() string
	Bytes() []byte
}
