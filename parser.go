package gocli

import (
	"strings"
)

// ParseContext holds the current context of the parser.
type ParseContext struct {
	CurrentCommand   *Command
	mixArgsAndFlags  bool
	argsOnly         bool
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
			return templateManager.makeError("FlagLongExistsTemplate", flag)
		}
		if _, ok := ctx.flags_lookup[string(flag.GetShort())]; ok {
			return templateManager.makeError("FlagShortExistsTemplate", flag)
		}
		ctx.flags_lookup[flag.GetName()] = flag
		ctx.flags_lookup[string(flag.GetShort())] = flag
	}
	return nil
}

func (ctx *ParseContext) mergeArgs(args []IArg) error {
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
	ctx.mergeFlags(app.Flags)
	ctx.mergeArgs(app.Args)
	ctx.CurrentCommand = &app.Command

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

	if ctx.CurrentCommand.HasSubCommands() {
		cmd, ok := ctx.CurrentCommand.commands_map[token]
		if !ok {
			return templateManager.makeError("UnexpectedTokenTemplate", TokenTemplateContext{Name: token, Extra: "command"})
		}
		ctx.CurrentCommand = cmd
		ctx.mergeArgs(cmd.Args)
		ctx.mergeFlags(cmd.Flags)
	} else if len(ctx.arguments_lookup) > 0 {
		if !ctx.mixArgsAndFlags {
			// no more flags
			ctx.argsOnly = true
		}
		arg := ctx.nextArg()
		if arg == nil {
			return templateManager.makeError("UnknownArgument", TokenTemplateContext{Name: token})
		}
		err := arg.SetValue(token)
		if err != nil {
			return err
		}
	}

	return nil
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
		return templateManager.makeError("UnknownFlagTemplate", FlagTemplateContext{Name: flag_name, Prefix: "--"})
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
				templateManager.makeError("UnexpectedFlagValueTemplate", FlagTemplateContext{Name: flag_name, Value: flag_value})
			}
		}
	}
	err := flag.SetValue(flag_value)
	if err != nil {
		return templateManager.makeError("FlagValidationFailed", FlagTemplateContext{Name: flag_name, Value: flag_value, Extra: err.Error()})
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
						return templateManager.makeError("UnexpectedFlagValueTemplate", FlagTemplateContext{Short: r, Prefix: "-"})
					}
				}
				err := flag.SetValue(flag_value)
				if err != nil {
					return templateManager.makeError("FlagValidationFailed", FlagTemplateContext{Short: r, Value: flag_value, Extra: err.Error()})
				}
				return nil
			}
		} else {
			return templateManager.makeError("UnknownFlagTemplate", FlagTemplateContext{Name: string(r), Prefix: "-"})
		}
	}

	return nil
}

// Group flags and arguments according to their validation group; ignore short flags
//
// only flags and arguments from one group can be set,  i.e groups are mutially exclusive
func (ctx *ParseContext) validateGrouping(set map[string]IFlagArg) error {

	groups := make(map[string][]IFlagArg)
	// group arguments and flags before validation
	for fa_name, fa := range set {
		if len(fa_name) == 1 {
			// skip short flags
			continue
		}
		flag_groups := fa.GetValidationGroups()
		if len(flag_groups) == 1 {
			g_name := flag_groups[0]
			g, ok := groups[g_name]
			if !ok {
				g = make([]IFlagArg, 0)

			}
			g = append(g, fa)
			groups[g_name] = g
		}
	}

	group_set := ""
	var fa_set IFlagArg = nil
	for g_name, g := range groups {
		for _, fa := range g {
			if fa.IsSetByUser() {
				if group_set != "" && group_set != g_name {
					// have values set for more then one exclusive group
					return templateManager.makeError("FlagsArgsFromMultipleGroups", TokenTemplateContext{Name: fa_set.GetName(), Extra: fa.GetName()})
				}
				fa_set = fa
				group_set = g_name
			}
		}
	}
	return nil
}
func (ctx *ParseContext) validate(app *Application) error {

	var err error

	// create a single map of things to validate
	flags_and_args := make(map[string]IFlagArg)
	for name, f := range ctx.flags_lookup {
		flags_and_args[name] = f
	}
	// add all arguments usign arg name
	for _, a := range ctx.arguments_lookup {
		flags_and_args["arg_"+a.GetName()] = a
	}

	err = ctx.validateGrouping(flags_and_args)
	if err != nil {
		return err
	}

	for _, a := range ctx.arguments_lookup {
		err = a.ValidateWrapper(app)
		if err != nil {
			return err
		}
	}

	// run validators on all flags
	for _, f := range ctx.flags_lookup {
		err = f.ValidateWrapper(app)
		if err != nil {
			return err
		}
	}

	// check for required flags
	for _, f := range ctx.flags_lookup {
		if f.IsRequired() && !f.IsSetByUser() {
			return templateManager.makeError("MissingRequired", FlagTemplateContext{Name: f.GetName(), Short: f.GetShort()})
		}
	}
	// check for required arguments
	for _, a := range ctx.arguments_lookup {
		if a.IsRequired() && !a.IsSetByUser() {
			return templateManager.makeError("MissingRequired", ArgTemplateContext{Name: a.GetName()})
		}
	}

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
		cmd = cmd.parent
	}
	return nil
}
