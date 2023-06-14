package main

import (
	"fmt"
	"net/mail"
	"os"

	"github.com/ez-leka/gocli"
	"github.com/ez-leka/gocli/i18n"
	"golang.org/x/text/language"
)

var app_usage = `
	test command line with flags and arguments
`
var getCmd = gocli.Command{
	Name:        "get",
	Alias:       []string{"list"},
	Description: "Display one or many resources",
	Usage: `
		test get [(-o|--output=)json|yaml|table] <resource-type> [<resource-name>]
	
	Examples:
		# List all resourses in specified format
		test get function -o json
	
		`,
	Flags: []gocli.IFlag{
		&gocli.Flag[[]gocli.File]{
			Name:    "filename",
			Short:   'f',
			Usage:   "yaml or json file (or files, if wildcard) identifying resources to delete",
			Default: "",
		},
		&gocli.Flag[gocli.Email]{
			Name:    "email",
			Short:   'e',
			Usage:   "email",
			Default: "",
		},
		&gocli.Flag[gocli.Bool]{
			Name:    "all",
			Short:   'a',
			Usage:   "yaml or json file (or files, if wildcard) identifying resources to delete",
			Default: "",
		},
		&gocli.Flag[gocli.String]{
			Name:    "to",
			Short:   't',
			Usage:   "timestamp ",
			Default: "",
			Validator: func(a *gocli.Application, f gocli.IFlag) error {
				resource_type, _ := a.GetStringArg("resourse-type")
				if resource_type == "metrics" {
					f.SetRequired(true)
				}
				return nil
			},
		},
	},
	Args: []gocli.IArg{
		&gocli.Arg[gocli.OneOf]{
			Name:     "resourse-type",
			Hints:    []string{"node(s)", "function(s)", "user(s)", "metrics"},
			Required: true,
			Usage:    "type of the resource to get",
		},
		&gocli.Arg[[]gocli.String]{
			Name:  "resource-name",
			Usage: "Name of the resource",
		},
	},
}
var createCmd = gocli.Command{
	Name:        "create",
	Description: "Create one or many resources from a file",
	Usage: `
	
	{{.AppName}} {{.Name}} {{range .Flags }}{{if .GetShort}}-{{.GetShort|Rune}}{{else}}--{{.GetName}}{{end}} <{{.GetName|ToUpper}}> {{if .IsCumulative}}[{{if .GetShort}}-{{.GetShort}}{{else}}--{{.GetName}}{{end}} <{{.GetName|ToUpper}}> ...]{{end}}{{end}}

	JSON and YAML file formats are accepted.
	
	
	Examples:
		# Create a function using the data in function.json
		{{.AppName}} {{.Name}} -f ./function.json
	
		# Create a function using the data in function.yaml
		{{.AppName}} {{.Name}} -f ./function.yaml
		
		# Create a function from command line arguments
		{{.AppName}} {{.Name}} function func1 --image kontainapp/pytorch-demo-cpu --port 8080 --shapshot --warmup_urls '{"url":"/", "method":"GET"}'    
		`,
	Flags: []gocli.IFlag{
		&gocli.Flag[[]gocli.String]{
			Name:             "filename",
			Short:            'f',
			Usage:            "yaml or json file (or files, if wildcard) identifying resources to delete",
			Default:          "",
			Required:         true,
			ValidationGroups: []string{"file"},
		},
		&gocli.Flag[gocli.OneOf]{
			Name:     "output",
			Short:    'o',
			Usage:    "Output format",
			Hints:    []string{"json", "table", "yaml"},
			Default:  "table",
			Required: false,
		},
	},
	Commands: []*gocli.Command{
		{
			Name:             "OptionalCommand",
			Optional:         true,
			ValidationGroups: []string{"opt_cmd1"},
			Flags: []gocli.IFlag{
				&gocli.Flag[gocli.String]{
					Name:             "port",
					Short:            'p',
					Usage:            "port",
					Default:          "8080",
					Required:         true,
					ValidationGroups: []string{"opt_cmd1"},
				},
			},
			Args: []gocli.IArg{
				&gocli.Arg[gocli.String]{
					Name:             "resource-name",
					Required:         true,
					ValidationGroups: []string{"opt_cmd1"},
				},
			},
		},
	},
}

