package gocli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/ez-leka/gocli/i18n"
	"golang.org/x/exp/slices"
	"golang.org/x/text/language"
)

type Terminator func(status int)

func NilTerminator(int) {}

// An Application contains the definitions of flags, arguments and commands
// for an application.
type Application struct {
	Command
	// generic preset flags and commands
	// Help flag. Can be customized  before calling Run
	HelpFlag IFlag
	// Version flag. Can be customized before calling Run
	VersionFlag           IFlag
	ShowHelpCommand       bool
	MixArgsAndFlags       bool
	Author                string
	Version               string
	Terminator            Terminator
	errorWriter           io.Writer // Destination for errors.
	usageWriter           io.Writer // Destination for usage
	context               *context
	stopActionPropagation bool
}

// Creates a new gocli application.
func New() *Application {

	app := &Application{
		Command: Command{
			Name:  filepath.Base(os.Args[0]),
			Usage: "",
		},
		MixArgsAndFlags: true, // default
		usageWriter:     os.Stdout,
		errorWriter:     os.Stderr,
		Terminator:      os.Exit,
		context:         &context{},
	}
	initTemplateManager()
	return app
}

func (a *Application) GetTemplateManager() *TemplateManager {
	return templateManager
}

func (a Application) SetLanguage(tag language.Tag) {
	templateManager.localizer.SetLanguage(tag)
}

func (a *Application) GetArgument(name string) (IArg, error) {

	idx := slices.IndexFunc(a.context.arguments_lookup, func(arg IArg) bool { return arg.GetName() == name })
	if idx < 0 {
		return nil, i18n.NewError("UnknownElementTemplate", ElementTemplateContext{Element: &Arg[String]{Name: name}})
	}
	return a.context.arguments_lookup[idx], nil
}

func (a *Application) GetArgumentValue(name string) (interface{}, error) {
	arg, err := a.GetArgument(name)

	if err != nil {
		return nil, err
	}

	return arg.GetValue(), nil
}

func (a *Application) GetFlagValue(name string) (interface{}, error) {
	f, err := a.GetFlag(name)

	if err != nil {
		return nil, err
	}

	return f.GetValue(), nil
}
func (a *Application) GetFlag(name string) (IFlag, error) {
	f, ok := a.context.flags_lookup[name]
	if !ok {
		return nil, i18n.NewError("UnknownElementTemplate", ElementTemplateContext{Element: &Flag[String]{Name: name}})
	}

	return f, nil
}

// teminate application with exit status
func (a *Application) Terminate(status int) {
	a.Terminator(status)
}

// ErrorWriter sets the io.Writer to use for errors.
func (a *Application) SetErrorWriter(w io.Writer) *Application {
	a.errorWriter = w
	return a
}

// Sets write to be used for uage and erros
func (a *Application) SetWriter(w io.Writer) *Application {
	a.usageWriter = w
	return a
}
func (a *Application) Stop() {
	a.stopActionPropagation = true
}

// Run :
//   - parses command-line arguments,
//   - populates all flags and argumants,
//   - validates arguments, flags and commands in that order
//   - executes appropriate command
func (a *Application) Run(args []string) (err error) {

	if err := a.init(); err != nil {
		return err
	}

	a.context.mixArgsAndFlags = a.MixArgsAndFlags

	err = a.context.parse(a, args[1:])
	if err != nil {
		a.printUsage(err)
		return err
	}

	// if hel flag was set app will exit with succsess
	if a.checkHelpRequested() {
		return nil
	}

	if a.checkVersionRequested() {
		return nil
	}
	// run custom validators and validate required flags and args
	err = a.context.validate(a)
	if err != nil {
		a.printUsage(err)
		return err
	}

	// execute command actions
	err = a.context.execute(a)
	if err != nil {
		a.printError(err)
		return err
	}

	return err
}

func (a *Application) checkHelpRequested() bool {

	if a.HelpFlag.GetValue().(bool) {
		a.printUsage(nil)
		return true
	} else {
		return false
	}
}

func (a *Application) checkVersionRequested() bool {
	if a.VersionFlag == nil {
		return false
	}

	if a.VersionFlag.GetValue().(bool) {
		fmt.Fprintln(a.usageWriter, a.Version)
		return true
	} else {
		return false
	}

}

func (a *Application) printError(err error) {

	if int_err, ok := err.(*i18n.Error); ok {
		templateManager.formatTemplate(a.errorWriter, int_err.GetKey(), int_err.GetData())
	} else {
		fmt.Fprintln(a.errorWriter, templateManager.GetLocalizedString("Error", err))
	}
}
func (a *Application) printUsage(err error) {
	if err != nil {
		a.printError(err)
	}

	if err := a.FormatUsage(); err != nil {
		fmt.Fprintln(a.errorWriter, err.Error())
		a.Terminate(1)
	}
	a.Terminate(0)
}

func (a *Application) FormatUsage() error {

	templateCtx := UsageTemplateContext{
		AppName:        a.Name,
		CurrentCommand: *a.context.CurrentCommand,
		Flags:          mapIFlag(a.context.flags_lookup),
		Args:           mapIArg(a.context.arguments_lookup),
	}
	return templateManager.formatTemplate(a.usageWriter, "AppUsageTemplate", templateCtx)
}

func (a *Application) init() error {
	if a.initialized {
		return nil
	}

	// add help flag - it is always present
	help_short, _ := utf8.DecodeRuneInString(templateManager.GetLocalizedString("HelpFlagShort"))
	a.HelpFlag = &Flag[Bool]{
		Name:  templateManager.GetLocalizedString("HelpCommandAndFlagName"),
		Short: help_short,
		Usage: templateManager.GetLocalizedString("HelpFlagUsageTemplate"),
	}
	a.AddFlag(a.HelpFlag)

	// If we have subcommands, add a help command at the top-level.
	if a.ShowHelpCommand {
		command_arg_name := templateManager.GetLocalizedString("CommandArgName")
		help_cmd := &Command{
			Name:  templateManager.GetLocalizedString("HelpCommandAndFlagName"),
			Usage: templateManager.GetLocalizedString("HelpCommandUsage"),
			Args: []IArg{
				&Arg[[]String]{
					Name:  command_arg_name,
					Usage: templateManager.GetLocalizedString("HelpCommandArgUsage"),
				},
			},
			Action: func(app *Application, c *Command, in_data interface{}) (interface{}, error) {
				cmd_arg, err := a.GetArgument(command_arg_name)
				command := cmd_arg.GetValue().([]string)
				if err != nil {
					a.printUsage(nil)
				}
				a.context.parse(a, command)

				a.printUsage(nil)
				a.Terminate(0)
				return nil, nil
			},
		}
		// make help first command
		a.Commands = append([]*Command{help_cmd}, a.Commands...)
	}
	// add version flag is version value is set
	if a.Version != "" {
		version_short, _ := utf8.DecodeRuneInString(templateManager.GetLocalizedString("VersionFlagShort"))

		a.VersionFlag = &Flag[Bool]{
			Name:  templateManager.GetLocalizedString("VersionFlagName"),
			Short: version_short,
			Usage: templateManager.GetLocalizedString("VersionFlagUsageTemplate"),
		}
		a.AddFlag(a.VersionFlag)
	}

	a.Command.init()

	return nil
}
