package gocli

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/olekukonko/ts"
	"golang.org/x/exp/slices"
)

// generic template functions
func tplTranslate(in string) string {
	return templateManager.localizer.Sprintf(in)
}

func tplHeaderLevel(level int) string {
	return strings.Repeat("#", level+templateManager.currentLevel)
}

func tplBlockBracket() string {
	return "```"
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

func tplDict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid dict call")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

func tplSynopsisFlag(flag IFlag) string {
	flag_str := ""
	if flag.GetShort() != 0 {
		flag_str = fmt.Sprintf("-%c ", flag.GetShort())
	}
	flag_str += fmt.Sprintf("--%s", flag.GetName())
	return flag_str
}

func tplSynopsys(ctx UsageTemplateContext) string {
	synopsis := ctx.CurrentCommand.Name
	for p := ctx.CurrentCommand.parent; p != nil; p = p.parent {
		parent_synopsis := p.Name

		// add any parent required flags
		// and make a note of any optional
		has_optional_flags := false
		for _, f := range p.Flags {
			if f.IsRequired() {
				parent_synopsis += " " + tplSynopsisFlag(f)
			} else {
				has_optional_flags = true
			}
			
		}
		if has_optional_flags {
			parent_synopsis += tplTranslate("options")
		}
		synopsis = parent_synopsis + synopsis
	}

	return synopsis
}

func tplFlagsArgsToTwoColumns(flags_args []IFlagArg, level int) [][2]string {
	rows := [][2]string{}
	var name string

	// sort flags by level
	slices.SortFunc(flags_args, flagSorter)

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
		templateManager.doFormatTemplate(buf, fa.GetUsage(), fa)
		usage := buf.String()
		usage = strings.TrimRight(usage, " \t.")
		if len(fa.GetHints()) > 0 {
			usage += " " + templateManager.localizer.Sprintf("FormatHints", strings.Join(fa.GetHints(), ","))
		}
		if fa.GetDefault() != "" {
			usage += " " + templateManager.localizer.Sprintf("FormatDefault", fa.GetDefault())
		}
		rows = append(rows, [2]string{name, usage})
	}
	return rows
}
func tplCommandCategories(commands []*Command) []*CommandCategory {
	categories := make([]*CommandCategory, 0)

	misc_cat := CommandCategory{Name: templateManager.localizer.Sprintf("FormatMisCommandsCategory"), Order: 99, commands: make([]*Command, 0)}
	categories = append(categories, &misc_cat)

	for _, cmd := range commands {
		if cmd.IsHidden() {
			continue
		}
		if cmd.Category != nil {
			// see if category already listed
			found := false
			for _, cat := range categories {
				if cat.Name == cmd.Category.Name {
					found = true
					cat.commands = append(cat.commands, cmd)
					break
				}
			}
			if !found {
				categories = append(categories, cmd.Category)
				cmd.Category.commands = make([]*Command, 0)
				cmd.Category.commands = append(cmd.Category.commands, cmd)
			}
		} else {
			misc_cat.commands = append(misc_cat.commands, cmd)
		}
	}
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Order < categories[j].Order
	})

	// clean up misc category
	if len(misc_cat.commands) == 0 {
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

		usage := tplFormatTemplate(cmd.Description, cmd)
		// take first line only
		lines := strings.Split(usage, "\n")
		rows = append(rows, [2]string{name, lines[0]})
	}
	return rows
}

func tplFormatTemplate(tpl string, obj any) string {

	buf := bytes.NewBuffer(nil)
	templateManager.doFormatTemplate(buf, tpl, obj)
	return buf.String()
}

func tplDefinitionList(rows [][2]string) string {
	result := "\n"

	for _, row := range rows {
		result += fmt.Sprintf("**%s**\n: %s\n\n", row[0], row[1])
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
