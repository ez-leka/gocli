package main

import (
	"fmt"

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
				resource_type, _ := a.GetOneOfArg("resourse-type")
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
			Name:    "filename",
			Usage:   "yaml or json file (or files, if wildcard) identifying resources to delete",
			Default: "",
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

func main() {

	testHelpFlag()
	// testHelpCommand()
	//testFlagArgumentParsing()
	//testValidationGrouping()

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

	args := []string{"test", "create"}
	// args := []string{"test", "delete", "--Помощь"}
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

func testFlagArgumentParsing() {
	app := makeApp()
	app.Version = "v1.2.3"
	var err error

	args := []string{"test", "get", "nodes", "-e", "my@gnode.org", "-ftest25.txt", "-f=test2.txt", "-f", "test3.txt", "--filename=test4.txt", "--filename", "test5.txt", "-vaftest6.txt", "-vaf=test6.txt", "-vaf", "test7.txt", "-vfa", "test8.txt", "node1,node2", "node3"}
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