var deleteCmd = gocli.Command{
	Name:        "delete",
	Description: "Delete resources by file names, stdin, resources and names",
	Usage: `
		
	JSON and YAML formats are accepted. Only one type of argument may be specified: file names or resources and names,
	
	Examples:
		# Delete a function using the type and name specified in function.json
		{{.FullCommand}} -f ./function.json
	
		# Delete function with names "baz" and "foo"
		{{.FullCommand}} function baz,foo
	
		# Delete resources from all files that end with '.json' - i.e. expand wildcard characters in file names
		{{.FullCommand}} -f '*.json'
		
		# Delete a pod based on the type and name in the JSON passed into stdin
		cat pod.json |{{.FullCommand}} -f -
		 
		`,
	Flags: []gocli.IFlag{
		&gocli.Flag[[]gocli.String]{
			Name:             "filename",
			Usage:            "Filename, directory, or URL to files to use to delete resources",
			Default:          "",
			ValidationGroups: []string{"files"},
			Required:         true,
		},
		&gocli.Flag[gocli.OneOf]{
			Name:     "output",
			Short:    'o',
			Usage:    "Output format",
			Hints:    []string{"json", "table", "yaml"},
			Default:  "table",
			Required: false,
		},
	},
	Args: []gocli.IArg{
		&gocli.Arg[gocli.String]{
			Name:             "resource-type",
			Usage:            "Type of the resource - one of {{.GetHints}}",
			Hints:            []string{"node(s)", "function(s)", "user(s)", "metrics"},
			ValidationGroups: []string{"resourses"},
			Required:         true,
			Placeholder:      "type",
		},
		&gocli.Arg[[]gocli.String]{
			Name:             "resource-name",
			Usage:            "Name of the resource",
			ValidationGroups: []string{"resourses"},
			Placeholder:      "foo|foo,bar",
		},
	},
}

var configCmd = gocli.Command{
	Name:        "config",
	Description: "Manage kontain config file",
	Commands: []*gocli.Command{
		{
			Name:        "set",
			Description: "change configuration setting",
			Usage: `
			Example:
				{{.FullCommand}} admin@company.com "SecretUYIUYIUTYTUYTUYT" https://faas.kontain.app:8443
				`,
			Args: []gocli.IArg{
				&gocli.Arg[gocli.Email]{
					Name:     "email",
					Usage:    "--email <email used to register>",
					Required: true,
					Validator: func(a *gocli.Application, arg gocli.IArg) error {
						email := arg.GetValue().(string)
						_, err := mail.ParseAddress(email)
						return err
					},
				},
				&gocli.Arg[gocli.String]{
					Name:     "secret",
					Usage:    "---secret <API Key secret provided to you>",
					Required: true,
				},
				&gocli.Arg[gocli.String]{
					Name:     "url",
					Usage:    "--url <Server url> - https://faas.kontain.app:8443",
					Required: true,
				},
			},
		},
		{
			Name:        "verify",
			Description: "Verify credentials and connectivity to server",
			Usage: `
			{{.FullCommand}}
	
			Use this command to verify your credentials and server connectivity configured via kctl configure 
			`,
		},
		{
			Name:        "view",
			Description: "Display kontain config settings",
			Usage: `
			{{.FullCommand}}
	
			Use this command to verify your credentials and server connectivity configured via kctl configure 
			`,
			Flags: []gocli.IFlag{
				&gocli.Flag[gocli.OneOf]{
					Name:     "output",
					Short:    'o',
					Usage:    "Output format",
					Hints:    []string{"json", "table", "yaml"},
					Default:  "table",
					Required: false,
				},
			},
		},
	},
}

