package gocli

import (
	"strings"

	"github.com/ez-leka/gocli/i18n"
	"golang.org/x/exp/slices"
)

// context holds the current context of the parser.
type context struct {
	CurrentCommand   *Command
	mixArgsAndFlags  bool
	argsOnly         bool
	noCommands       bool
	cli_args         []string
	arg_pos          int // Cursor into arguments - arguments are positioned, so they ar eprocessed in the order they arrive
	flags_lookup     map[string]IFlag
	arguments_lookup []IArg // arguments are positioned so array , not a map
	level            int    // depth of sub-command chain
}

func (ctx *context) nextArg() IArg {
	if ctx.arg_pos >= len(ctx.arguments_lookup) {
		// no more aruments possible
		return nil
	}
	arg := ctx.arguments_lookup[ctx.arg_pos]
	// we stay on last argument to allow to consume remainder if there are arguments left and this last argument is cumulative
	// once we reached end of command line arguments we need to allow for arg_pos to increase to set defaults
	if ctx.arg_pos < len(ctx.arguments_lookup)-1 || len(ctx.cli_args) == 0 || !arg.IsCumulative() {
		ctx.arg_pos++
	}
	return arg
}

func (ctx *context) mergeFlags(flags []IFlag) error {

	// remove all flags that do not belong to the group of current command
	current_cmd_groups := ctx.CurrentCommand.GetValidationGroups()
	if len(current_cmd_groups) > 0 {
		// current connamd has validation  group restrictions
		for fname, f := range ctx.flags_lookup {
			groups := f.GetValidationGroups()
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
				// remove from map
				delete(ctx.flags_lookup, fname)
			}
		}
	}
	// append flags specific to current command
	for _, flag := range flags {
		if _, ok := ctx.flags_lookup[flag.GetName()]; ok {
			return i18n.NewError("FlagLongExistsTemplate", flag)
		}
		ctx.flags_lookup[flag.GetName()] = flag

		// of short flag requested - add it too
		if flag.GetShort() != 0 {
			if _, ok := ctx.flags_lookup[string(flag.GetShort())]; ok {
				return i18n.NewError("FlagShortExistsTemplate", flag)
			}
			ctx.flags_lookup[string(flag.GetShort())] = flag
		}
		flag.SetLevel(ctx.level)
	}

	return nil
}

