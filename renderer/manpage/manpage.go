package manpage

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/ez-leka/gocli/renderer"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/russross/blackfriday/v2"
)

// _Manpage implements the blackfriday.Renderer interface for creating
// roff format (manpages) from markdown text
type _Manpage struct {
	cmd string
	//listCounters []int
	listCounters renderer.Stack[*int]
	startDef     bool
	listDepth    int

	currentTag blackfriday.NodeType

	current_content string
}

const (
	_TitleTag             = ".TH "
	_SectionTag           = ".SH "
	_SubSectionTag        = ".SS "
	topLevelHeader        = "\n\n.SH "
	secondLevelHdr        = "\n.SH "
	otherHeader           = "\n.SS "
	crTag                 = "\n"
	emphTag               = "\\fI"
	emphCloseTag          = "\\fR"
	strongTag             = "\\fB"
	strongCloseTag        = "\\fR"
	breakTag              = "\n.br\n"
	paraTag               = "\n.PP\n"
	hruleTag              = "\n.ti 0\n\\l'\\n(.lu'\n"
	_linkTag              = "\n.UR "
	_linkCloseTag         = "\n.UE ."
	codespanTag           = "\\fB\\fC"
	codespanCloseTag      = "\\fR"
	codeTag               = "\n.EX 4\n"
	codeCloseTag          = "\n.EE\n"
	quoteTag              = "\n.PP\n.RS\n"
	quoteCloseTag         = "\n.RE\n"
	_listTag              = "\n.RS\n"
	_listCloseTag         = "\n.RE\n"
	_listOrderedItemTag   = ".IP %3d. %d\n"
	_listUnorderedItemTag = ".IP \\(bu 2\n"
	dtTag                 = "\n.TP\n"
	dd2Tag                = "\n"
	_TableStartTag        = "\n.TS\nallbox;\n"
	_tableEndTag          = ".TE\n"
	tableCellStart        = "T{\n"
	tableCellEnd          = "\nT}\n"
)

// creates a new blackfriday Renderer for generating roff documents
// from markdown for man pages
func ManpageRenderer(cmd string) *_Manpage {
	return &_Manpage{
		cmd:          cmd,
		listCounters: renderer.Stack[*int]{},
	}
}

// RenderHeader handles outputting the header at document start
func (r *_Manpage) RenderHeader(w io.Writer, ast *blackfriday.Node) {
	// disable hyphenation
	r.out(w, ".nh\n")

	r.sprintf(w, "%s %s %d%s", _TitleTag, r.cmd, 1, renderer.Nl)
}

// RenderFooter handles outputting the footer at the document end; the roff
// renderer has no footer information
func (r *_Manpage) RenderFooter(w io.Writer, ast *blackfriday.Node) {
}

// RenderNode is called for each node in a markdown document; based on the node
// type the equivalent roff output is sent to the writer
func (r *_Manpage) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	walkAction := blackfriday.GoToNext

	r.current_content = w.(*bytes.Buffer).String()

	switch node.Type {
	case blackfriday.Text:
		lit := string(node.Literal)
		if r.currentTag == blackfriday.Heading {
			lit = text.FormatUpper.Apply(lit)
		}
		r.escapeSpecialChars(w, []byte(lit))
	case blackfriday.Softbreak:
		r.out(w, crTag)
	case blackfriday.Hardbreak:
		r.out(w, breakTag)
	case blackfriday.Emph:
		if entering {
			r.out(w, emphTag)
		} else {
			r.out(w, emphCloseTag)
		}
	case blackfriday.Strong:
		if entering {
			r.out(w, strongTag)
		} else {
			r.out(w, strongCloseTag)
		}
	case blackfriday.Link:
		if entering {
			escapedDest := r.prepareUrl(string(node.LinkData.Destination))
			r.sprintf(w, "%s %s\n", _linkTag, escapedDest)
		} else {
			r.out(w, _linkCloseTag)
		}
	case blackfriday.Image:
		// ignore images
		walkAction = blackfriday.SkipChildren
	case blackfriday.Code:
		r.out(w, codespanTag)
		r.escapeSpecialChars(w, node.Literal)
		r.out(w, codespanCloseTag)
	case blackfriday.Document:
		break
	case blackfriday.Paragraph:
		// roff .PP markers break lists
		if r.listDepth > 0 {
			return blackfriday.GoToNext
		}
		if entering {
			r.out(w, paraTag)
		} else {
			r.out(w, crTag)
		}
	case blackfriday.BlockQuote:
		if entering {
			r.out(w, quoteTag)
		} else {
			r.out(w, quoteCloseTag)
		}
	case blackfriday.Heading:
		if entering {
			r.currentTag = blackfriday.Heading
			if node.Level == 1 {
				r.out(w, _SectionTag)
			} else {
				r.out(w, _SubSectionTag)
			}
		} else {
			r.out(w, renderer.Nl)
			r.currentTag = blackfriday.Document
		}
	case blackfriday.HorizontalRule:
		r.out(w, hruleTag)
	case blackfriday.List:
		openTag := _listTag
		closeTag := _listCloseTag
		if node.ListFlags&blackfriday.ListTypeDefinition != 0 {
			// tags for definition lists handled within Item node
			openTag = ""
			closeTag = ""
		}
		if entering {
			r.listDepth++
			if node.ListFlags&blackfriday.ListTypeOrdered != 0 {
				idx := 1
				r.listCounters.Push(&idx)

			}
			r.out(w, openTag)
		} else {
			if node.ListFlags&blackfriday.ListTypeOrdered != 0 {
				r.listCounters.Pop()
			}
			r.out(w, closeTag)
			r.listDepth--
		}
	case blackfriday.Item:
		if entering {
			if node.ListFlags&blackfriday.ListTypeOrdered != 0 {
				idx, _ := r.listCounters.Peek()
				r.sprintf(w, _listOrderedItemTag, *idx, 4)
				*idx++
			} else if node.ListFlags&blackfriday.ListTypeTerm != 0 {
				// DT (definition term): line just before DD (see below).
				r.out(w, dtTag)
				r.startDef = true
			} else if node.ListFlags&blackfriday.ListTypeDefinition != 0 {
				if r.startDef {
					r.startDef = false
				} else {
					r.out(w, dd2Tag)
				}
			} else {
				r.out(w, _listUnorderedItemTag)
			}
		} else {
			r.out(w, "\n")
		}
	case blackfriday.CodeBlock:
		r.out(w, codeTag)
		r.escapeSpecialChars(w, node.Literal)
		r.out(w, codeCloseTag)
	case blackfriday.Table:
		r.handleTable(w, node, entering)
	case blackfriday.TableHead:
	case blackfriday.TableBody:
	case blackfriday.TableRow:
		// no action as cell entries do all the nroff formatting
		return blackfriday.GoToNext
	case blackfriday.TableCell:
		r.handleTableCell(w, node, entering)
	case blackfriday.HTMLSpan:
		// ignore other HTML tags
	default:
		// we do not handle this tag - skip children to be safe
		walkAction = blackfriday.SkipChildren
	}
	return walkAction
}