var argsAndCommands = gocli.Command{
	Name:        "get",
	Alias:       []string{"list"},
	Description: "Display one or many resources",
	Usage: `

	Examples:
		# List all resourses in specified format
		{{.FullCommand}} function -o json

	`,
	Commands: []*gocli.Command{
		{
			Name:             "metrics",
			Description:      "Get time metrics and usage information for a function",
			ValidationGroups: []string{"metrics"},
			Optional:         true,
			Flags: []gocli.IFlag{
				&gocli.Flag[gocli.String]{
					Name:    "from",
					Short:   'f',
					Usage:   "timestamp ",
					Default: "",
				},
				&gocli.Flag[gocli.String]{
					Name:    "to",
					Short:   't',
					Usage:   "timestamp ",
					Default: "",
				},
			},
			Validator: func(a *gocli.Application, c *gocli.Command) error {
				// metrics require resource name
				arg, _ := a.GetArgument("resource-name")
				arg.SetRequired(true)
				return nil
			},
			Action: func(app *gocli.Application, cmd *gocli.Command, data interface{}) (interface{}, error) {
				fmt.Println("Will get metrics and exit - no propagation")
				os.Exit(0)
				return nil, nil
			},
		},
	},
	Flags: []gocli.IFlag{
		&gocli.Flag[gocli.OneOf]{
			Name:     "output",
			Short:    'o',
			Usage:    "Output format",
			Hints:    []string{"json", "table", "yaml"},
			Default:  "table",
			Required: false,
		},
	},
	Args: []gocli.IArg{
		&gocli.Arg[gocli.OneOf]{
			Name:             "resource-type",
			Usage:            "type of the resourse to get",
			Hints:            []string{"user(s)", "function(s)", "node(s)"},
			Required:         true,
			ValidationGroups: []string{"top"},
		},
		&gocli.Arg[[]gocli.String]{
			Name:             "resource-name",
			Usage:            "Name of the resourse",
			Required:         false,
			ValidationGroups: []string{"top", "metrics"},
		},
	},
	Action: func(app *gocli.Application, cmd *gocli.Command, data interface{}) (interface{}, error) {
		// resource_type, err := app.GetStringArg("resource-type")
		// if err != nil {
		// 	return nil, err
		// }
		// resource_name, err := app.GetListArg("resource-name")
		// if err != nil {
		// 	return nil, err
		// }

		return nil, nil
	},
}

func main() {

	//testHelpFlag()
	// testHelpCommand()
	// testFlagArgumentParsing()
	// testValidationGrouping()
	// testOptionalCommand()
	testUngroupedCommand()
	testMixOfArgsAndCommands()

}
func makeApp() *gocli.Application {
	app := gocli.New("en_us")

	app.Description = `{{.Name}} is a test program for gocli`

	app.AddCommand(getCmd)
	app.AddCommand(createCmd)
	app.AddCommand(deleteCmd)
	// prevent usage from exiting so we can run multiple tests
	//app.Terminate(nil)

	return app
}

func testHelpFlag() {

	app := gocli.New("ru")

	app.Description = `{{.Name}} is a test program for gocli`

	app.AddTranslation(language.MustParse("en_us"), i18n.Entries{
		"VersionFlagUsageTemplate": "show git SHA",
	})

	app.AddTranslation(language.MustParse("ru"), i18n.Entries{
		"HelpCommandAndFlagName": "Помощь",
		"Flags":                  "Флаги",
		"Arguments":              "Аргументы",
		"Usage":                  "Использование",
		"HelpFlagUsageTemplate":  "распечатать информацию о флагах",
	})
	app.AddCommand(getCmd)
	app.AddCommand(createCmd)
	app.AddCommand(deleteCmd)
	app.Version = "v1.2.3"

	args := []string{"test", "delete", "--Помощь"}
	err := app.Run(args)
	if err != nil {
		fmt.Println("FAILED:", err)
	} else {
		fmt.Println("PASSED")
	}
}
func testHelpCommand() {
	app := makeApp()

	// add help command to app
	app.ShowHelpCommand = true

	var args []string
	var err error

	// test --help

	args = []string{"test", "help", "create"}
	err = app.Run(args)
	if err != nil {
		fmt.Println("FAILED:", err)
	} else {
		fmt.Println("PASSED")
	}
}

