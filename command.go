package gocli

import (
	"strings"
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
	initialized      bool
	commands_map     map[string]*Command
	parent           *Command
	setByUser        bool
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

func (c *Command) IsRequired() bool {
	return true // if command was parced it was required
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
func (c Command) GetGroupedFlagsAndArgs() GroupedFlagsArgs {

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

func (c *Command) ActionWrapper(app *Application, in_data interface{}) (interface{}, error) {
	var data interface{} = nil
	var err error = nil
	if c.Action != nil {
		data, err = c.Action(app, c, in_data)
	}
	return data, err
}
