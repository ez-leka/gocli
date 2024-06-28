package terminal

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/ez-leka/gocli/renderer"
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/olekukonko/ts"
	"github.com/russross/blackfriday/v2"
)

type _Terminal struct {
	lastOutputLen int

	width int

	element renderer.IElement
}

// RenderFooter implements blackfriday.Renderer.
func (*_Terminal) RenderFooter(w io.Writer, ast *blackfriday.Node) {
	// nothing to do for terminal
}

// RenderHeader implements blackfriday.Renderer.
func (*_Terminal) RenderHeader(w io.Writer, ast *blackfriday.Node) {
	// noting to do for terminal
}

func (r *_Terminal) openTag(tag_type renderer.TElementTag) {
	r.element = r.element.MakeChild(tag_type, r.element)
}

func (r *_Terminal) closeTag() {
	r.element = r.element.Close()
}

func (r *_Terminal) out(text string) {
	r.element.Out(0, text)
	r.lastOutputLen = len(text)
}

func (r *_Terminal) cr() {
	if r.lastOutputLen > 0 {
		r.element.CR()
	}
}

func (r *_Terminal) Literal(literal string) {
	r.lastOutputLen = r.element.Literal(literal)
}

// RenderNode is a default renderer of a single node of a syntax tree. For
// block nodes it will be called twice: first time with entering=true, second
// time with entering=false, so that it could know when it's working on an open
// tag and when on close. It writes the result to w.
//
// The return value is a way to tell the calling walker to adjust its walk
// pattern: e.g. it can terminate the traversal by returning Terminate. Or it
// can ask the walker to skip a subtree of this node by returning SkipChildren.
// The typical behavior is to return GoToNext, which asks for the usual
// traversal to the next node.
func (r *_Terminal) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {

	switch node.Type {
	case blackfriday.Text:
		r.Literal(string(node.Literal))
	case blackfriday.Softbreak:
		r.cr()
	case blackfriday.Hardbreak:
		r.cr()
	case blackfriday.Emph:
		if entering {
			r.openTag(renderer.TagEmph)
		} else {
			r.closeTag()
		}
	case blackfriday.Strong:
		if entering {
			r.openTag(renderer.TagStrong)
		} else {
			r.closeTag()
		}
	case blackfriday.Del:
		if entering {
			r.openTag(renderer.TagDel)
		} else {
			r.closeTag()
		}
	case blackfriday.HTMLSpan:
		r.element.Literal(string(node.Literal))
	case blackfriday.Link:
		// mark it but don't link it if it is not a safe link: no smartypants
		if entering {
			r.openTag(renderer.TagLink)
			dest := node.LinkData.Destination
			link := r.element.(*_Link)
			link.startLink(string(dest))
		} else {
			r.closeTag()
		}
	case blackfriday.Image:
		// skip images
		return blackfriday.SkipChildren
	case blackfriday.Code:
		// this tag does not have enter/exit - it is same as literal
		// but we write it faint
		r.out(text.Faint.Sprint(string(node.Literal)))
	case blackfriday.Document:
		if entering {
			r.element = document()
		} else {
			w.Write(r.element.Bytes())
		}
	case blackfriday.Paragraph:
		if skipParagraphTags(node) {
			break
		}
		if entering {
			if node.Prev != nil {
				switch node.Prev.Type {
				case blackfriday.HTMLBlock, blackfriday.List, blackfriday.Paragraph, blackfriday.Heading, blackfriday.CodeBlock, blackfriday.BlockQuote, blackfriday.HorizontalRule:
					r.cr()
				}
			}
			if node.Parent.Type == blackfriday.BlockQuote && node.Prev == nil {
				r.cr()
			}
			r.out(renderer.Nl)
		} else {
			r.out(renderer.Nl)
			if !(node.Parent.Type == blackfriday.Item && node.Next == nil) {
				r.cr()
			}
		}
	case blackfriday.BlockQuote:

		if entering {
			r.openTag(renderer.TagBlockquot)
		} else {
			r.closeTag()
		}

	case blackfriday.HTMLBlock:
		r.cr()
		r.out(string(node.Literal))
		r.cr()
	case blackfriday.Heading:
		if entering {
			r.openTag(renderer.TagHeading)
			r.element.SetIndent((node.Level - 1) * 4)
			r.cr()
		} else {
			// reset attributes
			r.closeTag()
			if !(node.Parent.Type == blackfriday.Item && node.Next == nil) {
				r.cr()
			}
		}
	case blackfriday.HorizontalRule:
		r.cr()
		r.element.HR()
		r.cr()
	case blackfriday.List:
		if entering {
			r.openTag(renderer.TagList)
		} else {
			r.closeTag()
			if node.Parent.Type == blackfriday.Item && node.Next != nil {
				r.cr()
			}
			if node.Parent.Type == blackfriday.Document || node.Parent.Type == blackfriday.BlockQuote {
				r.cr()
			}
		}
	case blackfriday.Item:
		if entering {
			list_item_type := renderer.TagListItemUnordered
			if node.ListFlags&blackfriday.ListTypeOrdered != 0 {
				list_item_type = renderer.TagListItemOrdered

			} else if node.ListFlags&blackfriday.ListTypeTerm != 0 {
				list_item_type = renderer.TagListItemTerm
			} else if node.ListFlags&blackfriday.ListTypeDefinition != 0 {
				list_item_type = renderer.TagListItemDefinition
			}
			r.openTag(list_item_type)
		} else {
			r.closeTag()
		}
	case blackfriday.CodeBlock:
		r.cr()
		r.openTag(renderer.TagCodeblock)
		r.element.(*_Codeblock).lang = string(node.CodeBlockData.Info)
		r.Literal(string(node.Literal))
		r.closeTag()
		r.cr()
	case blackfriday.Table:
		if entering {
			r.cr()
			r.openTag(renderer.TagTable)
		} else {
			r.closeTag()
			r.cr()
		}
	case blackfriday.TableCell:
		if entering {
			r.openTag(renderer.TagTableCell)
			r.element.(*_TableCell).setAlignment(cellAlignment(node.Align))
		} else {
			r.closeTag()
		}
	case blackfriday.TableHead:
		if entering {
			r.openTag(renderer.TagTableHead)
		} else {
			r.closeTag()
		}
	case blackfriday.TableBody:
		if entering {
			r.openTag(renderer.TagTableBody)
		} else {
			r.closeTag()
		}
	case blackfriday.TableRow:
		if entering {
			r.openTag(renderer.TagTableRow)
		} else {
			r.closeTag()
		}
	default:
		panic("Unknown node type " + node.Type.String())
	}
	return blackfriday.GoToNext
}

