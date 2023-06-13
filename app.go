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

// An Application contains the definitions of flags, arguments and commands
// for an application.
type Application struct {
	Command
	// generic preset flags and commands
	// Help flag. Can be customized  before calling Run
	HelpFlag IFlag
	// Version flag. Can be customized before calling Run
	VersionFlag     IFlag
	ShowHelpCommand bool
	MixArgsAndFlags bool
	Author          string
	Version         string
	errorWriter     io.Writer // Destination for errors.
	usageWriter     io.Writer // Destination for usage
	terminate       func(status int)
	context         *ParseContext
	language        language.Tag
}

// Creates a new gocli application.
func New(lang string) *Application {

	var err error
	app := &Application{
		Command: Command{
			Name:  filepath.Base(os.Args[0]),
			Usage: "",
		},
		MixArgsAndFlags: true, // default
		usageWriter:     os.Stdout,
		errorWriter:     os.Stderr,
		terminate:       os.Exit,
		context:         &ParseContext{},
	}

	app.language = language.Make(lang)
	templateManager, err = NewTemplateManager(app.language)
	if err != nil {
		panic(err)
	}

	return app
}

func (a *Application) AddTemplateFunction(name string, f any) {
	templateManager.AddFunction(name, f)
}

func (a *Application) AddTranslation(lang language.Tag, entries i18n.Entries) {
	templateManager.AddTranslation(lang, entries)
}

func (a *Application) GetArgument(name string) (IArg, error) {

	idx := slices.IndexFunc(a.context.arguments_lookup, func(arg IArg) bool { return arg.GetName() == name })
	if idx < 0 {
		return nil, templateManager.makeError("UnknownElementTemplate", ElementTemplateContext{Element: &Arg[String]{Name: name}})
	}
	return a.context.arguments_lookup[idx], nil
}

func (a *Application) GetStringArg(name string) (string, error) {

	// find argument by name
	arg, err := a.GetArgument(name)
	if err != nil {
		return "", err
	}
	if arg.IsCumulative() {
		return "", templateManager.makeError("WrongElementTypeTemplate", ElementTemplateContext{Element: arg})
	}
	return arg.GetValue().(string), nil
}

func (a *Application) GetListArg(name string) ([]string, error) {
	// find argument by name
	arg, err := a.GetArgument(name)
	if err != nil {
		return []string{}, err
	}

	if arg.IsCumulative() {
		return arg.GetValue().([]string), nil
	}
	return nil, templateManager.makeError("WrongElementTypeTemplate", ElementTemplateContext{Element: arg})
}

func (a *Application) GetFlag(name string) (IFlag, error) {
	f, ok := a.context.flags_lookup[name]
	if !ok {
		return nil, templateManager.makeError("UnknownElementTemplate", ElementTemplateContext{Element: &Flag[String]{Name: name}})
	}

	return f, nil
}

func (a *Application) GetBoolFlag(name string) (bool, error) {

	f, err := a.GetFlag(name)
	if err != nil {
		return false, err
	}
	if !f.IsBool() {
		return false, templateManager.makeError("WrongElementTypeTemplate", ElementTemplateContext{Element: f})
	}
	return f.GetValue().(bool), nil
}

func (a *Application) GetStringFlag(name string) (string, error) {
	f, err := a.GetFlag(name)
	if err != nil {
		return "", err
	}
	if f.IsBool() || f.IsCumulative() {
		return "", templateManager.makeError("WrongElementTypeTemplate", ElementTemplateContext{Element: f})
	}

	return f.GetValue().(string), nil
}

func (a *Application) GetListFlag(name string) ([]string, error) {
	f, err := a.GetFlag(name)
	if err != nil {
		return []string{}, err
	}
	if f.IsCumulative() {
		return f.GetValue().([]string), nil
	}
	return nil, templateManager.makeError("WrongElementTypeTemplate", ElementTemplateContext{Element: f})

}

