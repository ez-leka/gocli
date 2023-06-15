package gocli

import (
	"strings"

	"github.com/ez-leka/gocli/i18n"
	"golang.org/x/exp/slices"
)

// ParseContext holds the current context of the parser.
type ParseContext struct {
	CurrentCommand   *Command
	mixArgsAndFlags  bool
	argsOnly         bool
	noCommands       bool
	cli_args         []string
	flags_lookup     map[string]IFlag
	arguments_lookup []IArg // arguments are positioned so array , not a map
	arg_pos          int    // Cursor into arguments - arguments are positioned, so they ar eprocessed in the order they arrive
}

func (ctx *ParseContext) nextArg() IArg {
	if ctx.arg_pos >= len(ctx.arguments_lookup) {
		// no more aruments possible
		return nil
	}
	arg := ctx.arguments_lookup[ctx.arg_pos]
	// we stay on last argument to allow to consume remainder if there are arguments left
	// once we reached end of command line arguments we need to allow for arg_pos to increase to set defaults
	if ctx.arg_pos < len(ctx.arguments_lookup)-1 || len(ctx.cli_args) == 0 {
		ctx.arg_pos++
	}
	return arg
}

func (ctx *ParseContext) mergeFlags(flags []IFlag) error {
	for _, flag := range flags {
		if _, ok := ctx.flags_lookup[flag.GetName()]; ok {
			return i18n.NewError("FlagLongExistsTemplate", flag)
		}
		if _, ok := ctx.flags_lookup[string(flag.GetShort())]; ok {
			return i18n.NewError("FlagShortExistsTemplate", flag)
		}
		ctx.flags_lookup[flag.GetName()] = flag
		ctx.flags_lookup[string(flag.GetShort())] = flag
	}

	return nil
}

func (ctx *ParseContext) mergeArgs(args []IArg) error {
	// because arguments are positined we have to remove all arguments that do not belong to the group of current command
	current_cmd_groups := ctx.CurrentCommand.GetValidationGroups()
	if len(current_cmd_groups) > 0 {
		// current connamd has validation  group restrictions
		for i := 0; i < len(ctx.arguments_lookup); i++ {
			groups := ctx.arguments_lookup[i].GetValidationGroups()
			if len(groups) == 0 {
				//ungrouped arg, belongs to every group - leave it
				continue
			}
			in_group := false
			for _, g := range current_cmd_groups {
				idx := slices.Index(groups, g)
				if idx != -1 {
					in_group = true
					break
				}
			}
			if !in_group {
				ctx.arguments_lookup = append(ctx.arguments_lookup[:i], ctx.arguments_lookup[i+1:]...)
				i--
			}
		}
	}
	// add command specific arguments
	ctx.arguments_lookup = append(ctx.arguments_lookup, args...)

	return nil
}

func (ctx *ParseContext) popCliArg() (string, bool) {
	if len(ctx.cli_args) > 0 {
		// pop next command line argument for consideration
		token := ctx.cli_args[0]
		ctx.cli_args = ctx.cli_args[1:]

		return token, true
	}
	return "", false
}

