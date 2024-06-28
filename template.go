package gocli

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/ez-leka/gocli/i18n"
	"github.com/ez-leka/gocli/renderer/manpage"
	"github.com/ez-leka/gocli/renderer/terminal"
	"github.com/russross/blackfriday/v2"
	"golang.org/x/text/language"
)

var templateManager *TemplateManager

type OutputFormat string

const (
	TemplateTerminal OutputFormat = "terminal"
	TemplateHTML     OutputFormat = "html"
	TemplateMarkdown OutputFormat = "markdown"
	TemplateManpage  OutputFormat = "manpage"
	TemplateText     OutputFormat = "text"
)

type TemplateManager struct {
	CustomFuncs  template.FuncMap
	localizer    *i18n.Localizer
	withTitle    string
	outputFormat OutputFormat
	css          string
	icon         string
	TOC          bool

	currentLevel int
}

type TokenTemplateContext struct {
	Name  string
	Extra string
}

type ElementTemplateContext struct {
	Element IValidatable
	Extra   string
}

type UsageTemplateContext struct {
	AppName           string
	CurrentCommand    Command
	Flags             []IFlagArg
	Args              []IFlagArg
	Level             int
	UseOptionsCommand bool
	DocGeneration     bool
}

func initTemplateManager() {

	default_lang := language.MustParse("en_us")
	templateManager = &TemplateManager{
		CustomFuncs: template.FuncMap{
			"Translate":             tplTranslate,
			"Dict":                  tplDict,
			"HLevel":                tplHeaderLevel,
			"BlockBracket":          tplBlockBracket,
			"ToUpper":               strings.ToUpper,
			"ToLower":               strings.ToLower,
			"Rune":                  tplRune,
			"IsFlag":                tplIsFlag,
			"IsArg":                 tplIsArg,
			"Synopsis":              tplSynopsys,
			"SynopsisFlag":          tplSynopsisFlag,
			"DefinitionList":        tplDefinitionList,
			"FlagsArgsToTwoColumns": tplFlagsArgsToTwoColumns,
			"CommandCategories":     tplCommandCategories,
			"CommandsToTwoColumns":  tplCommandsToTwoColumns,
			"FormatTemplate":        tplFormatTemplate,
		},
		localizer:    i18n.NewLocalizer(default_lang, default_lang),
		outputFormat: TemplateTerminal,
		css:          "",
		icon:         "",
		TOC:          false,
	}

	templateManager.localizer.AddUpdateTranslation(default_lang, GoCliStrings)
}

func (t TemplateManager) AddTranslation(lang language.Tag, entries i18n.Entries) {
	t.localizer.AddUpdateTranslation(lang, entries)
}

func (t TemplateManager) UpdateTranslation(lang language.Tag, key string, obj any) {
	t.localizer.AddUpdateTranslation(lang, i18n.Entries{key: obj})
}

func (t TemplateManager) AddFunction(name string, function any) {
	t.CustomFuncs[name] = function
}

func (t TemplateManager) GetLocalizedString(key string, a ...interface{}) string {

	s := t.localizer.Sprintf(key, a...)
	return s
}

type TemplateFormatOption func(f *TemplateManager)

func WithTitle(title string) TemplateFormatOption {
	return func(c *TemplateManager) {
		c.withTitle = title
	}
}

func WithOutput(output OutputFormat) TemplateFormatOption {
	return func(c *TemplateManager) {
		c.outputFormat = output
	}
}

func WithCSS(css string) TemplateFormatOption {
	return func(c *TemplateManager) {
		c.css = css
	}
}

func WithIcon(icon string) TemplateFormatOption {
	return func(c *TemplateManager) {
		c.icon = icon
	}
}

func WithTOC(toc bool) TemplateFormatOption {
	return func(c *TemplateManager) {
		c.TOC = toc
	}
}

func (t *TemplateManager) FormatTemplate(writer io.Writer, tpl string, obj any, opts ...TemplateFormatOption) error {

	buf := bytes.NewBuffer(nil)
	if err := t.doFormatTemplate(buf, tpl, obj); err != nil {
		return err
	}

	return t.generateTemplateOutput(writer, buf, opts...)
}

func (t *TemplateManager) generateTemplateOutput(out io.Writer, from *bytes.Buffer, opts ...TemplateFormatOption) error {
	for _, opt := range opts {
		opt(t)
	}

	var output []byte

	switch t.outputFormat {
	case TemplateText:
		// TODO - output text. For now assume it is not a markdown
		output = from.Bytes()
	case TemplateMarkdown:
		output = from.Bytes()
	case TemplateHTML:
		params := blackfriday.HTMLRendererParameters{
			Title: t.withTitle,
			CSS:   t.css,
			Icon:  t.icon,
			Flags: blackfriday.CommonHTMLFlags | blackfriday.CompletePage,
		}
		if t.TOC {
			params.Flags |= blackfriday.TOC
		}
		renderer := blackfriday.NewHTMLRenderer(params)
		output = blackfriday.Run(from.Bytes(), blackfriday.WithRenderer(renderer))
	case TemplateTerminal:
		renderer := terminal.TerminalRenderer(0)
		output = blackfriday.Run(from.Bytes(), blackfriday.WithRenderer(renderer))
	case TemplateManpage:
		renderer := manpage.TRoffRenderer(t.withTitle)
		output = blackfriday.Run(from.Bytes(), blackfriday.WithRenderer(renderer))
	}

	if _, err := out.Write(output); err != nil {
		return err
	}

	return nil

}

func (t TemplateManager) doFormatTemplate(writer io.Writer, tpl string, obj any) error {

	// pre-format
	// remove leading new lines
	tpl = t.GetLocalizedString(tpl, obj)

	tpl = t.fixTemplateAlignment(tpl)

	templ, err := template.New("temp_tpl").Funcs(t.CustomFuncs).Parse(tpl)
	if err != nil {
		return err
	}
	// load all template defines we know about in case we using one
	for i, sub_t_key := range t.localizer.TemplatesDefined {
		sub_t := t.localizer.Sprintf(sub_t_key)
		sub_t = strings.Trim(sub_t, "\t \n")
		if _, err = templ.New(fmt.Sprint("_", i)).Parse(sub_t); err != nil {
			return err
		}
	}

	return templ.Execute(writer, obj)
}

func (t TemplateManager) fixTemplateAlignment(tpl string) string {
	min_indent := 0
	tpl = strings.TrimPrefix(tpl, "\n")
	tpl = strings.TrimSuffix(tpl, "\n")
	tpl = strings.ReplaceAll(tpl, "\t", "    ")

	new_tpl := ""
	scanner := bufio.NewScanner(strings.NewReader(tpl))
	i := 0
	first_indent := 0
	for scanner.Scan() {
		s := scanner.Text()
		trimmed := strings.TrimLeft(s, " \t")
		indent := len(s) - len(trimmed)
		real_indent := min_indent + indent - first_indent
		if len(trimmed) == 0 {
			if i == 0 {
				// first empty line - ignore
				continue
			} else {
				//empty line - no inden tneeded
				real_indent = 0
			}
		}
		if i == 0 {
			first_indent = indent
			real_indent = min_indent
		}
		s = strings.Repeat(" ", real_indent) + trimmed
		if i != 0 {
			new_tpl += "\n"
		}
		new_tpl += s
		i++
	}

	return new_tpl
}
