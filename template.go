package gocli

import (
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
	templ, err := template.New("temp_tpl").Funcs(t.CustomFuncs).Parse(tpl_content)
	if err != nil {
		return err
	}
	return templ.Execute(writer, obj)
}