func (r *_Manpage) handleTable(w io.Writer, node *blackfriday.Node, entering bool) {
	if entering {
		r.out(w, _TableStartTag)
		// call walker to count cells (and rows?) so format section can be produced
		columns := countColumns(node)
		r.out(w, strings.Repeat("l ", columns)+"\n")
		r.out(w, strings.Repeat("l ", columns)+".\n")
	} else {
		r.out(w, _tableEndTag)
	}
}

func (r *_Manpage) handleTableCell(w io.Writer, node *blackfriday.Node, entering bool) {
	if entering {
		var start string
		if node.Prev != nil && node.Prev.Type == blackfriday.TableCell {
			start = "\t"
		}
		if node.IsHeader {
			start += strongTag
		} else if nodeLiteralSize(node) > 30 {
			start += tableCellStart
		}
		r.out(w, start)
	} else {
		var end string
		if node.IsHeader {
			end = strongCloseTag
		} else if nodeLiteralSize(node) > 30 {
			end = tableCellEnd
		}
		if node.Next == nil && end != tableCellEnd {
			// Last cell: need to carriage return if we are at the end of the
			// header row and content isn't wrapped in a "tablecell"
			end += crTag
		}
		r.out(w, end)
	}
}

func nodeLiteralSize(node *blackfriday.Node) int {
	total := 0
	for n := node.FirstChild; n != nil; n = n.FirstChild {
		total += len(n.Literal)
	}
	return total
}

// because roff format requires knowing the column count before outputting any table
// data we need to walk a table tree and count the columns
func countColumns(node *blackfriday.Node) int {
	var columns int

	node.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		switch node.Type {
		case blackfriday.TableRow:
			if !entering {
				return blackfriday.Terminate
			}
		case blackfriday.TableCell:
			if entering {
				columns++
			}
		default:
		}
		return blackfriday.GoToNext
	})
	return columns
}

func (r *_Manpage) prepareUrl(url string) string {
	escapedLink := strings.ReplaceAll(url, "-", "\\-")
	parts := strings.Split(escapedLink, "//")

	if len(parts) == 2 {
		url = parts[0] + "//\\:" + strings.ReplaceAll(parts[1], "/", "/\\:")
	} else {
		url = parts[0]
	}
	return url
}

func (r *_Manpage) out(w io.Writer, output string) {
	io.WriteString(w, output)
}

func (r *_Manpage) sprintf(w io.Writer, format string, args ...interface{}) {
	io.WriteString(w, fmt.Sprintf(format, args...))
}

func (r *_Manpage) escapeSpecialChars(w io.Writer, text []byte) {
	for i := 0; i < len(text); i++ {
		// escape initial apostrophe or period
		if len(text) >= 1 && (text[0] == '\'' || text[0] == '.') {
			r.out(w, "\\&")
		}

		// directly copy normal characters
		org := i

		for i < len(text) && text[i] != '\\' {
			i++
		}
		if i > org {
			w.Write(text[org:i]) // nolint: errcheck
		}

		// escape a character
		if i >= len(text) {
			break
		}

		w.Write([]byte{'\\', text[i]}) // nolint: errcheck
	}
}
