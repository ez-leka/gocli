package gocli

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/ez-leka/gocli/i18n"
	"golang.org/x/text/language"
)

var templateManager *TemplateManager

type TemplateManager struct {
	CustomFuncs template.FuncMap
	localizer   *i18n.Localizer
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
	AppName        string
	CurrentCommand Command
	Flags          []IFlagArg
	Args           []IFlagArg
}

var (
	indent  = 4
	padding = 4
)

func initTemplateManager() {

	default_lang := language.MustParse("en_us")
	// tf := TemplateFuncs{}
	templateManager = &TemplateManager{
		CustomFuncs: template.FuncMap{
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
		},
		localizer: i18n.NewLocalizer(default_lang, default_lang),
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

func (t TemplateManager) FormatTemplate(writer io.Writer, tpl string, obj any) error {
	tpl_content := t.GetLocalizedString(tpl, obj)
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