func testUngroupedCommand() {
	app := makeApp()
	app.Terminate(nil)

	app.AddCommand(configCmd)

	args := []string{"test", "config"}
	fmt.Println(args)
	err := app.Run(args)
	// should fail because no required command
	if err != nil {
		fmt.Println("PASSED")
	} else {
		fmt.Println("FAILED")
	}
	fmt.Println("------------------------")

	args = []string{"test", "config", "set"}
	fmt.Println(args)
	err = app.Run(args)
	// should fail because no required args
	if err != nil {
		fmt.Println("PASSED")
	} else {
		fmt.Println("FAILED")
	}
	fmt.Println("------------------------")

}

func testMixOfArgsAndCommands() {
	app := makeApp()
	app.Terminate(nil)

	app.AddCommand(argsAndCommands)

	args := []string{"test", "get"}
	fmt.Println(args)
	err := app.Run(args)
	// should fail - missing required arg or command
	if err != nil {
		fmt.Println("PASSED")
	} else {
		fmt.Println("FAILED")
	}
	fmt.Println("------------------------")

	args = []string{"test", "get", "function", "fun1"}
	fmt.Println(args)
	err = app.Run(args)
	// should pass
	if err != nil {
		fmt.Println("FAILED")
	} else {
		fmt.Println("PASSED")
	}
	fmt.Println("------------------------")

	args = []string{"test", "get", "metrics", "fun1"}
	fmt.Println(args)
	err = app.Run(args)
	// should pass
	if err != nil {
		fmt.Println("FAILED")
	} else {
		fmt.Println("PASSED")
	}
	fmt.Println("------------------------")

	args = []string{"test", "get", "metrics", "fun1", "--from", "01/01/2023 04:01:00 PM"}
	fmt.Println(args)
	err = app.Run(args)
	// should pass
	if err != nil {
		fmt.Println("FAILED")
	} else {
		fmt.Println("PASSED")
	}
	fmt.Println("------------------------")

	args = []string{"test", "get", "function", "metrics"}
	fmt.Println(args)
	err = app.Run(args)
	// should pass  - metrics becomes a function name as functions and commands canno tbe inrespursed
	if err != nil {
		fmt.Println("FAILED")
	} else {
		fmt.Println("PASSED")
	}
	fmt.Println("------------------------")

	args = []string{"test", "get", "metrics", "func1", "test"}
	fmt.Println(args)
	err = app.Run(args)
	// should pass - test is optional sub-command specific argument
	if err != nil {
		fmt.Println("FAILED")
	} else {
		fmt.Println("PASSED")
	}
	fmt.Println("------------------------")

	args = []string{"test", "get", "jobs"}
	fmt.Println(args)
	err = app.Run(args)
	// should fail - unknown argument value
	if err != nil {
		fmt.Println("PASSED")
	} else {
		fmt.Println("FAILED")
	}
	fmt.Println("------------------------")

}
func testOptionalCommand() {
	app := makeApp()
	app.Terminate(nil)

	args := []string{"test", "create"}
	fmt.Println(args)
	err := app.Run(args)
	// should fail because either command or -f flag required
	if err != nil {
		fmt.Println("PASSED")
	} else {
		fmt.Println("FAILED")
	}
	fmt.Println("------------------------")

	args = []string{"test", "create", "-fcmd"}
	err = app.Run(args)
	fmt.Println(args)
	if err != nil {
		fmt.Println("FAILED:", err)
	} else {
		fmt.Println("PASSED")
	}
	fmt.Println("------------------------")

	args = []string{"test", "create", "OptionalCommand"}
	fmt.Println(args)
	err = app.Run(args)
	// shoudl fail as required argument is missing
	if err != nil {
		fmt.Println("PASSED")
	} else {
		fmt.Println("FAILED:", err)
	}
	fmt.Println("------------------------")

	args = []string{"test", "create", "OptionalCommand", "func1"}
	fmt.Println(args)
	err = app.Run(args)
	// shoudl fail - required flag --port is missing
	if err != nil {
		fmt.Println("PASSED:")
	} else {
		fmt.Println("FAILED")
	}
	fmt.Println("------------------------")

	args = []string{"test", "create", "OptionalCommand", "func1", "-fcmd"}
	fmt.Println(args)
	err = app.Run(args)
	// should fail - both groups used
	if err != nil {
		fmt.Println("PASSED:")
	} else {
		fmt.Println("FAILED")
	}
	fmt.Println("------------------------")

}

