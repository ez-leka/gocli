package gocli

import (
	"errors"
	"os"

	"reflect"
	"testing"

	"github.com/ez-leka/gocli/i18n"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"golang.org/x/text/language"
)

func TestApplication_Run(t *testing.T) {

	// var output bytes.Buffer
	// var rescueStdout *os.File

	expected_files := []string{"cmd", "test2.txt", "test3.txt", "test4.txt", "test5.txt", "test6.txt", "test6.txt", "test7.txt", "a"}
	for _, f := range expected_files {
		file, _ := os.OpenFile(f, os.O_RDONLY|os.O_CREATE, 0666)
		file.Close()
	}
	arg_expected_files := []string{"test2.txt", "test3.txt", "test4.txt"}

	var action_result string

	tests := []struct {
		name    string
		args    []string
		setup   func() *Application
		wantErr bool
		check   func(a *Application) error
	}{
		{
			name: "no cmd",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
				})
				app.AddCommand(Command{
					Name:        "command2",
					Description: `{{.FullCommand}} second command`,
				})
				app.Terminator = NilTerminator
				return app
			},
			args:    []string{"test"},
			wantErr: true,
		},
		{
			name: "short help flag",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
				})
				app.AddCommand(Command{
					Name:        "command2",
					Description: `{{.FullCommand}} second command`,
				})
				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "-h"},
			wantErr: false,
		},
		{
			name: "long help flag",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
				})
				app.AddCommand(Command{
					Name:        "command2",
					Description: `{{.FullCommand}} second command`,
				})
				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "--help"},
			wantErr: false,
		},
		{
			name: "long version flag",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
				})
				app.AddCommand(Command{
					Name:        "command2",
					Description: `{{.FullCommand}} second command`,
				})
				app.Terminator = NilTerminator

				app.Version = "v1.2.3"
				return app
			},
			args:    []string{"test", "--version"},
			wantErr: false,
		},
		{
			name: "short version flag",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
				})
				app.AddCommand(Command{
					Name:        "command2",
					Description: `{{.FullCommand}} second command`,
				})
				app.Terminator = NilTerminator

				app.Version = "v1.2.3"
				return app
			},
			args:    []string{"test", "-v"},
			wantErr: false,
		},
		{
			name: "localizaion/match localized flag",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
				})
				app.AddCommand(Command{
					Name:        "command2",
					Description: `{{.FullCommand}} second command`,
				})
				app.GetTemplateManager().UpdateTranslation(language.MustParse("en_us"), "VersionFlagUsageTemplate", "show git SHA")

				ru_tag := language.MustParse("ru")
				app.GetTemplateManager().AddTranslation(ru_tag, i18n.Entries{
					"HelpCommandAndFlagName": "Помощь",
					"Flags":                  "Флаги",
					"Arguments":              "Аргументы",
					"Usage":                  "Использование",
					"HelpFlagUsageTemplate":  "распечатать информацию о флагах",
				})
				app.Terminator = NilTerminator
				app.SetLanguage(ru_tag)
				return app
			},
			args:    []string{"test", "--Помощь"},
			wantErr: false,
		},
		{
			name: "flag sorting",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddFlag(&Flag[Bool]{
					Name: "appf1",
				})
				app.AddFlag(&Flag[Bool]{
					Name: "appf2",
				})
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
					Flags: []IFlag{
						&Flag[Bool]{
							Name: "cmdf1",
						},
						&Flag[Bool]{
							Name: "cmdf2",
						},
					},
				})
				app.AddCommand(Command{
					Name:        "command2",
					Description: `{{.FullCommand}} second command`,
				})
				app.Terminator = NilTerminator
				return app
			},
			args:    []string{"test", "command1", "-h"},
			wantErr: false,
		},

		{
			// testing parsing of and validation of :
			// email,
			// shot flags compressed together,
			// short flags compressed  together with last one having a value,
			// short and log flags with =, without = (short flags can have value attached -ftext.txt), separated value
			// cululative flags (filename in this case )
			name: "flag parsing",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
					Flags: []IFlag{
						&Flag[[]File]{
							Name:    "filename",
							Short:   'f',
							Usage:   "yaml or json file (or files, if wildcard) identifying resources to delete",
							Default: "",
						},
						&Flag[Email]{
							Name:    "email",
							Short:   'e',
							Usage:   "email",
							Default: "",
						},
						&Flag[Bool]{
							Name:    "all",
							Short:   'a',
							Usage:   "yaml or json file (or files, if wildcard) identifying resources to delete",
							Default: "",
						},
						&Flag[TimeStamp]{
							Name:    "to",
							Short:   't',
							Usage:   "timestamp",
							Default: "",
						},
						&Flag[OneOf]{
							Name:     "output",
							Short:    'o',
							Usage:    "Output format",
							Hints:    []string{"json", "table", "yaml"},
							Default:  "table",
							Required: false,
						},
					},
				})
				app.Version = "v1.2.3" //need -v flag for testing
				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "command1", "--output=json", "-e", "my@gnode.org", "-fcmd", "-f=test2.txt", "-f", "test3.txt", "--filename=test4.txt", "--filename", "test5.txt", "-vaftest6.txt", "-vaf=test6.txt", "-vaf", "test7.txt", "-vfa"},
			wantErr: false,
			check: func(a *Application) error {
				// make flags are collected as expected
				f, err := a.GetFlag("email")
				if err != nil {
					return err
				}
				if f.GetValue().(string) != "my@gnode.org" {
					return errors.New("email did not match")
				}
				f, err = a.GetFlag("output")
				if err != nil {
					return err
				}
				if f.GetValue().(string) != "json" {
					return errors.New("output did not match")
				}
				f, err = a.GetFlag("filename")
				if err != nil {
					return err
				}
				files := f.GetValue().([]string)
				slices.Sort(files)
				slices.Sort(expected_files)
				if !reflect.DeepEqual(expected_files, files) {
					return errors.New("files did not match")
				}
				return nil
			},
		},
		{
			// testign parsing of arguments
			// email
			// cumulative argument with comma separated valus nd consuming teh rest of arguments
			name: "argument parsing",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
					Args: []IArg{
						&Arg[Email]{
							Name:    "email",
							Usage:   "email",
							Default: "",
						},
						&Arg[[]File]{
							Name:    "filename",
							Usage:   "yaml or json file (or files, if wildcard) identifying resources to delete",
							Default: "",
						},
					},
				})
				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "command1", "my@gnode.org", "test2.txt,test3.txt", "test4.txt"},
			wantErr: false,
			check: func(a *Application) error {
				arg, err := a.GetArgument("email")
				if err != nil {
					return err
				}
				if arg.GetValue().(string) != "my@gnode.org" {
					return errors.New("email did not match")
				}
				arg, err = a.GetArgument("filename")
				if err != nil {
					return err
				}
				files := arg.GetValue().([]string)
				slices.Sort(files)
				slices.Sort(arg_expected_files)
				if !reflect.DeepEqual(arg_expected_files, files) {
					return errors.New("files did not match")
				}

				return nil
			},
		},
		{
			// testign mixing flags and arguments
			name: "mixing flags and arguments",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
					Flags: []IFlag{
						&Flag[OneOf]{
							Name:     "output",
							Short:    'o',
							Usage:    "Output format",
							Hints:    []string{"json", "table", "yaml"},
							Default:  "table",
							Required: false,
						},
					},
					Args: []IArg{
						&Arg[Email]{
							Name:    "email",
							Usage:   "email",
							Default: "",
						},
						&Arg[[]File]{
							Name:    "filename",
							Usage:   "yaml or json file (or files, if wildcard) identifying resources to delete",
							Default: "",
						},
					},
				})
				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "command1", "my@gnode.org", "-o", "json", "test2.txt,test3.txt", "test4.txt"},
			wantErr: false,
			check: func(a *Application) error {
				arg, err := a.GetArgument("email")
				if err != nil {
					return err
				}
				if arg.GetValue().(string) != "my@gnode.org" {
					return errors.New("email did not match")
				}
				arg, err = a.GetArgument("filename")
				if err != nil {
					return err
				}
				files := arg.GetValue().([]string)
				slices.Sort(files)
				slices.Sort(arg_expected_files)
				if !reflect.DeepEqual(arg_expected_files, files) {
					return errors.New("files did not match")
				}

				f, err := a.GetFlag("output")
				if err != nil {
					return err
				}
				if f.GetValue().(string) != "json" {
					return errors.New("output did not match")
				}

				return nil
			},
		},
		{
			// testign mixing flags and arguments when feature is disabled, i.e once first argument is parsed no more flags are allowed
			// Shoudl fail
			name: "mixing flags and arguments - not allowed ",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
					Flags: []IFlag{
						&Flag[OneOf]{
							Name:     "output",
							Short:    'o',
							Usage:    "Output format",
							Hints:    []string{"json", "table", "yaml"},
							Default:  "table",
							Required: false,
						},
					},
					Args: []IArg{
						&Arg[Email]{
							Name:    "email",
							Usage:   "email",
							Default: "",
						},
						&Arg[[]File]{
							Name:    "filename",
							Usage:   "yaml or json file (or files, if wildcard) identifying resources to delete",
							Default: "",
						},
					},
				})
				app.MixArgsAndFlags = false
				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "command1", "my@gnode.org", "-o", "json", "test2.txt,test3.txt", "test4.txt"},
			wantErr: true,
		},
		{
			// grouping of of flags and args
			// user can spefify either -f flag (as many times as needed) or resourse-type/resourse-name arguments but not both
			// USage for cammand looks like
			// test command1 -o[=]=<optput format> (-f filename | resource-type resource-name)
			// Should succeed as we specifued -f flag
			name: "grouping flags and arguments - flag group only",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
					Flags: []IFlag{
						&Flag[OneOf]{
							Name:        "output",
							Short:       'o',
							Placeholder: "output format",
							Usage:       "Output format",
							Hints:       []string{"json", "table", "yaml"},
							Default:     "table",
							Required:    false,
						},
						&Flag[[]File]{
							Name:             "filename",
							Short:            'f',
							Usage:            "yaml or json file (or files, if wildcard) identifying resources to delete",
							ValidationGroups: []string{"file"},
							Default:          "",
						},
					},
					Args: []IArg{
						&Arg[String]{
							Name:             "resource-type",
							Usage:            "type",
							Default:          "",
							ValidationGroups: []string{"name"},
						},
						&Arg[String]{
							Name:             "resource-name",
							Usage:            "name",
							Default:          "",
							ValidationGroups: []string{"name"},
						},
					},
				})

				app.MixArgsAndFlags = false
				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "command1", "-o", "json", "-ftest2.txt", "-f", "test3.txt"},
			wantErr: false,
		},
		{
			// Same as above
			//Should succeed as we specified type and name
			name: "grouping flags and arguments - name group only",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
					Flags: []IFlag{
						&Flag[OneOf]{
							Name:        "output",
							Short:       'o',
							Placeholder: "output format",
							Usage:       "Output format",
							Hints:       []string{"json", "table", "yaml"},
							Default:     "table",
							Required:    false,
						},
						&Flag[[]File]{
							Name:             "filename",
							Short:            'f',
							Usage:            "yaml or json file (or files, if wildcard) identifying resources to delete",
							Default:          "",
							ValidationGroups: []string{"file"},
						},
					},
					Args: []IArg{
						&Arg[String]{
							Name:             "resource-type",
							Usage:            "type",
							Default:          "",
							ValidationGroups: []string{"name"},
						},
						&Arg[String]{
							Name:             "resource-name",
							Usage:            "name",
							Default:          "",
							ValidationGroups: []string{"name"},
						},
					},
				})

				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "command1", "-o", "json", "my_type", "my_name"},
			wantErr: false,
		},
		{
			// Same as above
			//Should fail - has both flag and type
			name: "grouping flags and arguments - both groups -",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
					Flags: []IFlag{
						&Flag[OneOf]{
							Name:        "output",
							Short:       'o',
							Placeholder: "output format",
							Usage:       "Output format",
							Hints:       []string{"json", "table", "yaml"},
							Default:     "table",
							Required:    false,
						},
						&Flag[[]File]{
							Name:             "filename",
							Short:            'f',
							ValidationGroups: []string{"file"},
							Usage:            "yaml or json file (or files, if wildcard) identifying resources to delete",
							Default:          "",
						},
					},
					Args: []IArg{
						&Arg[String]{
							Name:             "resource-type",
							Usage:            "type",
							Default:          "",
							ValidationGroups: []string{"name"},
						},
						&Arg[String]{
							Name:             "resource-name",
							Usage:            "name",
							Default:          "",
							ValidationGroups: []string{"name"},
						},
					},
				})

				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "command1", "-o", "json", "-f", "test2.txt", "my_type", "my_name"},
			wantErr: true,
		},
		{
			// Optional sub-commands - works only with resource_name but not resource-type nor filename flag
			// usage for the command1 looks like :
			// test command1 (-f[=]<filename>| type name | sub-com name)
			// Should succees -f <filename>
			name: "optional subcommand - command -f filename  ",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
					Commands: []*Command{
						{
							Name:             "sub-com",
							Description:      "Optional sub-command",
							ValidationGroups: []string{"subcommand"},
							Optional:         true,
						},
					},
					Flags: []IFlag{
						&Flag[OneOf]{
							Name:        "output",
							Short:       'o',
							Placeholder: "output format",
							Usage:       "Output format",
							Hints:       []string{"json", "table", "yaml"},
							Default:     "table",
							Required:    false,
						},
						&Flag[[]File]{
							Name:             "filename",
							Short:            'f',
							ValidationGroups: []string{"file"},
							Usage:            "yaml or json file (or files, if wildcard) identifying resources to delete",
							Default:          "",
						},
					},
					Args: []IArg{
						&Arg[String]{
							Name:             "resource-type",
							Usage:            "type",
							Default:          "",
							ValidationGroups: []string{"name"},
						},
						&Arg[String]{
							Name:             "resource-name",
							Usage:            "name",
							Default:          "",
							ValidationGroups: []string{"name", "subcommand"},
						},
					},
				})

				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "command1", "-o", "json", "-f", "test2.txt"},
			wantErr: false,
		},
		{
			// same as above
			// test command1 (-f[=]<filename>| type name | sub-com name)
			// Should succees command1 sub-com name
			name: "optional subcommand -  command1 type name",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
					Commands: []*Command{
						{
							Name:             "sub-com",
							Description:      "Optional sub-command",
							ValidationGroups: []string{"subcommand"},
							Optional:         true,
						},
					},
					Flags: []IFlag{
						&Flag[OneOf]{
							Name:        "output",
							Short:       'o',
							Placeholder: "output format",
							Usage:       "Output format",
							Hints:       []string{"json", "table", "yaml"},
							Default:     "table",
							Required:    false,
						},
						&Flag[[]File]{
							Name:             "filename",
							Short:            'f',
							ValidationGroups: []string{"file"},
							Usage:            "yaml or json file (or files, if wildcard) identifying resources to delete",
							Default:          "",
						},
					},
					Args: []IArg{
						&Arg[String]{
							Name:             "resource-type",
							Usage:            "type",
							Default:          "",
							ValidationGroups: []string{"name"},
						},
						&Arg[String]{
							Name:             "resource-name",
							Usage:            "name",
							Default:          "",
							ValidationGroups: []string{"name", "subcommand"},
						},
					},
				})

				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "command1", "-o", "json", "my_type", "my_name"},
			wantErr: false,
		},
		{
			// same as above
			// test command1 (-f[=]<filename>| type name | sub-com name)
			// Should succees sub-com my_name
			name: "optional subcommand - sub_cmd my_name ",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
					Commands: []*Command{
						{
							Name:             "sub-com",
							Description:      "Optional sub-command",
							ValidationGroups: []string{"subcommand"},
							Optional:         true,
						},
					},
					Flags: []IFlag{
						&Flag[OneOf]{
							Name:        "output",
							Short:       'o',
							Placeholder: "output format",
							Usage:       "Output format",
							Hints:       []string{"json", "table", "yaml"},
							Default:     "table",
							Required:    false,
						},
						&Flag[[]File]{
							Name:             "filename",
							Short:            'f',
							ValidationGroups: []string{"file"},
							Usage:            "yaml or json file (or files, if wildcard) identifying resources to delete",
							Default:          "",
						},
					},
					Args: []IArg{
						&Arg[String]{
							Name:             "resource-type",
							Usage:            "type",
							Default:          "",
							ValidationGroups: []string{"name"},
						},
						&Arg[String]{
							Name:             "resource-name",
							Usage:            "name",
							Default:          "",
							ValidationGroups: []string{"name", "subcommand"},
						},
					},
				})

				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "command1", "-o", "json", "sub-com", "my_name"},
			wantErr: false,
		},
		{
			// same as above
			// test command1 (-f[=]<filename>| type name | sub-com name)
			// Should fail sub_com type my_name because sub-com only takes one argument but 2 are present
			// Note: if resource-name argument is cumulative, i.e []String it will consume all remainign arguments and NOT fail
			name: "optional subcommand - sub-com my_type my_name ",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
					Commands: []*Command{
						{
							Name:             "sub-com",
							Description:      "Optional sub-command",
							ValidationGroups: []string{"subcommand"},
							Optional:         true,
						},
					},
					Flags: []IFlag{
						&Flag[OneOf]{
							Name:        "output",
							Short:       'o',
							Placeholder: "output format",
							Usage:       "Output format",
							Hints:       []string{"json", "table", "yaml"},
							Default:     "table",
							Required:    false,
						},
						&Flag[[]File]{
							Name:             "filename",
							Short:            'f',
							ValidationGroups: []string{"file"},
							Usage:            "yaml or json file (or files, if wildcard) identifying resources to delete",
							Default:          "",
						},
					},
					Args: []IArg{
						&Arg[String]{
							Name:             "resource-type",
							Usage:            "type",
							Default:          "",
							ValidationGroups: []string{"name"},
						},
						&Arg[String]{
							Name:             "resource-name",
							Usage:            "name",
							Default:          "",
							ValidationGroups: []string{"name", "subcommand"},
						},
					},
				})

				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "command1", "-o", "json", "sub-com", "my_type", "my_name"},
			wantErr: true,
		},
		{
			// nested commands with global and local flags
			// Should fail and print proper usage for command 1 - command required
			name: "nested commands with global and local flags - non-optional command ",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `Description for command1 goes here `,
					Commands: []*Command{
						{
							Name:        "sub-com",
							Description: "sub-command",
							// Optional:         true,
							// ValidationGroups: []string{"sub-com"},
							Flags: []IFlag{
								&Flag[TimeStamp]{
									Name:  "from",
									Short: 'f',
								},
								&Flag[TimeStamp]{
									Name:  "to",
									Short: 't',
								},
							},
						},
					},
					Flags: []IFlag{
						&Flag[OneOf]{
							Name:        "output",
							Short:       'o',
							Placeholder: "output format",
							Usage:       "Output format",
							Hints:       []string{"json", "table", "yaml"},
							Default:     "table",
							Required:    false,
						},
						&Flag[[]File]{
							Name:             "filename",
							Short:            'f',
							ValidationGroups: []string{"file"},
							Usage:            "yaml or json file (or files, if wildcard) identifying resources to delete",
							Default:          "",
						},
					},
					Args: []IArg{
						&Arg[String]{
							Name:             "resource-type",
							Usage:            "type",
							Default:          "",
							ValidationGroups: []string{"name", "sub-com"},
							Required:         true,
						},
						&Arg[String]{
							Name:             "resource-name",
							Usage:            "name",
							Default:          "",
							ValidationGroups: []string{"name", "sub-com"},
						},
					},
				})

				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "command1", "-o", "json"},
			wantErr: true,
		},
		{
			// nested commands with global and local flags
			// Should fail and print proper usage for command 1 - missing required arg
			name: "nested commands with global and local flags - optional subcommand",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `Description for command1 goes here `,
					Commands: []*Command{
						{
							Name:             "sub-com",
							Description:      "sub-command",
							Optional:         true,
							ValidationGroups: []string{"sub-com"},
							Flags: []IFlag{
								&Flag[TimeStamp]{
									Name:  "from",
									Short: 'f',
								},
								&Flag[TimeStamp]{
									Name:  "to",
									Short: 't',
								},
							},
						},
					},
					Flags: []IFlag{
						&Flag[OneOf]{
							Name:        "output",
							Short:       'o',
							Placeholder: "output format",
							Usage:       "Output format",
							Hints:       []string{"json", "table", "yaml"},
							Default:     "table",
							Required:    false,
						},
						&Flag[[]File]{
							Name:             "filename",
							Short:            'f',
							ValidationGroups: []string{"file"},
							Usage:            "yaml or json file (or files, if wildcard) identifying resources to delete",
							Default:          "",
						},
					},
					Args: []IArg{
						&Arg[String]{
							Name:             "resource-type",
							Usage:            "type",
							Default:          "",
							ValidationGroups: []string{"name", "sub-com"},
							Required:         true,
						},
						&Arg[String]{
							Name:             "resource-name",
							Usage:            "name",
							Default:          "",
							ValidationGroups: []string{"name", "sub-com"},
						},
					},
				})

				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "command1", "-o", "json"},
			wantErr: true,
		},
		{
			// nested commands with global and local flags
			// Should fail and print proper usage for command1 sub-com  - missing required arg
			name: "nested commands with global and local flags - optional subcommand is called ",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `Description for command1 goes here `,
					Commands: []*Command{
						{
							Name:             "sub-com",
							Description:      "Description for sub-command",
							Optional:         true,
							ValidationGroups: []string{"sub-com"},
							Flags: []IFlag{
								&Flag[TimeStamp]{
									Name:  "from",
									Short: 'r',
								},
								&Flag[TimeStamp]{
									Name:  "to",
									Short: 't',
								},
							},
						},
					},
					Flags: []IFlag{
						&Flag[OneOf]{
							Name:        "output",
							Short:       'o',
							Placeholder: "output format",
							Usage:       "Output format",
							Hints:       []string{"json", "table", "yaml"},
							Default:     "table",
							Required:    false,
						},
						&Flag[[]File]{
							Name:             "filename",
							Short:            'f',
							ValidationGroups: []string{"file"},
							Usage:            "yaml or json file (or files, if wildcard) identifying resources to delete",
							Default:          "",
						},
					},
					Args: []IArg{
						&Arg[String]{
							Name:             "resource-type",
							Usage:            "type",
							Default:          "",
							ValidationGroups: []string{"name", "sub-com"},
							Required:         true,
						},
						&Arg[String]{
							Name:             "resource-name",
							Usage:            "name",
							Default:          "",
							ValidationGroups: []string{"name", "sub-com"},
						},
					},
				})

				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "command1", "-o", "json", "sub-com"},
			wantErr: true,
		},
		{
			// command categories top level
			// Should fail and print proper usage
			name: "command categories top level",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name: "command1",
					Category: &CommandCategory{
						Name:  "Beginner",
						Order: 1,
					},
					Description: `Description for command1 goes here `,
				})
				app.AddCommand(Command{
					Name: "command2",
					Category: &CommandCategory{
						Name:  "Beginner",
						Order: 1,
					},
					Description: `Description for command2 goes here `,
				})
				app.AddCommand(Command{
					Name: "command3",
					Category: &CommandCategory{
						Name:  "Intermediate",
						Order: 2,
					},
					Description: `Description for command3 goes here `,
				})

				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test"},
			wantErr: true,
		},
		{
			// command categories top level
			// Should fail and print proper usage
			name: "command categories sub-commands",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name: "command1",
					Category: &CommandCategory{
						Name:  "Beginner",
						Order: 1,
					},
					Description: `Description for command1 goes here `,
				})
				app.AddCommand(Command{
					Name: "command2",
					Category: &CommandCategory{
						Name:  "Beginner",
						Order: 1,
					},
					Description: `Description for command2 goes here `,
					Commands: []*Command{
						{
							Name:        "sub-cmd1",
							Description: "sub-cmd1 description",
							Category: &CommandCategory{
								Name:  "Advanced",
								Order: 1,
							},
						},
						{
							Name:        "sub-cmd2",
							Description: "sub-cmd2 description",
							Category: &CommandCategory{
								Name:  "Advanced",
								Order: 1,
							},
						},
						{
							Name:        "sub-cmd3",
							Description: "sub-cmd3 description",
							Category: &CommandCategory{
								Name:  "Expert",
								Order: 1,
							},
						},
					},
				})
				app.AddCommand(Command{
					Name: "command3",
					Category: &CommandCategory{
						Name:  "Intermediate",
						Order: 2,
					},
					Description: `Description for command3 goes here `,
				})

				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "command2"},
			wantErr: true,
		},
		{
			// Action propagation with nested  commands
			// Should succeed
			name: "action propagation allowed ",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
					Commands: []*Command{
						{
							Name:             "sub-com",
							Description:      "Optional sub-command",
							ValidationGroups: []string{"subcommand"},
							Action: func(a *Application, c *Command, i interface{}) (interface{}, error) {
								return "passed from sub-com", nil
							},
						},
					},
					Flags: []IFlag{
						&Flag[OneOf]{
							Name:        "output",
							Short:       'o',
							Placeholder: "output format",
							Usage:       "Output format",
							Hints:       []string{"json", "table", "yaml"},
							Default:     "table",
							Required:    false,
						},
					},
					Args: []IArg{
						&Arg[String]{
							Name:             "resource-type",
							Usage:            "type",
							Default:          "",
							ValidationGroups: []string{"name"},
						},
						&Arg[String]{
							Name:             "resource-name",
							Usage:            "name",
							Default:          "",
							ValidationGroups: []string{"name", "subcommand"},
						},
					},
					Action: func(a *Application, c *Command, i interface{}) (interface{}, error) {
						passed_down := i.(string)
						action_result = "received " + passed_down
						return nil, nil
					},
				})

				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "command1", "-o", "json", "sub-com", "my_name"},
			wantErr: false,
			check: func(a *Application) error {
				if action_result != "received "+"passed from sub-com" {
					return errors.New("data did not propagate fromsub command ")
				}
				return nil
			},
		},
		{
			// Action propagation not allowed with nested  commands
			// Should succeed
			name: "action propagation not allowed ",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
					Commands: []*Command{
						{
							Name:             "sub-com",
							Description:      "Optional sub-command",
							ValidationGroups: []string{"subcommand"},
							Action: func(a *Application, c *Command, i interface{}) (interface{}, error) {
								a.Stop()
								action_result = "passed from sub-com"
								return action_result, nil
							},
						},
					},
					Flags: []IFlag{
						&Flag[OneOf]{
							Name:        "output",
							Short:       'o',
							Placeholder: "output format",
							Usage:       "Output format",
							Hints:       []string{"json", "table", "yaml"},
							Default:     "table",
							Required:    false,
						},
					},
					Args: []IArg{
						&Arg[String]{
							Name:             "resource-type",
							Usage:            "type",
							Default:          "",
							ValidationGroups: []string{"name"},
						},
						&Arg[String]{
							Name:             "resource-name",
							Usage:            "name",
							Default:          "",
							ValidationGroups: []string{"name", "subcommand"},
						},
					},
					Action: func(a *Application, c *Command, i interface{}) (interface{}, error) {
						// shoudl never be called
						passed_down := i.(string)
						action_result = "received " + passed_down
						return nil, nil
					},
				})

				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "command1", "-o", "json", "sub-com", "my_name"},
			wantErr: false,
			check: func(a *Application) error {
				if action_result != "passed from sub-com" {
					return errors.New("data did not propagate fromsub command ")
				}
				return nil
			},
		},
		{
			// Usinf flag validation to set global optioosn like log level
			// Should succeed
			name: "set log level app wide",
			setup: func() *Application {
				app := New()
				app.Description = `{{.Name}} is a test program for gocli`
				app.AddCommand(Command{
					Name:        "command1",
					Description: `{{.FullCommand}} first command`,
					Flags: []IFlag{
						&Flag[OneOf]{
							Name:        "output",
							Short:       'o',
							Placeholder: "output format",
							Usage:       "Output format",
							Hints:       []string{"json", "table", "yaml"},
							Default:     "table",
							Required:    false,
						},
						&Flag[OneOf]{
							Name:    "log-level",
							Short:   'l',
							Usage:   `Specify verbosity level.`,
							Hints:   []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"},
							Default: "ERROR",
							Validator: func(a *Application, f IFlag) error {
								ls := f.GetValue().(string)
								level, _ := log.ParseLevel(ls)
								log.SetLevel(level)
								return nil
							},
						},
					},
					Args: []IArg{
						&Arg[String]{
							Name:             "resource-type",
							Usage:            "type",
							Default:          "",
							ValidationGroups: []string{"name"},
						},
						&Arg[String]{
							Name:             "resource-name",
							Usage:            "name",
							Default:          "",
							ValidationGroups: []string{"name"},
						},
					},
					Action: func(a *Application, c *Command, i interface{}) (interface{}, error) {
						log.Errorln("Printing error message ")
						log.Traceln("Printing trace message")
						return nil, nil
					},
				})

				app.Terminator = NilTerminator

				return app
			},
			args:    []string{"test", "command1", "-l", "TRACE", "-o", "json", "my_type", "my_name"},
			wantErr: false,
			check: func(a *Application) error {
				// check log level

				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			t.Logf("Running test %s", tt.name)
			a := tt.setup()

			if err := a.Run(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("Application.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.check != nil {
				tt.check(a)
			}
		})
	}

	// cleanup
	for _, f := range expected_files {
		os.Remove(f)
	}
}