// Terminate specifies the termination handler. Defaults to os.Exit(status).
// If nil is passed, a no-op function will be used.
func (a *Application) Terminate(terminate func(int)) {
	if terminate == nil {
		terminate = func(int) {}
	}
	a.terminate = terminate
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

// Run :
//   - parses command-line arguments,
//   - populates all flags and argumants,
//   - validates arguments, flags and commands in that order
//   - executes appropriate command
func (a *Application) Run(args []string) (err error) {

	if err := a.Init(); err != nil {
		return err
	}

	a.context.mixArgsAndFlags = a.MixArgsAndFlags

	err = a.context.parse(a, args[1:])
	if err != nil {
		a.printUsage(err)
		return err
	}

	// if hel flag was set app will exit with succsess
	a.checkHelpRequested()

	// run custom validators and validate required flags and args
	err = a.context.validate(a)
	if err != nil {
		a.printUsage(err)
		return err
	}

	// execute command actions
	err = a.context.execute(a)
	if err != nil {
		a.printUsage(err)
		return err
	}

	return err
}

func (a *Application) checkHelpRequested() {

	need_help, err := a.GetBoolFlag(a.HelpFlag.GetName())
	if err != nil {
		a.printUsage(err)
	}

	if need_help {
		a.printUsage(nil)
	}
}

func (a *Application) printUsage(err error) {
	if err != nil {
		if int_err, ok := err.(*i18n.Error); ok {
			fmt.Fprintln(a.errorWriter, templateManager.formatTemplate(a.errorWriter, int_err.GetKey(), int_err.GetData()))
		} else {
			fmt.Fprintln(a.errorWriter, templateManager.GetMessage("Error", err))
		}
	}

	if err := a.FormatUsage(); err != nil {
		fmt.Println(err.Error())
		a.terminate(1)
	}
	a.terminate(0)
}

func (a *Application) FormatUsage() error {

	templateCtx := UsageTemplateContext{
		AppName:        a.Name,
		CurrentCommand: *a.context.CurrentCommand,
		Flags:          MapIFlag(a.context.flags_lookup),
		Args:           MapIArg(a.context.arguments_lookup),
	}
	return templateManager.formatTemplate(a.usageWriter, "AppUsageTemplate", templateCtx)
}

func (a *Application) Init() error {
	if a.initialized {
		return nil
	}
	if len(a.Commands) > 0 && len(a.Args) > 0 {
		return templateManager.makeError("MixArgsCommandsTemplate", a.Command)
	}

	// add help flag - it is always present
	help_short, _ := utf8.DecodeRuneInString(templateManager.GetMessage("HelpFlagShort"))
	a.HelpFlag = &Flag[Bool]{
		Name:  templateManager.GetMessage("HelpCommandAndFlagName"),
		Short: help_short,
		Usage: templateManager.GetMessage("HelpFlagUsageTemplate"),
	}
	a.AddFlag(a.HelpFlag)

	// If we have subcommands, add a help command at the top-level.
	if a.ShowHelpCommand {
		command_arg_name := templateManager.GetMessage("CommandArgName")
		help_cmd := &Command{
			Name:  templateManager.GetMessage("HelpCommandAndFlagName"),
			Usage: templateManager.GetMessage("HelpCommandUsage"),
			Args: []IArg{
				&Arg[[]String]{
					Name:  command_arg_name,
					Usage: templateManager.GetMessage("HelpCommandArgUsage"),
				},
			},
			Action: func(app *Application, c *Command, in_data interface{}) (interface{}, error) {
				command, err := a.GetListArg(command_arg_name)
				if err != nil {
					a.printUsage(nil)
				}
				a.context.parse(a, command)

				a.printUsage(nil)
				a.terminate(0)
				return nil, nil
			},
		}
		// make help first command
		a.Commands = append([]*Command{help_cmd}, a.Commands...)
	}
	// add version flag is version value is set
	if a.Version != "" {
		version_short, _ := utf8.DecodeRuneInString(templateManager.GetMessage("VersionFlagShort"))

		a.VersionFlag = &Flag[Bool]{
			Name:  templateManager.GetMessage("VersionFlagName"),
			Short: version_short,
			Usage: templateManager.GetMessage("VersionFlagUsageTemplate"),
		}
		a.AddFlag(a.VersionFlag)
	}

	a.init()

	return nil
}
