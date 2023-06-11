package gocli

import "github.com/ez-leka/gocli/i18n"

var GoCliStrings = i18n.Entries{
	"AppUsageTemplate": `
{{- define "CmdFlag"}}
{{- if .GetShort}} -{{.GetShort|Rune}}{{else}} --{{.GetName}}{{end -}}
{{- if not .IsBool}}[=]<{{if .GetPlaceholder}}{{.GetPlaceholder}}{{else}}{{.GetName|ToUpper}}{{end}}>{{end -}}
{{- end -}}

{{- define "CmdArg"}}<{{if .GetPlaceholder}}{{.GetPlaceholder}}{{else}}{{.GetName}}{{end}}>{{- end -}}

{{- define "CmdGroup"}}
{{- range .RequiredFlags}}{{template "CmdFlag" .}}{{end -}}
{{- if .OptionalFlags}} [{{end -}}
{{- range .OptionalFlags}}{{template "CmdFlag" .}}{{end -}}	
{{- if .OptionalFlags}} ]{{end -}}
{{- range .RequiredArgs}}
{{- template "CmdArg" .}}
{{- end -}}
{{- if .OptionalArgs}} [{{end -}}
{{- range .OptionalArgs}}
{{- template "CmdArg" .}}
{{- end -}}	
{{- if .OptionalArgs}} ]{{end -}}
{{- end -}}

{{define "FormatCommandCategory"}}
{{range .}}
{{.Name}}
{{.Commands|CommandsToTwoColumns|TwoColumns}}
{{end}}
{{end}}

{{FormatTemplate .CurrentCommand.Description .CurrentCommand 4}}
{{if .CurrentCommand.Commands}}
  {{template "FormatCommandCategory" .CurrentCommand.Commands|CommandCategories}}
{{end}}
{{if .Flags -}}
{{Translate "Flags"}}:
{{.Flags|FlagsArgsToTwoColumns|TwoColumns}}
{{end -}}
{{if .Args -}}
{{Translate "Arguments"}}:
{{.Args|FlagsArgsToTwoColumns|TwoColumns}}
{{end -}}
{{Translate "Usage"}}: 
{{- $groups := .CurrentCommand.GetValidationGroups -}}
{{- $group_idx := 0}}
{{Indent 4}}{{.CurrentCommand.FullCommand}} 
{{- template "CmdGroup" .CurrentCommand.GetGlobalFlags}} {{template "CmdGroup" $groups.Ungrouped -}} 
{{- if $groups.Groups}} ({{end -}}
  {{- range $groups.Groups}}
  {{- if eq $group_idx 1}} | {{end}} 
  {{- template "CmdGroup" .}}{{- $group_idx = 1}}
  {{- end -}}
  {{- if $groups.Groups}} ){{end -}}
{{- if .CurrentCommand.Commands}} {{Translate "command"}}{{end}}
{{if .CurrentCommand.Usage}}
{{FormatTemplate .CurrentCommand.Usage .CurrentCommand 4}}
{{end}}
`,
	"HelpCommandAndFlagName":      `help`,
	"HelpFlagShort":               `h`,
	"HelpCommandUsageTemplate":    `show help`,
	"HelpCommandArgUsageTemplate": `show help for <command>`,
	"CommandArgName":              `command`,
	"VersionFlagName":             `version`,
	"VersionFlagShort":            `v`,
	"VersionFlagUsageTemplate":    `show version`,
	"HelpFlagUsageTemplate":       `Show context-sensitive help`,

	// Errors
	"Error":                         "Error: %s",
	"FlagLongExistsTemplate":        `flag --{{.Name}} already exists`,
	"FlagShortExistsTemplate":       `flag -{{.Short}} already exists`,
	"MixArgsCommandsTemplate":       `can't mix Arg()s with Command()s`,
	"UnknownArgument":               `unknown argument {{.Name}}`,
	"UnknownFlagTemplate":           `unknow {{.Prefix}}{{.Name}} flag`,
	"UnexpectedFlagValueTemplate":   `expected argument for flag {{if .Short}}-{{.Short|Rune}}{{else}}--{{.Name}}{{end}} {{if .Value}got '{{.Value}}'{{end}}}}`,
	"UnexpectedTokenTemplate":       `expected {{.Extra}} but got {{.Name}}`,
	"WrongFlagArgumentTypeTemplate": `wrong {{.GetType}} type`,
	"FlagAlreadySet":                `flag {{.GetName}} already have been set. This flag is not cumulative and can only appear once on colland line`,
	"NoHintsForEnumArg":             `no hints speciffied for argument {{.GetName}}`,
	"UnknownArgumentValue":          `unsupported value {{.Extra}} for argument {{.Name}}`,
	"MissingRequired":               `required {{.GetType}} --{{.Name}}{{if .Short}}(-{{.Short|Rune}}){{end}} is missing `,
	"FlagsArgsFromMultipleGroups":   `either {{.Name}} or {{.Extra}} can be specified, but not both`,
	"FlagValidationFailed":          `Invalid flag value {{.Value}} for flag {{if .Short}}-{{.Short|Rune}}{{else}}--{{.Name}}{{end}}: {{.Extra}}`,

	"FormatCommandsCategory":    "Commands",
	"FormatMisCommandsCategory": "Miscellaneous Commands",
	"FormatFlagWithShort":       "-%c, --%s",
	"FormatFlagNoShort":         "--%s",
	"FormatFlagShort":           "-%c",
	"FormatArg":                 "%s",
}
