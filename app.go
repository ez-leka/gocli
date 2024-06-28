package gocli

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/ez-leka/gocli/i18n"
	"golang.org/x/exp/slices"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Terminator func(status int)
type GlobalFlagsHandler func()

func NilTerminator(int) {}

// An Application contains the definitions of flags, arguments and commands
// for an application.
type Application struct {
	Command
	// generic preset flags and commands
	// Help flag. Can be customized  before calling Run. Use GetHelpFlag to access
	helpFlag IFlag
	// Version flag. Can be customized before calling Run. Use GetVersionFlag to access
	versionFlag       IFlag
	ShowHelpCommand   bool
	UseOptionsCommand bool
	MixArgsAndFlags   bool
	Author            string
	Version           string
	ShellCompletion   bool // if set to true generate command is added that will generate bash or zsh completing shell
	Terminator        Terminator
	// this handler is called after command oline is parced but vefore any validation or prcessing.
	// it is useful if you have such global flags as log level, output format , etc that you want to confgure BEFOER caling custom (or any) validators
	GlobalFlagsHandler    GlobalFlagsHandler
	errorWriter           io.Writer // Destination for errors.
	usageWriter           io.Writer // Destination for usage
	context               *context
	stopActionPropagation bool
	bashCompletionFlag    IFlag
	Path                  string
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

func (c *Application) AddArgs(args []IArg) {
	panic("Argument specification is not allowed. Application can have commands and global options(flag) only.")
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
func (a *Application) SetErrorWriter(w io.Writer) {
	a.errorWriter = w
}

// Sets write to be used for uage and erros
func (a *Application) SetWriter(w io.Writer) {
	a.usageWriter = w
}

func (a *Application) GetUsageWriter() io.Writer {
	return a.usageWriter
}

func (a *Application) GetErrorWriter() io.Writer {
	return a.errorWriter
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
	// first check if we doing shell completion and only report parse errors if not
	if a.checkCompletion(args) {
		return nil
	}
	if err != nil {
		a.printUsage(err)
		return err
	}

	if a.GlobalFlagsHandler != nil {
		a.GlobalFlagsHandler()
	}

	// if help flag was set app will exit with succsess
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

func (a *Application) checkCompletion(args []string) bool {
	if a.bashCompletionFlag != nil && a.bashCompletionFlag.GetValue().(bool) {
		completions := a.context.resolveCompletion(a, args)

		fmt.Printf("%s", strings.Join(completions, "\n"))
		return true
	}
	return false
}

func (a *Application) checkHelpRequested() bool {

	if a.helpFlag.GetValue().(bool) {
		a.printUsage(nil)
		return true
	} else {
		return false
	}
}

func (a *Application) checkVersionRequested() bool {
	if a.versionFlag == nil {
		return false
	}

	if a.versionFlag.GetValue().(bool) {
		fmt.Fprintln(a.usageWriter, a.Version)
		return true
	} else {
		return false
	}

}

func (a *Application) printError(err error) {

	if int_err, ok := err.(*i18n.Error); ok {
		templateManager.FormatTemplate(a.errorWriter, int_err.GetKey(), int_err.GetData())
		fmt.Fprintln(a.errorWriter)
	} else {
		fmt.Fprintln(a.errorWriter, templateManager.GetLocalizedString("Error", err))
	}
}
func (a *Application) printUsage(err error) {
	if err != nil && err.Error() != "" {
		a.printError(err)
	}

	if err := a.formatUsage(); err != nil {
		fmt.Fprintln(a.errorWriter, err.Error())
		a.Terminate(1)
	}
	a.Terminate(0)
}

func (a *Application) formatUsage() error {

	show_hidden_flags := true
	if a.UseOptionsCommand && a.context.CurrentCommand.level == 0 {
		// hide hidden application options that are shown with built in options command
		show_hidden_flags = false
	}

	templateCtx := UsageTemplateContext{
		AppName: a.Name,
		// Synopsis:          a.context.CurrentCommand.GetSynopsis(),
		CurrentCommand:    *a.context.CurrentCommand,
		Flags:             lookupFlagsForUsage(a.context.flags_lookup, a.context.CurrentCommand.level, show_hidden_flags),
		Args:              lookupArgsForUsage(a.context.arguments_lookup),
		Level:             a.context.CurrentCommand.level,
		UseOptionsCommand: a.UseOptionsCommand,
	}

	return templateManager.FormatTemplate(a.usageWriter, "AppUsageTemplate", templateCtx, WithOutput(TemplateTerminal))
}

func (a *Application) GetHelpFlag() IFlag {

	if a.helpFlag == nil {
		// add help flag - it is always present
		help_short, _ := utf8.DecodeRuneInString(templateManager.GetLocalizedString("HelpFlagShort"))
		a.helpFlag = &Flag[Bool]{
			Name:  templateManager.GetLocalizedString("HelpCommandAndFlagName"),
			Short: help_short,
			Usage: templateManager.GetLocalizedString("HelpFlagUsageTemplate"),
		}
		a.AddFlag(a.helpFlag)
	}
	return a.helpFlag
}
func (a *Application) GetVersionFlag() IFlag {
	// add version flag is version value is set
	if a.Version != "" && a.versionFlag == nil {
		version_short, _ := utf8.DecodeRuneInString(templateManager.GetLocalizedString("VersionFlagShort"))

		a.versionFlag = &Flag[Bool]{
			Name:  templateManager.GetLocalizedString("VersionFlagName"),
			Short: version_short,
			Usage: templateManager.GetLocalizedString("VersionFlagUsageTemplate"),
		}
		a.AddFlag(a.versionFlag)
	}
	return a.versionFlag
}

func (a *Application) GenerateBashCompletion(writer io.Writer, kind string) error {
	template := cases.Title(templateManager.localizer.GetLanguage()).String(kind) + "CompletionTemplate"
	return templateManager.FormatTemplate(writer, template, a, WithOutput(TemplateText))
}

func (a *Application) init() error {
	if a.initialized {
		return nil
	}

	// make sure we create help flag
	a.GetHelpFlag()

	if a.ShellCompletion {
		a.AddCommand(Command{
			Name:        templateManager.GetLocalizedString("ShellCompletionCommand"),
			Description: templateManager.GetLocalizedString("ShellCompletionCommandDesc"),
			Args: []IArg{
				&Arg[OneOf]{
					Name:     templateManager.GetLocalizedString("ShellCompletionArgName"),
					Usage:    templateManager.GetLocalizedString("ShellCompetionArgUsage"),
					Hints:    []string{"bash", "zsh"},
					Default:  "bash",
					Required: false,
				},
			},
			Action: func(a *Application, c *Command, i interface{}) (interface{}, error) {

				a.Path = os.Args[0]

				shell, _ := a.GetArgumentValue("shell")

				f, err := os.Create("completion." + shell.(string))
				if err != nil {
					fmt.Println(err.Error())
				}
				defer f.Close()
				w := bufio.NewWriter(f)

				err = a.GenerateBashCompletion(w, shell.(string))
				if err != nil {
					fmt.Println(err.Error())
					return nil, err
				}
				w.Flush()
				return nil, nil
			},
		})

		// add bash completion flag
		a.bashCompletionFlag = &Flag[Bool]{
			Name:     "bash-completions", // not localizable - internal
			Usage:    templateManager.GetLocalizedString("ShellCompletionFlagUsageTemplate"),
			Hidden:   true,
			internal: true,
		}
		a.AddFlag(a.bashCompletionFlag)
	}

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
	a.GetVersionFlag()

	// add command to generate documentation
	a.AddCommand(Command{
		Name:        templateManager.GetLocalizedString("DocGenerationCommand"),
		Description: templateManager.GetLocalizedString("DocGenerationCommandDesc"),
		Usage:       "",
		Args: []IArg{
			&Arg[OneOf]{
				Name:     templateManager.GetLocalizedString("DocGenerationFormatArgName"),
				Usage:    templateManager.GetLocalizedString("DocGenerationFormatArgUsage"),
				Hints:    []string{string(TemplateHTML), string(TemplateMarkdown), string(TemplateManpage)},
				Required: false,
				Default:  string(TemplateMarkdown),
			}},
		Flags: []IFlag{
			&Flag[String]{
				Name:     templateManager.GetLocalizedString("DocGenerationCssFlagName"),
				Usage:    templateManager.GetLocalizedString("DocGenerationCssFlagUsage"),
				Default:  "",
				Required: false,
			},
			&Flag[String]{
				Name:     templateManager.GetLocalizedString("DocGenerationIconFlagName"),
				Usage:    templateManager.GetLocalizedString("DocGenerationIconFlagUsage"),
				Default:  "",
				Required: false,
			},
			&Flag[Bool]{
				Name:     templateManager.GetLocalizedString("DocGenerationTocFlagName"),
				Usage:    templateManager.GetLocalizedString("DocGenerationTocFlagUsage"),
				Required: false,
				Default:  "false",
			},
		},
		Action: func(a *Application, c *Command, i interface{}) (interface{}, error) {
			format, _ := a.GetArgumentValue(templateManager.GetLocalizedString("DocGenerationFormatArgName"))
			css, _ := a.GetFlagValue(templateManager.GetLocalizedString("DocGenerationCssFlagName"))
			icon, _ := a.GetFlagValue(templateManager.GetLocalizedString("DocGenerationIconFlagName"))
			toc, _ := a.GetFlagValue(templateManager.GetLocalizedString("DocGenerationTocFlagName"))

			// documentation is generated recurcively starting with app
			buf := bytes.NewBuffer(nil)

			if err := a.generateDocumentation(buf, a.Command, make([]string, 0), 0); err != nil {
				return nil, err
			}
			return nil, templateManager.generateTemplateOutput(a.usageWriter, buf,
				WithTitle(a.Name),
				WithOutput(OutputFormat(format.(string))),
				WithCSS(css.(string)),
				WithIcon(icon.(string)),
				WithTOC(toc.(bool)),
			)
		},
	})

	a.Command.init()

	return nil
}

func (a *Application) generateDocumentation(buf *bytes.Buffer, cmd Command, args []string, level int) error {

	// make context for current command
	args = append(args, cmd.Name)

	a.context.level = 0
	a.context.parse(a, args[1:])

	templateCtx := UsageTemplateContext{
		AppName:        a.Name,
		CurrentCommand: *a.context.CurrentCommand,
		Flags:          lookupFlagsForUsage(a.context.flags_lookup, a.context.CurrentCommand.level, true),
		Args:           lookupArgsForUsage(a.context.arguments_lookup),
		Level:          level,
		DocGeneration:  true,
	}
	templateManager.currentLevel = level

	if err := templateManager.doFormatTemplate(buf, "AppUsageTemplate", templateCtx); err != nil {
		return err
	}

	for _, sub_cmd := range cmd.Commands {
		if err := a.generateDocumentation(buf, *sub_cmd, args, level+1); err != nil {
			return err
		}
	}

	return nil
}
