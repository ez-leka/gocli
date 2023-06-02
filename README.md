# Overview
gocli is a fully customizable and localizable CLI parser/processor that suports nested commands, positioned arguments, short and long flags and command grouping for Usage 

To install
```
go get https://github.com/ez-leka/gocli
```

To use 

```
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
## Commands
Commands can be grouped into categories for help printout. If some commands do have category and some do not, those without a category will appper in the list as "Miscellanuious Commands" (this can be changed by changing corresponding template - see Templates). If no command has category, all commands will be listed alphabetically under "Commands" (this can be changed by changing corresponding template - see Templates)

## Flags
--logflag <value> or --longflag=<value>
-f test.txt , -ftest.txt , -f=test.txt , -cv, -cvftest.txt, -cvf test.txt where c and v are boolean flags are possibble combinations

Non-boolean flags can be cumulative, i.e flag can be present multiple times on commanfd line by specifying either List or OnOfList as flag type. however, unlike agruments, comma-separated values are not supported 

Flag can have a Validator function that is called after command line flags and argumets have been parsed but before required flags and arguments are validated. This function can reset required value and any other public property of this flag according to all other parsed flags and arguments as needed

## Arguments 

arguments are positioned and can be located anywhere in the command line 
arguments cannot be combined wth sub commands 

arguments with hints are trited as enumeration; only one of listed values can be used as argument value. If you have to support single and plural value, i.e function and functions are valif and interchangable, you can specify hint as function(s). Note: app.GetEnumArg(name) will return single version of value regardless wheteher user specified singe or plural form

If you want to make flag required based on argument or arguments, you can supply Validator function for a flag and set its Required() according to all other flags and argumnts set by user or default values

An argument can be cumulative, i.e allow multiple values. This can be achieved the following ways: 

- by specifying List as Arg type, multiple values of the flag are expected as a comma-separated list:
    cmd foo,bar
- by specifying OneOfList as Arg type, multiple values of the flag are expected as a comma-separated list with only values from specified hints possible

    cmd foo,bar

- if argument type is List and this is last defned argument, the argument will accumulate all  the rest of arguments on command line, as shown below

    ip 10.10.10.1,10.10.10.2 10.10.10.3 

will put all 3 Ip addreses into single list argument 

Arguments can have a Validator function that is called after command line flags and argumets have been parsed but before required flags and arguments are validated. This function can reset required value and any other public property of this argument according to all other parsed flags and arguments as needed

## Validation 

### Validation Grouping
For validation purposes flags and arguments can be groupped, meanign only once from the same group can be used together, i.e. groups are mutially exclusive. For example, 

```
test update (-f FILENAME | <resource-type> [<resource-name>])
```

command can either have -f flag or arguments but not both. Other flags may be applicable to both cases

This can ve achived by implementing csutom validator on the command, arguments and flags  or, much easier by defining validation grouping:



### Validation order
argument validators 
flag validators
validate for required flags and arguments 
All validators in the command chain will be executed in reverse order : current command, it parent, and so on up to and including application action


## Actions 

All actions in the command chain will be executed in reverse order : current command, it parent, and so on up to and including application action


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



```
    AppUsageTemplate = <actual app usage template>
	HelpCommandUsageTemplate    = `show help`
	HelpCommandArgUsageTemplate = `show help for <command>`
	VersionFlagUsageTemplate    = `show version`
	HelpFlagUsageTemplate       = `Show context-sensitive help`

	FlagLongExistsTemplate        = `flag --{{.Name}} already exists`
	FlagShortExistsTemplate       = `flag -{{.Short}} already exists`
	MixArgsCommandsTemplate       = `can't mix Arg()s with Command()s`
	UnknownArgument               = `unknown argument {{.Name}}`
	UnknownFlagTemplate           = `unknow {{.Prefix}}{{.Name}} flag`
	UnexpectedFlagValueTemplate   = `expected argument for flag {{.Prefix}}{{.Name} {{if .Extra}got '{{.Extra}}'{{end}}}}`
	UnexpectedTokenTemplate       = `expected {{.Extra}} but got {{.Name}}`
	WrongFlagArgumentTypeTemplate = `wrong {{.GetType}} type`
	FlagAlreadySet                = `flag {{.GetName}} already have been set. This flag is not cumulative and can only appear once on colland line`
	NoHintsForEnumArg             = `no hints speciffied for argument {{.GetName}}`
	UnknownArgumentValue          = `unsupported value {{.Extra}} for argument {{.Name}}`
	MissingRequired               = `required {{.GetType}} --{{.Name}}{{.Short|Rune}} is missing `

	FormatCommandsCategory    = "Commands"
	FormatMisCommandsCategory = "Miscellaneous Commands"
	FormatFlagWithShort       = "-%c, --%s"
	FormatFlagNoShort         = "--%s"
	FormatArg                 = "%s"
```
Description and Usage values for commands, flags and arguments can be a go tamplate and can only refer to its own object. 

