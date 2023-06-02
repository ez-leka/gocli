package gocli

import (
	"strings"
)

type Action func(*Application, *Command) error
type CommandValidator func(*Application, *Command) error

type Command struct {
	Name         string
	Alias        []string
	Description  string
	Usage        string
	Category     *CommandCategory
	Flags        []IFlag
	Args         []IArg
	Commands     []*Command
	Action       Action
	Validator    CommandValidator
	initialized  bool
	commands_map map[string]*Command
	parent       *Command
}

func (c Command) FullCommand() string {
	full_cmd := []string{c.Name}
	for p := c.parent; p != nil; p = p.parent {
		full_cmd = append([]string{p.Name}, full_cmd...)
	}
	return strings.Join(full_cmd, " ")

}

// all required flags will be first  followed by all optional flags
func (c Command) GetGlobalFlags() ValidationGroup {
	global := ValidationGroup{
		RequiredFlags: make([]IFlag, 0),
		OptionalFlags: make([]IFlag, 0),
		RequiredArgs:  make([]IArg, 0),
		OptionalArgs:  make([]IArg, 0),
	}
	for p := c.parent; p != nil; p = p.parent {
		for _, f := range p.Flags {
			if f.IsRequired() {
				global.RequiredFlags = append(global.RequiredFlags, f)
			} else {
				global.OptionalFlags = append(global.OptionalFlags, f)
			}
		}
	}
	return global

}

// all required flags will be first  followed by all optional flags in every group
// followed by all required args followed by all optional args
func (c Command) GetValidationGroups() GroupedFlagsArgs {

	grouped := GroupedFlagsArgs{
		Ungrouped: ValidationGroup{
			RequiredFlags: make([]IFlag, 0),
			OptionalFlags: make([]IFlag, 0),
			RequiredArgs:  make([]IArg, 0),
			OptionalArgs:  make([]IArg, 0),
		},
		Groups: make(map[string]ValidationGroup, 0),
	}

	for _, fa := range c.Flags {
		if len(fa.GetValidationGroups()) == 0 {
			if fa.IsRequired() {
				grouped.Ungrouped.RequiredFlags = append(grouped.Ungrouped.RequiredFlags, fa)
			} else {
				grouped.Ungrouped.OptionalFlags = append(grouped.Ungrouped.OptionalFlags, fa)
			}
			continue
		}
		for _, gname := range fa.GetValidationGroups() {
			var g ValidationGroup
			ok := false
			if g, ok = grouped.Groups[gname]; !ok {
				g = ValidationGroup{
					RequiredFlags: make([]IFlag, 0),
					OptionalFlags: make([]IFlag, 0),
					RequiredArgs:  make([]IArg, 0),
					OptionalArgs:  make([]IArg, 0),
				}
			}
			if fa.IsRequired() {
				g.RequiredFlags = append(g.RequiredFlags, fa)
			} else {
				g.OptionalFlags = append(g.OptionalFlags, fa)
			}
			grouped.Groups[gname] = g
		}
	}
	for _, fa := range c.Args {
		if len(fa.GetValidationGroups()) == 0 {
			if fa.IsRequired() {
				grouped.Ungrouped.RequiredArgs = append(grouped.Ungrouped.RequiredArgs, fa)
			} else {
				grouped.Ungrouped.OptionalArgs = append(grouped.Ungrouped.OptionalArgs, fa)
			}
			continue
		}
		for _, gname := range fa.GetValidationGroups() {
			var g ValidationGroup
			ok := false
			if g, ok = grouped.Groups[gname]; !ok {
				g = ValidationGroup{
					RequiredFlags: make([]IFlag, 0),
					OptionalFlags: make([]IFlag, 0),
					RequiredArgs:  make([]IArg, 0),
					OptionalArgs:  make([]IArg, 0),
				}

			}
			if fa.IsRequired() {
				g.RequiredArgs = append(g.RequiredArgs, fa)
			} else {
				g.OptionalArgs = append(g.OptionalArgs, fa)
			}
			grouped.Groups[gname] = g

		}
	}
	return grouped
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

func (c *Command) AddFlag(flag IFlag) {
	c.Flags = append(c.Flags, flag)
}

func (c *Command) AddArg(arg IArg) {
	c.Args = append(c.Args, arg)
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

func (c *Command) ActionWrapper(app *Application) error {
	if c.Action != nil {
		return c.Action(app, c)
	}
	return nil
}
