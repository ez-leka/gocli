package gocli

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/mitchellh/go-wordwrap"
	"github.com/olekukonko/ts"
)

func tplTranslate(in string) string {
	return templateManager.localizer.Sprintf(in)
}
func tplIndent(indent int, s string) string {
	s = strings.Repeat(" ", indent) + s
	return s
}
func tplRune(c rune) string {
	if c != 0 {
		return string(c)
	}
	return ""
}
func tplIsFlag(fa IFlagArg) bool {
	_, ok := fa.(IFlag)
	return ok
}
func tplIsArg(fa IFlagArg) bool {
	_, ok := fa.(IArg)
	return ok
}
func tplTwoColumns(rows [][2]string) string {
	return formatTwoColumns(rows)
}
func tplFlagsArgsToTwoColumns(flags_args []IFlagArg) [][2]string {
	rows := [][2]string{}
	var name string
	for _, fa := range flags_args {
		if f, ok := fa.(IFlag); ok {
			if f.GetShort() != 0 {
				name = templateManager.localizer.Sprintf("FormatFlagWithShort", f.GetShort(), f.GetName())
			} else {
				name = templateManager.localizer.Sprintf("FormatFlagNoShort", f.GetName())
			}
		} else {
			name = templateManager.localizer.Sprintf("FormatArg", fa.GetName())
		}

		// usage can be a template - so make it first
		buf := bytes.NewBuffer(nil)
		templateManager.formatTemplate(buf, fa.GetUsage(), fa)
		rows = append(rows, [2]string{name, buf.String()})
	}
	return rows
}
func tplCommandCategories(commands []*Command) []*CommandCategory {
	categories := make([]*CommandCategory, 0)

	misc_cat := CommandCategory{Name: templateManager.localizer.Sprintf("FormatMisCommandsCategory"), Order: 99, Commands: make([]*Command, 0)}
	categories = append(categories, &misc_cat)

	for _, cmd := range commands {
		if cmd.Category != nil {
			// see if category already listed
			found := false
			for _, cat := range categories {
				if cat.Name == cmd.Category.Name {
					found = true
					cat.Commands = append(cat.Commands, cmd)
					break
				}
			}
			if !found {
				categories = append(categories, cmd.Category)
				cmd.Category.Commands = make([]*Command, 0)
				cmd.Category.Commands = append(cmd.Category.Commands, cmd)
			}
		} else {
			misc_cat.Commands = append(misc_cat.Commands, cmd)
		}
	}
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Order < categories[j].Order
	})

	// clean up misc category
	if len(misc_cat.Commands) == 0 {
		// remove misc as there are no uncategorised commands
		categories = categories[0 : len(categories)-1]
	}
	if len(categories) == 1 {
		// we only have one category, it is is unnamed by user, name it Commands
		if categories[0] == &misc_cat {
			misc_cat.Name = templateManager.localizer.Sprintf("FormatCommandsCategory")
		}
	}

	return categories
}
func tplCommandsToTwoColumns(commands []*Command) [][2]string {

	rows := [][2]string{}
	for _, cmd := range commands {
		name := cmd.Name
		aliases := strings.Join(cmd.Alias, ",")
		if len(aliases) > 0 {
			name = name + "(" + aliases + ")"
		}

		rows = append(rows, [2]string{name, tplFormatTemplate(cmd.Description, cmd)})
	}
	return rows
}

func tplFormatTemplate(tpl string, obj any) string {
	buf := bytes.NewBuffer(nil)
	templ, err := template.New("temp_tpl").Funcs(templateManager.CustomFuncs).Parse(tpl)
	if err != nil {
		return err.Error()
	}
	templ.Execute(buf, obj)
	return buf.String()
}

func formatTwoColumns(rows [][2]string) string {
	result := ""
	width := terminalWidth()
	// calculate max width
	first_col_max_width := width/2 - indent - padding
	// Find size of first column.
	first_col_width := 0
	for _, row := range rows {
		if c := len(row[0]); c > first_col_width && c < first_col_max_width {
			first_col_width = c
		}
	}

	second_col_width := width - first_col_width - indent - padding

	format := fmt.Sprintf("%%-%ds%%-%ds%%-%ds%%-%ds\n", indent, first_col_width, padding, second_col_width)

	for _, row := range rows {
		col2_lines := strings.Split(wordwrap.WrapString(row[1], uint(second_col_width)), "\n")

		col1 := row[0]
		for _, col2 := range col2_lines {
			result += fmt.Sprintf(format, "", col1, "", col2)
			col1 = ""
		}
	}
	return result
}

func terminalWidth() int {
	size, _ := ts.GetSize()
	if size.Col() == 0 {
		return 80
	}
	return size.Col()
}