func TerminalRenderer(flags int) blackfriday.Renderer {
	dir := "."
	if _, filename, _, ok := runtime.Caller(0); ok {
		dir = path.Dir(filename)
	}

	// load highlight style
	r, err := os.Open(filepath.Join(dir, "code-block.xml"))
	if err != nil {
		panic(err)
	}

	style, err := chroma.NewXMLStyle(r)
	if err != nil {
		panic(err)
	}
	styles.Register(style)

	return &_Terminal{
		width: terminalWidth(),
	}

}

func terminalWidth() int {
	size, _ := ts.GetSize()
	if size.Col() == 0 {
		return 80
	}
	return size.Col()
}

func cellAlignment(align blackfriday.CellAlignFlags) text.Align {
	switch align {
	case blackfriday.TableAlignmentLeft:
		return text.AlignLeft
	case blackfriday.TableAlignmentRight:
		return text.AlignRight
	case blackfriday.TableAlignmentCenter:
		return text.AlignCenter
	default:
		return text.AlignDefault
	}
}

func skipParagraphTags(node *blackfriday.Node) bool {
	grandparent := node.Parent.Parent
	if grandparent == nil || grandparent.Type != blackfriday.List {
		return false
	}
	tightOrTerm := grandparent.Tight || node.Parent.ListFlags&blackfriday.ListTypeTerm != 0
	return grandparent.Type == blackfriday.List && tightOrTerm
}
