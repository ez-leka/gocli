package gocli

import (
	"reflect"
	"strings"

	"golang.org/x/exp/slices"
)

type Action func(*Application, *Command, interface{}) (interface{}, error)
type CommandValidator func(*Application, *Command) error

type Command struct {
	Name             string
	Alias            []string
	Description      string
	Usage            string
	Category         *CommandCategory
	Flags            []IFlag
	Args             []IArg
	Commands         []*Command
	Action           Action
	Validator        CommandValidator
	ValidationGroups []string
	Optional         bool
	Hidden           bool // can be used on command line but will not show on help
	initialized      bool
	commands_map     map[string]*Command
	parent           *Command
	setByUser        bool
	validatables     map[string]IValidatable
}

func (c Command) FullCommand() string {
	full_cmd := []string{c.Name}
	for p := c.parent; p != nil; p = p.parent {
		full_cmd = append([]string{p.Name}, full_cmd...)
	}
	return strings.Join(full_cmd, " ")

}

func (c *Command) GetName() string {
	return c.Name
}

func (c *Command) GetPlaceholder() string {
	return c.Name
}

func (c *Command) GetType() string {
	return "command"
}

func (c *Command) IsHidden() bool {
	return c.Hidden
}

func (c *Command) IsRequired() bool {
	return true // if command was parced it was required
}

// all required flags will be first  followed by all optional flags in every group
// followed by all required args followed by all optional args
func (c Command) GetGroupedFlagsAndArgs() groupedFlagsArgs {

	grouped := groupedFlagsArgs{
		Ungrouped: validationGroup{
			requiredFlags: make([]IValidatable, 0),
			optionalFlags: make([]IValidatable, 0),
			requiredArgs:  make([]IValidatable, 0),
			optionalArgs:  make([]IValidatable, 0),
		},
		Groups: make(map[string]validationGroup, 0),
	}

	flags := make([]IFlag, 0)
	args := make([]IArg, 0)
	sub_cmds := make([]ICommand, 0)
	for _, v := range c.validatables {
		if f, ok := v.(IFlag); ok {
			flags = append(flags, f)
			continue
		}
		if a, ok := v.(IArg); ok {
			args = append(args, a)
			continue
		}
		if sc, ok := v.(ICommand); ok {
			sub_cmds = append(sub_cmds, sc)
			continue
		}

	}

	for _, fa := range flags {
		if len(fa.GetValidationGroups()) == 0 {
			if fa.IsRequired() {
				grouped.Ungrouped.requiredFlags = append(grouped.Ungrouped.requiredFlags, fa)
			} else {
				grouped.Ungrouped.optionalFlags = append(grouped.Ungrouped.optionalFlags, fa)
			}
			continue
		}
		for _, gname := range fa.GetValidationGroups() {
			var g validationGroup
			ok := false
			if g, ok = grouped.Groups[gname]; !ok {
				g = validationGroup{
					requiredFlags: make([]IValidatable, 0),
					optionalFlags: make([]IValidatable, 0),
					requiredArgs:  make([]IValidatable, 0),
					optionalArgs:  make([]IValidatable, 0),
				}
			}
			if fa.IsRequired() {
				g.requiredFlags = append(g.requiredFlags, fa)
			} else {
				g.optionalFlags = append(g.optionalFlags, fa)
			}
			grouped.Groups[gname] = g
		}
	}

	for _, fa := range args {
		if len(fa.GetValidationGroups()) == 0 {
			if fa.IsRequired() {
				grouped.Ungrouped.requiredArgs = append(grouped.Ungrouped.requiredArgs, fa)
			} else {
				grouped.Ungrouped.optionalArgs = append(grouped.Ungrouped.optionalArgs, fa)
			}
			continue
		}
		for _, gname := range fa.GetValidationGroups() {
			var g validationGroup
			ok := false
			if g, ok = grouped.Groups[gname]; !ok {
				g = validationGroup{
					requiredFlags: make([]IValidatable, 0),
					optionalFlags: make([]IValidatable, 0),
					requiredArgs:  make([]IValidatable, 0),
					optionalArgs:  make([]IValidatable, 0),
				}

			}
			if fa.IsRequired() {
				g.requiredArgs = append(g.requiredArgs, fa)
			} else {
				g.optionalArgs = append(g.optionalArgs, fa)
			}
			grouped.Groups[gname] = g

		}
	}

	for _, subc := range sub_cmds {
		if len(subc.GetValidationGroups()) == 0 {
			grouped.Ungrouped.Command = "command"
			grouped.Ungrouped.IsGenericCommand = true
			continue
		}
		for _, gname := range subc.GetValidationGroups() {
			var g validationGroup
			ok := false
			if g, ok = grouped.Groups[gname]; !ok {
				g = validationGroup{
					Command:       subc.GetName(),
					requiredFlags: make([]IValidatable, 0),
					optionalFlags: make([]IValidatable, 0),
					requiredArgs:  make([]IValidatable, 0),
					optionalArgs:  make([]IValidatable, 0),
				}

			}
			g.Command = subc.GetName()
			grouped.Groups[gname] = g

		}
	}

	// remove duplicates in groups
	for gname, g := range grouped.Groups {
		for gname1, g1 := range grouped.Groups {
			if gname1 == gname {
				continue
			}
			if reflect.DeepEqual(g, g1) {
				delete(grouped.Groups, gname1)
			}
		}
	}

	// finally sort all flags by level
	for _, g := range grouped.Groups {
		slices.SortFunc(g.requiredFlags, validatableSorter)
		slices.SortFunc(g.optionalFlags, validatableSorter)
	}
	slices.SortFunc(grouped.Ungrouped.requiredFlags, validatableSorter)
	slices.SortFunc(grouped.Ungrouped.optionalFlags, validatableSorter)
	return grouped
}