func (ctx *ParseContext) parse(app *Application, args []string) error {

	var err error
	// reset context
	ctx.argsOnly = false
	ctx.noCommands = false
	ctx.arg_pos = 0
	// crear out all flags and args - should only bee needed if Run is called muptiple times
	for _, a := range ctx.arguments_lookup {
		a.Clear()
	}
	// run validators on all flags
	for _, f := range ctx.flags_lookup {
		f.Clear()
	}

	ctx.flags_lookup = make(map[string]IFlag, 0)
	ctx.arguments_lookup = make([]IArg, 0)

	// initiaze context
	ctx.mixArgsAndFlags = app.MixArgsAndFlags
	ctx.cli_args = args
	ctx.CurrentCommand = &app.Command
	err = ctx.mergeFlags(app.Flags)
	if err != nil {
		return err
	}
	ctx.mergeArgs(app.Args)

	for token, ok := ctx.popCliArg(); ok; token, ok = ctx.popCliArg() {
		if ctx.argsOnly || token == "-" || token == "--" {
			// uncoditional arg
			err = ctx.processArg(token)
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(token, "--") {
			err = ctx.processLongFlag(token)
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(token, "-") {
			err = ctx.processShortFlag(token)
			if err != nil {
				return err
			}
		} else {
			// some other argument or command
			err = ctx.processArg(token)
			if err != nil {
				return err
			}
		}

	}

	// Set defaults for all flags that are not set by user and have a default value
	// Note: using internal function so SetByUser is not set
	for _, f := range ctx.flags_lookup {
		if !f.IsSetByUser() && f.GetDefault() != "" {
			setFlagArgValue(f, f.GetDefault())
		}
	}
	// Set defaults for all remaining arguments
	for arg := ctx.nextArg(); arg != nil; arg = ctx.nextArg() {
		if !arg.IsSetByUser() && arg.GetDefault() != "" {
			setFlagArgValue(arg, arg.GetDefault())
		}
	}

	return nil
}
func (ctx *ParseContext) processArg(token string) error {

	if cmd, ok := ctx.CurrentCommand.commands_map[token]; ok && !ctx.noCommands {
		// this is command

		ctx.CurrentCommand = cmd
		ctx.mergeArgs(cmd.Args)
		err := ctx.mergeFlags(cmd.Flags)
		if err != nil {
			return err
		}

		return nil
	}
	// was not a sub-command 0check for argument
	if len(ctx.arguments_lookup) > 0 {
		// we got argument - no more commands
		ctx.noCommands = true
		if !ctx.mixArgsAndFlags {
			// no more flags
			ctx.argsOnly = true
		}
		arg := ctx.nextArg()
		if arg == nil {
			return i18n.NewError("UnknownArgument", TokenTemplateContext{Name: token})
		}
		err := arg.SetValue(token)
		if err != nil {
			return err
		}
		return nil
	}

	return i18n.NewError("UnexpectedTokenTemplate", TokenTemplateContext{Name: token, Extra: "command"})

}

func (ctx *ParseContext) processLongFlag(flag_token string) error {

	flag_token = flag_token[2:]

	// long flag can be of 3 types
	// --flag - it is a bool flag
	// --flag=value
	// --flag value
	// first isolate actual flag name
	flag_parts := strings.SplitN(flag_token, "=", 2)
	flag_name := flag_parts[0]

	// find flag
	flag, ok := ctx.flags_lookup[flag_name]
	if !ok {
		return i18n.NewError("UnknownElementTemplate", ElementTemplateContext{Element: &Flag[String]{Name: flag_name}})
	}

	// figure out flag value
	var flag_value string
	if flag.IsBool() {
		flag_value = "true"
	} else {
		if len(flag_parts) == 2 {
			// value was assigned via =
			flag_value = flag_parts[1]
		} else {
			// flag value must be next cli argument
			flag_value, ok = ctx.popCliArg()
			if !ok {
				i18n.NewError("UnexpectedFlagValueTemplate", ElementTemplateContext{Element: flag, Extra: flag_value})
			}
		}
	}
	err := flag.SetValue(flag_value)
	if err != nil {
		return i18n.NewError("FlagValidationFailed", ElementTemplateContext{Element: flag, Extra: flag_value})
	}

	return nil
}

func (ctx *ParseContext) processShortFlag(flag_token string) error {

	flag_token = flag_token[1:]

	// short flags consist of runes and (potencially values)
	//-f test.txt , -ftest.txt , -f=test.txt , -cv, -cvftest.txt, -cvf test.txt where c and v are boolean flags are possibble combinations
	runes := []rune(flag_token)
	for i, r := range runes {
		flag, ok := ctx.flags_lookup[string(r)]
		if ok {
			// this rune is a flag
			if flag.IsBool() {
				flag.SetValue("true")
				continue
			} else {
				//we have non-boolean flag se we need a value
				var flag_value string
				if len(runes) > i+1 {
					//the rest of the runes are flag's value, but there maybe = between flag and value
					flag_value = string(runes[i+1:])
					flag_value = strings.TrimPrefix(flag_value, "=")
				} else {
					// next argument is a flag value
					flag_value, ok = ctx.popCliArg()
					if !ok {
						return i18n.NewError("UnexpectedFlagValueTemplate", ElementTemplateContext{Element: flag, Extra: flag_value})
					}
				}
				err := flag.SetValue(flag_value)
				if err != nil {
					return i18n.NewError("FlagValidationFailed", ElementTemplateContext{Element: flag, Extra: flag_value})
				}
				return nil
			}
		} else {
			return i18n.NewError("UnknownElementTemplate", ElementTemplateContext{Element: &Flag[String]{Name: string(r)}})
		}
	}

	return nil
}

// Group flags and arguments according to their validation group; ignore short flags
//
// only flags and arguments from one group can be set,  i.e groups are mutially exclusive
func (ctx *ParseContext) validateGrouping(set map[string]IValidatable) ([]IValidatable, error) {

	groups := make(map[string][]IValidatable)        // hold all objects in every group
	unique_groups := make(map[string][]IValidatable) // hold objects that only have one group, i.e objects that are mutually exclusive
	ungrouped := make([]IValidatable, 0)             // hols all ungrouped; they will be added to final group

	// group arguments, flags, and commands before validation
	for _, vo := range set {
		obj_groups := vo.GetValidationGroups()
		if len(obj_groups) == 0 {
			// ungrouped element
			ungrouped = append(ungrouped, vo)
		}
		for _, g_name := range obj_groups {
			g, ok := groups[g_name]
			if !ok {
				g = make([]IValidatable, 0)

			}
			g = append(g, vo)
			groups[g_name] = g
		}
		if len(obj_groups) == 1 {
			g_name := obj_groups[0]
			g, ok := unique_groups[g_name]
			if !ok {
				g = make([]IValidatable, 0)

			}
			g = append(g, vo)
			unique_groups[g_name] = g
		}
	}

	// make sure that no 2 emutially exclusive objects are set
	set_in_group := ""
	var va_set IValidatable = nil
	for g_name, g := range unique_groups {
		for _, vo := range g {
			if vo.IsSetByUser() {
				if set_in_group != "" && set_in_group != g_name {
					// have values set for more then one exclusive group
					return []IValidatable{}, i18n.NewError("FlagsArgsFromMultipleGroups", TokenTemplateContext{Name: va_set.GetName(), Extra: vo.GetName()})
				}
				va_set = vo
				set_in_group = g_name
			}
		}
	}
	var final_group []IValidatable
	if len(groups) > 1 && len(unique_groups) == 0 {
		// there ar emultiple named groups and no unique groups
		return []IValidatable{}, i18n.NewError("NoUniqueFlagArgCommandInGroup", nil)
	}

	if set_in_group == "" {
		// no group has a value set by user - most likely command has been unfinished
		// we will returns everything so it can be valudated for required
		for _, vo := range set {
			final_group = append(final_group, vo)
		}
		return final_group, nil
	}

	// we have unique group
	final_group = append(groups[set_in_group], ungrouped...)

	return final_group, nil
}
func (ctx *ParseContext) validate(app *Application) error {

	var err error

	// we also show help if last parsed command was not the leaf of the command chain
	if !ctx.CurrentCommand.isLeaf() {
		return i18n.NewError("CommandRequired", ctx.CurrentCommand)
	}

	// create a single map of things to validate
	validatables := make(map[string]IValidatable)
	for name, f := range ctx.flags_lookup {
		if len(name) == 1 {
			// skip short flags because ther is always a long one for it
			continue
		}
		validatables[name] = f
	}
	// add all arguments usign arg name
	for _, a := range ctx.arguments_lookup {
		validatables["arg_"+a.GetName()] = a
	}
	for p := ctx.CurrentCommand; p != nil; p = p.parent {
		p.setByUser = true
		validatables["cmd_"+p.GetName()] = p
	}

	validation_group, err := ctx.validateGrouping(validatables)
	if err != nil {
		return err
	}

	// call validators for flags and arguments
	for _, vo := range validation_group {
		if _, ok := vo.(ICommand); ok {
			// command is validated last
			continue
		}
		err = vo.ValidateWrapper(app)
		if err != nil {
			return err
		}
	}

	// validate group again, this time for required
	// we want to validate arguments before flags so we sort
	// we do not validate commands for required
	args := make([]IArg, 0)
	flags := make([]IFlag, 0)
	for _, vo := range validation_group {
		if f, ok := vo.(IFlag); ok {
			flags = append(flags, f)
		} else if arg, ok := vo.(IArg); ok {
			args = append(args, arg)
		}
	}
	for _, vo := range args {
		if vo.IsRequired() && !vo.IsSetByUser() {
			return i18n.NewError("MissingRequiredArg", vo)
		}
	}
	for _, vo := range flags {
		if vo.IsRequired() && !vo.IsSetByUser() {
			return i18n.NewError("MissingRequiredFlag", vo)
		}
	}

	// finally call validator for commands
	cmd := ctx.CurrentCommand
	for cmd != nil {
		err := cmd.ValidateWrapper(app)
		if err != nil {
			return err
		}
		cmd = cmd.parent
	}
	return nil
}

func (ctx *ParseContext) execute(app *Application) error {
	var data interface{} = nil
	var err error
	cmd := ctx.CurrentCommand
	for cmd != nil {
		data, err = cmd.ActionWrapper(app, data)
		if err != nil {
			return err
		}
		if app.stopActionPropagation {
			break
		}
		cmd = cmd.parent
	}
	return nil
}
