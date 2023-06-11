package gocli

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"text/template"

	"github.com/ez-leka/gocli/i18n"
	"golang.org/x/text/language"
)

var templateManager *TemplateManager

type TemplateManager struct {
	CustomFuncs template.FuncMap
	localizer   i18n.Localizer
}

type IElementContext interface {
	GetType() string
}
type TokenTemplateContext struct {
	Name  string
	Extra string
}

func (f TokenTemplateContext) GetType() string {
	return "token"
}

type FlagTemplateContext struct {
	Name   string
	Short  rune
	Value  string
	Extra  string
	Prefix string
}

func (f FlagTemplateContext) GetType() string {
	return "flag"
}

type ArgTemplateContext struct {
	Name  string
	Extra string
}

func (f ArgTemplateContext) GetType() string {
	return "argument"
}

type UsageTemplateContext struct {
	AppName        string
	CurrentCommand Command
	Flags          []IFlagArg
	Args           []IFlagArg
}

type testRenderer struct {
	buf bytes.Buffer
}

func (f *testRenderer) Arg(i int) interface{} { return nil }
func (f *testRenderer) Render(s string)       { f.buf.WriteString(s) }

var (
	indent  = 4
	padding = 4
)

func NewTemplateManager(lang language.Tag) (*TemplateManager, error) {

	tm := &TemplateManager{}

	tm.CustomFuncs = template.FuncMap{
		"Translate":             tplTranslate,
		"Indent":                tplIndent,
		"ToUpper":               strings.ToUpper,
		"ToLower":               strings.ToLower,
		"Rune":                  tplRune,
		"IsFlag":                tplIsFlag,
		"IsArg":                 tplIsArg,
		"TwoColumns":            tplTwoColumns,
		"FlagsArgsToTwoColumns": tplFlagsArgsToTwoColumns,
		"CommandCategories":     tplCommandCategories,
		"CommandsToTwoColumns":  tplCommandsToTwoColumns,
		"FormatTemplate":        tplFormatTemplate,
	}
	fallback_lang := language.MustParse("en_us")
	tm.localizer = *i18n.NewLocalizer(lang, fallback_lang)

	tm.localizer.AddUpdateTranslation(fallback_lang, GoCliStrings)

	return tm, nil
}

func (t TemplateManager) AddTranslation(lang language.Tag, entries i18n.Entries) {
	t.localizer.AddUpdateTranslation(lang, entries)
}
func (t TemplateManager) AddFunction(name string, function any) {
	t.CustomFuncs[name] = function
}

func (t TemplateManager) makeError(key string, obj interface{}) error {
	buf := bytes.NewBuffer(nil)

	template_str := t.GetMessage(key)
	templ := template.Must(template.New("err").Funcs(t.CustomFuncs).Parse(template_str))
	templ.Execute(buf, obj)
	return errors.New(buf.String())
}

func (t TemplateManager) GetMessage(key string, a ...interface{}) string {

	s := t.localizer.Sprintf(key, a...)
	return s
}

func (t TemplateManager) formatTemplate(writer io.Writer, tpl string, obj any) error {
	tpl_content := t.GetMessage(tpl, obj)
	return t.doFormatTemplate(writer, tpl_content, obj, 0)
}

func (t TemplateManager) doFormatTemplate(writer io.Writer, tpl string, obj any, indent int) error {

	// pre-format
	// remove leading new lines
	tpl = t.fixTemplateAlignment(tpl, indent)

	templ, err := template.New("temp_tpl").Funcs(t.CustomFuncs).Parse(tpl)
	if err != nil {
		return err
	}
	return templ.Execute(writer, obj)
}

func (t TemplateManager) fixTemplateAlignment(tpl string, min_indent int) string {
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
		if i == 0 {
			if len(trimmed) == 0 {
				// empty first string - do not calculate first indent
				continue
			}
			first_indent = indent
		}
		s = strings.Repeat(" ", min_indent+indent-first_indent) + trimmed
		if i != 0 {
			new_tpl += "\n"
		}
		new_tpl += s
		i++
	}

	return new_tpl
}