func testFlagArgumentParsing() {
	app := makeApp()
	app.Version = "v1.2.3"
	var err error

	args := []string{"test", "get", "nodes", "-e", "my@gnode.org", "-fcmd", "-f=test2.txt", "-f", "test3.txt", "--filename=test4.txt", "--filename", "test5.txt", "-vaftest6.txt", "-vaf=test6.txt", "-vaf", "test7.txt", "-vfa", "test8.txt", "node1,node2", "node3"}
	err = app.Run(args)
	if err != nil {
		fmt.Println("FAILED:", err)
	} else {

		files, _ := app.GetListFlag("filename")
		fmt.Println(files)
		fmt.Println("PASSED")
	}
}

func testValidationGrouping() {
	app := makeApp()
	app.Version = "v1.2.3"
	var err error

	updateCmd := gocli.Command{
		Name:        "update",
		Description: "update resources by file names, stdin, resources and names",
		Usage: `
			
		JSON and YAML formats are accepted. Only one type of argument may be specified: file names or resources and names
		
		Examples:
			# Update a resource using the type and name specified in resourse.json
			test update -f ./resourse.json
		
			# Update resourse with names "baz" and "foo"
			test delete resourse baz,foo
			 
			`,
		Flags: []gocli.IFlag{
			&gocli.Flag[[]gocli.File]{
				Name:             "filename",
				Short:            'f',
				Usage:            "Filename, directory, or URL to files to use to delete resources",
				Default:          "",
				ValidationGroups: []string{"file"},
			},
			&gocli.Flag[gocli.OneOf]{
				Name:    "output",
				Short:   'o',
				Usage:   "output format",
				Hints:   []string{"json", "table", "yaml"},
				Default: "table",
				// this flag does not specify validation group - belongs to all of them
			},
		},
		Args: []gocli.IArg{
			&gocli.Arg[gocli.String]{
				Name:             "resource-type",
				Usage:            "One of node or function",
				Hints:            []string{"node(s)", "function(s)", "user(s)"},
				ValidationGroups: []string{"resourse"},
			},
			&gocli.Arg[[]gocli.String]{
				Name:             "resource-name",
				Usage:            "Name of the resource",
				ValidationGroups: []string{"resourse"},
			},
		},
	}
	app.AddCommand(updateCmd)
	// add help command to app
	app.ShowHelpCommand = true

	// valid case with flag only
	args := []string{"test", "update", "-f", "resourse.json"}
	err = app.Run(args)
	if err != nil {
		fmt.Println("FAILED:", err)
	} else {
		fmt.Println("PASSED")
	}

	// valid case with arguments only
	args = []string{"test", "update", "resourse", "name"}
	err = app.Run(args)
	if err != nil {
		fmt.Println("FAILED:", err)
	} else {
		fmt.Println("PASSED")
	}

	// must fail as both are present
	args = []string{"test", "update", "--help", "-f", "resourse.json", "resourse", "name"}
	err = app.Run(args)
	if err != nil {
		fmt.Println("PASSED:", err)
	} else {
		fmt.Println("FAILED")
	}

	// print help for update command
	args = []string{"test", "help", "create"}
	err = app.Run(args)
	if err != nil {
		fmt.Println("FAILED:", err)
	} else {
		fmt.Println("PASSED")
	}
}