func (c *Command) GetValidationGroups() []string {
	return c.ValidationGroups
}

func (c *Command) init() error {
	c.commands_map = make(map[string]*Command)

	for _, sub_c := range c.Commands {
		// add command by name and alliases to command map
		c.commands_map[sub_c.Name] = sub_c
		for _, alias := range sub_c.Alias {
			c.commands_map[alias] = sub_c
		}
		sub_c.parent = c
		_ = sub_c.init()
	}

	c.initialized = true

	return nil
}

func (c *Command) HasSubCommands() bool {
	return len(c.commands_map) > 0
}

func (c *Command) HasSubCommand(name string) bool {
	_, ok := c.commands_map[name]
	return ok
}

func (c *Command) isLeaf() bool {
	leaf := true
	// the command is a leaf if it has no sub-commands that are not optional
	for _, sc := range c.Commands {
		if !sc.Optional {
			leaf = false
			break
		}
	}
	return leaf
}

func (c *Command) IsSetByUser() bool {
	return c.setByUser
}

func (c *Command) AddFlag(flag IFlag) {
	c.Flags = append(c.Flags, flag)
}

func (c *Command) AddFlags(flags []IFlag) {
	c.Flags = append(c.Flags, flags...)
}

func (c *Command) AddArg(arg IArg) {
	c.Args = append(c.Args, arg)
}

func (c *Command) AddArgs(args []IArg) {
	c.Args = append(c.Args, args...)
}

func (c *Command) AddCommand(cmd Command) {
	c.Commands = append(c.Commands, &cmd)
}

func (c *Command) ValidateWrapper(app *Application) error {
	if c.Validator != nil {
		return c.Validator(app, c)
	}
	return nil
}

func (c *Command) ActionWrapper(app *Application, in_data interface{}) (interface{}, error) {
	var data interface{} = nil
	var err error = nil
	if c.Action != nil {
		data, err = c.Action(app, app.context.CurrentCommand, in_data)
	}
	return data, err
}
