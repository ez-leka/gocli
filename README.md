# GOCLI - customizable and localized command line parser and processor

- [Overview](#overview)
- [Features](#features)
- [Reference Documentation](#reference)
  - [Commands](#commands)
  - [Flags and Arguments Types](#flags-and-arguments-types)
  - [Flags and Arguments Validation](#flags-and-arguments-validation)
  - [Actions](#actions)
  - [Templates And Localization](#templates-and-localization)
## Overview
gocli is a fully customizable and localizable CLI parser/processor that suports nested commands, positioned arguments, short and long flags and command grouping

To install
```
go get https://github.com/ez-leka/gocli
```

To use 

```go
    app := gocli.New("en_us")
    app.Description = `{{.Name}} is a test program for gocli`

    app.AddCommand(gocli.Command{
            Name:        "create",
            Description: "Create one or many resources from a file",
            Usage: `
            
            JSON and YAML file formats are accepted.
            
           
            Examples:
                # Create a function using the data in function.json
                {{.FillCommand}} -f ./function.json
            
                # Create a function using the data in function.yaml
                 {{.FillCommand}} -f ./function.yaml
                
                `,
            Flags: []gocli.IFlag{
                &gocli.Flag[gocli.List]{
                    Name:    "filename",
                    Usage:   "yaml or json file (or files, if wildcard) identifying resources to delete",
                    Default: "",
                },
            },
            Action: func(*Application, *Command) error {
                fmt.Println("Running command create")
            }
        })
    app.ShowHelpCommand = true

    err = app.Run(os.Args)
```
More examples can be found in cmd/main.go 
## Features

- Generated help output and error reporting that can be fully [customized and localized](#templates-and-localization)
- Type-safe flags and arguments
- Support for required flags and required positional arguments 
- Support for nested commands
- flags and arguments can be grouped for either/or validation
- Validation callbacks for flags and positional arguments
- POSIX-style short flag combining (`-a -b` -> `-ab`).
- Short-flag+parameter combining (`-f file` -> `--ffile` or `-f=file`)
- Long parameters with or without '=' (`--file filename` or `--file=filename`)
- Flags and arguments can be in any order unless specified otherwise

## Reference
### Commands
Commands can be grouped into categories for help printout. If some commands do have category and some do not, those without a category will appper in the list as "Miscellanuious Commands" (this can be changed by changing corresponding template - see Templates). If no command has category, all commands will be listed alphabetically under "Commands" (this can be changed by changing corresponding template - see Templates)

Commands can have  positined arguments and non-optional subcommands.

If command needs to have either positined argument or sub-command, subcommand MUST have Optional:true and a Validation group. All agruments of parent command(s) that are still relevant to this subcommand must have this validation group listed in their own ValidationGroups. Arguments specified as part of  the subcommand definition do not require validation group to be set and will be added after all relevant arguments of parent command(s)

The following example shows update command with flag/argument grouping and custom templated usage 

```go
app.AddCommand(gocli.Command{
		Name:        "update",
		Description: "update resources by file names, stdin, resources and names",
		Usage: `
			
			JSON and YAML formats are accepted. Only one type of argument may be specified: file names or resources and names,
		
		Examples:
			# Update a resource using the type and name specified in resourse.json
			{{.FullCommand}} -f ./resourse.json
		
			# Update resourse with names "baz" and "foo"
			{{.FullCommand}} resourse baz,foo
			 
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
				Usage:            "One of node, user or function",
				Hints:            []string{"node(s)", "function(s)", "user(s)"},
				ValidationGroups: []string{"resourse"},
			},
			&gocli.Arg[[]gocli.String]{
				Name:             "resource-name",
				Usage:            "Name of the resource",
				ValidationGroups: []string{"resourse"},
			},
		},
	})
```

The help generaged for this command will look like 
```
update resources by file names, stdin, resources and names
Flags:
    -o, --output      output format                                                                                                                                                                                            
    -h, --help        Show context-sensitive help                                                                                                                                                                              
    -v, --version     show version                                                                                                                                                                                             
    -f, --filename    Filename, directory, or URL to files to use to delete resources                                                                                                                                          

Arguments:
    resource-type    One of node or function                                                                                                                                                                                   
    resource-name    Name of the resource                                                                                                                                                                                      

Usage: 


test update [ -h -v ]  [ -o[=]<OUTPUT> ] ( [ -f[=]<FILENAME> ] |  [<resource-type><resource-name> ] )

			
		JSON and YAML formats are accepted. Only one type of argument may be specified: file names or resources and names,
		
		Examples:
			# Update a resource using the type and name specified in resourse.json
			test update -f ./resourse.json
		
			# Update resourse with names "baz" and "foo"
			test delete resourse baz,foo
``` 
### Flags and Arguments Types

Flags and argumens can be a single value or cumulative, alowing for multiple values for a given flag or apositined argument. 

#### Single Value Types

- String (`Flag[String]{}`) - regular string without any additional validation. The value can be retrieved using `app.GetStringArg(<argument name>)` or `app.GetStringFlag(<flag name>)`
- Bool (`Flag[Bool]{}`) - boolean flag (is not applicable to Arguments). The value can be retrived by `app.GetBoolFlag(<flag name>)`
- OneOf  (`Flag[OneOf]{}`) - restricted string aflag or argument. Such flag orargument MUST have Hists array specifying set of possible values. For example:
```go
		Flags: []gocli.IFlag{
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

```
- Email  (`Flag[Email]{}`) - value of the flag or argument of this type must validate as a valid email. The value can be retrieved using `app.GetStringArg(<argument name>)` or `app.GetStringFlag(<flag name>)`
- File (`Flag[File]{}`) - value of the flag or argument of this type must validate as existing file path to a file or directory. If path contains wildcard, validation will make sure thatat least one match exist. The value can be retrieved using `app.GetStringArg(<argument name>)` or `app.GetStringFlag(<flag name>)`

To retrieve the value use `app.GetStringArg(<argument name>)` or `app.GetStringFlag(<flag name>)`

On command line long name flags appear as `--boolflag`, `--logflag <value>` or `--longflag=<value>`, short flags are optional version of a given long flag and can appear on command line as 
`-f test.txt`, `-ftest.txt`, or `-f=test.txt`. Short flags can be combined together and all but last of combined flags MUST be boolean. Last flag in combination can be a regular flag of any type: 
`vaftest6.txt`, `-vaf=test6.txt`, `-vaf test7.txt`

#### Cumulative types 

All non-booles types can be cumulative and are sopecified as slice of the desired undelying type. For example, cumulative file flag -f  can be specified as `Flag[[]File]{}` and cumulative string argument asn  (`Arg[[]String]{}`)

To retrieve value of cumulative flags and arguments, use `app.GetListArg(<argument name>)` or `app.GetListFlag(<flag name>)`

Cumulative flag can appear on the command line multiple times, for example, `test -f file1 -f file2`.

Since arguments are positined, to pass multiple values to an argument, use comma separated list. For example, `test user1,user2`
To retrieve value, use `app.GetListArg(<argument name>)` or `app.GetListFlag(<flag name>)`

If argument is cumulative and is a last positined argument, all remaining values from command line will be consumed by this argument. Considerind example of the [command](#command), update command line may look like:
```
update user user1,user2 <----- using comma separated values for cumulative argument resource-name
update user user1 user2 <----- consuming the rest of the arguments
```

## Flags and Arguments Validation

Flag and argumentd are first validated agains their type(see [Flags and Arguments Types](#flags-and-arguments-types))

Flags and arguments can be assigned ValidationGroup. If validationGroup is specified, only flags and arguments that belong to a single group or are ungrouped can be present on command line. 

Consider the case  of delete command that can accept either a file, listing objects to delete or have objects listed directly on command line. In addition, there is optional output format flag that applies to both cases.

Usage for such commanfd will look like the following:
```
test delete (-f <filename> [-f <filename>] | <type> [<names>]) [-o <format>]
```
and calls can look like:
```
test delete -f resourse.json
test delete -f resourse.json -f resourse2.json -f resourse3.json -o table
test delete users  <------------------------ all users since no names specified as names parameter is optional
test delete user user1
test delete user user2,user3 -o json
```
We can define that command as following:
```go 
	deleteCmd := gocli.Command{
		Name:        "delete",
		Description: "delete resources by file names, resources and names",
		Flags: []gocli.IFlag{
			&gocli.Flag[[]gocli.File]{
				Name:             "filename",
				Short:            'f',
				Usage:            "Filename, directory, or URL to files to use to delete resources",
				Default:          "",
                Required:         true,
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
				Usage:            "One of user, organization, group",
				Hints:            []string{"user(s)", "organization(s)", "group(s)"},
				ValidationGroups: []string{"resourse"},
                Required:         true
			},
			&gocli.Arg[[]gocli.String]{
				Name:             "resource-name",
				Usage:            "Name of the resource",
				ValidationGroups: []string{"resourse"},
			},
		},
	}

```


In addition, custom validators can be specifies if an additional logic to validate flags and arguments is required. Custom Validator functions are called after command line flags and arguments have been parsed, required type validation passed for all but before required flags and arguments are validated. This function can reset required value and any other public property of this flag according to all other parsed flags and arguments as needed

After custom validators, the required flags and arguments validated, so you can feel free to change whether the flag or argument is required in castom validators. 

## Actions 

All actions in the command chain will be executed in reverse order : current command, it parent, and so on up to and including application action
Each action, except the very first one in the chain can accepts returned interface{} from  previous command. 

Consider command chain:
```
    user
        create
        update
        delete

```
Action for commands create, update and delete will be executed first. if each of those actions return user struct, the user command action will receive that data and take ocare of printing it uniformly so you only need to implement formatting and printing on one place

## Templates And Localization
Any and all strings in gocli can be customized and/or localized. 

Default language ot the library is en_us. 

If you want to customize an entry withing en_us localization or translate strings to another language:

	app.AddTranslation("en_us", i18n.Entries{
        "VersionFlagUsageTemplate": "show git SHA", // original value was `show version`
	})

	app.AddTranslation("ru", i18n.Entries{
		"HelpCommandAndFlagName": "Помощь",
		"Flags":                  "Флаги",
		"Arguments":              "Аргументы",
		"Usage":                  "Использование",
        "HelpFlagUsageTemplate": "распечатать информацию о флагах",
	})

To generate full set of translatable entries, add teh following directive to one of your main.go 
```go 
//go:generate go run translate.go <language>
```

the file <language>-strings.go will be genetated in the directory translations

```
    AppUsageTemplate = <actual app usage template>
	HelpCommandUsageTemplate    = `show help`
	HelpCommandArgUsageTemplate = `show help for <command>`
	VersionFlagUsageTemplate    = `show version`
	HelpFlagUsageTemplate       = `Show context-sensitive help`

	"Error":                         "Error: %s",
	"FlagLongExistsTemplate":        `flag --{{.Name}} already exists`,
	"FlagShortExistsTemplate":       `flag -{{.Short}} already exists`,
	"UnknownElementTemplate":        `unknown {{.GetType}} {{.GetPlaceholder}}`,
	"UnexpectedFlagValueTemplate":   `expected argument for flag --{{.Element.Name}} {{if .Element.Short}}(-{{.Element.Short|Rune}}{{end}}) {{if .Extra}got '{{.Extra}}'{{end}}}}`,
	"UnexpectedTokenTemplate":       `expected {{.Extra}} but got {{.Name}}`,
	"WrongElementTypeTemplate":      `wrong {{.Element.GetType}} type`,
	"FlagAlreadySet":                `flag {{.GetName}} already have been set. This flag is not cumulative and can only appear once on command line`,
	"NoHintsForOneOf":               `no hints speciffied for {{.GetType}} {{.GetName}}`,
	"UnknownOneOfValue":             `unsupported value {{.Extra}} for {{.Element.GetType}} {{.Element.GetPlaceholder}}`,
	"InvalidTimeFormat":             `invalid timestamp string {{.Extra}} for {{.Element.GetType}} {{.Element.GetPlaceholder}}`,
	"MissingRequiredFlag":           `required {{.GetType}} --{{.Name}}{{if .Short}}(-{{.Short|Rune}}){{end}} is missing `,
	"MissingRequiredArg":            `required {{.GetType}} {{.GetPlaceholder}} is missing `,
	"FlagsArgsFromMultipleGroups":   `either {{.Name}} or {{.Extra}} can be specified, but not both`,
	"NoUniqueFlagArgCommandInGroup": `must specify flag, argument or command. Try --help`,
	"FlagValidationFailed":          `Invalid flag value {{.Extra}} for flag --{{.Element.Name}}{{if .Element.Short}}(-{{.Element.Short|Rune}}{{end}})`,
	"CommandRequired":               `Command required. Try --help`,
	"FormatCommandsCategory":        "Commands",
	"FormatMisCommandsCategory":     "Miscellaneous Commands",
	"FormatFlagWithShort":           "-%c, --%s",
	"FormatFlagNoShort":             "--%s",
	"FormatFlagShort":               "-%c",
	"FormatArg":                     "%s",
```
Description and Usage values for commands, flags and arguments can be a go tamplate and can only refer to its own object. 