func (ctx *context) mergeArgs(args []IArg) error {
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

func (ctx *context) popCliArg() (string, bool) {
	if len(ctx.cli_args) > 0 {
		// pop next command line argument for consideration
		token := ctx.cli_args[0]
		ctx.cli_args = ctx.cli_args[1:]

		return token, true
	}
	return "", false
}

func (ctx *context) parse(app *Application, args []string) error {

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
	// the very first command is app itself
	ctx.mixArgsAndFlags = app.MixArgsAndFlags
	ctx.cli_args = args
	ctx.CurrentCommand = &app.Command
	err = ctx.mergeFlags(app.Flags)
	if err != nil {
		return err
	}
	ctx.mergeArgs(app.Args)
	ctx.updateCommandValidatables()

	for token, ok := ctx.popCliArg(); ok; token, ok = ctx.popCliArg() {
		// when debugging in VSCODE with input variable
		// we can end up with empty strings in args array (does not happen on real command line)
		// so we just going to ignore empty strings in args array
		if len(token) == 0 {
			continue
		}
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

func (ctx *context) resolveCompletion(app *Application, args []string) []string {

	completions := make([]string, 0)

	var (
		currArg string
		prevArg string
	)

	num_args := len(args)
	if num_args > 1 {
		currArg = args[len(args)-1]
	}
	if num_args > 2 {
		prevArg = args[len(args)-2]
	}

	// check if flag started - skip if special completion flag
	if currArg != "" && strings.HasPrefix(currArg, "--") {
		if ctx.argsOnly {
			return nil
		}
		flag_pref := currArg[2:]
		for _, flag := range ctx.CurrentCommand.Flags {
			if !flag.IsSetByUser() || flag.IsCumulative() {
				// if fully matched do options if any
				if flag.GetName() == flag_pref {
					completions = append(completions, flag.GetHints()...)
					return completions
				}
				// check if partially match - still loking for a flag name
				if strings.HasPrefix(flag.GetName(), flag_pref) {
					completions = append(completions, "--"+flag.GetName())
				}
			}
		}
		return completions
	}

	// prev argument was a flag and we are trying to get complition for possible values (if has hints)
	prev_name := prevArg[2:]
	if strings.HasPrefix(prevArg, "--") && prev_name != app.bashCompletionFlag.GetName() {
		if flag, ok := ctx.flags_lookup[prev_name]; ok {
			for _, hint := range flag.GetHints() {
				if strings.HasPrefix(hint, currArg) {
					completions = append(completions, hint)
				}
			}
			return completions
		}
	}

	for _, subc := range ctx.CurrentCommand.Commands {
		completions = append(completions, subc.Name)
	}

	for _, arg := range ctx.CurrentCommand.Args {
		if !arg.IsSetByUser() || arg.IsCumulative() {
			for _, hint := range arg.GetHints() {
				if strings.HasPrefix(hint, currArg) {
					completions = append(completions, hint)
				}
			}
		}
	}

	for _, flag := range ctx.CurrentCommand.Flags {
		if !flag.IsSetByUser() || flag.IsCumulative() {
			completions = append(completions, "--"+flag.GetName())
		}
	}
	return completions
}

func (ctx *context) processArg(token string) error {

	if cmd, ok := ctx.CurrentCommand.commands_map[token]; ok && !ctx.noCommands {
		// this is command

		ctx.CurrentCommand = cmd
		ctx.level++
		ctx.mergeArgs(cmd.Args)
		err := ctx.mergeFlags(cmd.Flags)
		ctx.updateCommandValidatables()
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
			return i18n.NewError("ExtraArgument", TokenTemplateContext{Extra: token})
		}
		err := arg.SetValue(token)
		if err != nil {
			return err
		}
		return nil
	}

	return i18n.NewError("UnexpectedTokenTemplate", TokenTemplateContext{Name: token, Extra: "command"})

}

func (ctx *context) processLongFlag(flag_token string) error {

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

func (ctx *context) processShortFlag(flag_token string) error {

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
func (ctx *context) validateGrouping(set map[string]IValidatable) ([]IValidatable, error) {

	grouped := ctx.CurrentCommand.GetGroupedFlagsAndArgs()

	// make sure  that only one group have set values
	// because the same element can be in more then one group we need to keep track of set elements
	set_elements := make(map[IValidatable]string)
	set_group_name := ""
	for gname, g := range grouped.Groups {
		if g.Command != "" && g.Command != ctx.CurrentCommand.Name {
			// this is group for subcommand and that subcommand is not currentn command - needed for usage but ignored for validation
			continue
		}
		for _, v := range mergeValidatables(g) {
			// check if already set
			_, ok := set_elements[v]
			if !ok && v.IsSetByUser() && (set_group_name != "" && set_group_name != gname) {
				var last_set IValidatable
				for other_v, other_gname := range set_elements {
					if other_gname != gname {
						last_set = other_v
						break
					}
				}
				return []IValidatable{}, i18n.NewError("FlagsArgsFromMultipleGroups", TokenTemplateContext{Name: v.GetPlaceholder(), Extra: last_set.GetName()})
			}
			if v.IsSetByUser() {
				set_elements[v] = gname
				set_group_name = gname
			}
		}
	}

	validate_set := make([]IValidatable, 0)
	if set_group_name != "" {
		validate_set = mergeValidatables(grouped.Groups[set_group_name])
	} else {
		// add all groups
		for _, g := range grouped.Groups {
			validate_set = append(validate_set, mergeValidatables(g)...)
		}
	}
	// append all ungrouped elements
	validate_set = append(validate_set, mergeValidatables(grouped.Ungrouped)...)
	return validate_set, nil

}
func (ctx *context) updateCommandValidatables() {

	ctx.CurrentCommand.setByUser = true
	ctx.CurrentCommand.level = ctx.level

	// create a single map of things to validate
	ctx.CurrentCommand.validatables = make(map[string]IValidatable)
	for name, f := range ctx.flags_lookup {
		if len(name) == 1 {
			// skip short flags because there is always a long one for it
			continue
		}
		ctx.CurrentCommand.validatables[name] = f
	}
	// add all arguments usign arg name
	for _, a := range ctx.arguments_lookup {
		ctx.CurrentCommand.validatables["arg_"+a.GetName()] = a
	}

	// add sub-command to validatable set
	for _, sc := range ctx.CurrentCommand.Commands {
		ctx.CurrentCommand.validatables["cmd"+sc.GetName()] = sc
	}
}

func (ctx *context) validate(app *Application) error {

	var err error

	// we also show help if last parsed command was not the leaf of the command chain
	if !ctx.CurrentCommand.isLeaf() {
		return i18n.NewError("CommandRequired", ctx.CurrentCommand)
	}

	validation_group, err := ctx.validateGrouping(ctx.CurrentCommand.validatables)
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

func (ctx *context) execute(app *Application) error {
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
