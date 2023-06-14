package gocli

import "github.com/ez-leka/gocli/i18n"

var GoCliStrings = i18n.Entries{
	"AppUsageTemplate": `
{{- define "CmdFlag"}}
{{- if .GetShort}} -{{.GetShort|Rune}}{{else}} --{{.GetName}}{{end -}}
{{- if not .IsBool}}[=]<{{.GetPlaceholder}}>{{end -}}
{{- end -}}

{{- define "CmdArg"}} <{{.GetPlaceholder}}>{{- end -}}

{{- define "CmdGroup"}}
{{- if .Command}}{{Translate .Command}}{{end -}}
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
{{- $groups := .CurrentCommand.GetGroupedFlagsAndArgs -}}
{{- $group_idx := 0}}
{{Indent 4}}{{.CurrentCommand.FullCommand}} 
{{- template "CmdGroup" .CurrentCommand.GetGlobalFlags}} {{template "CmdGroup" $groups.Ungrouped }} 
{{- if $groups.Groups}} ({{end -}}
  {{- range $groups.Groups}}
  {{- if eq $group_idx 1}} | {{end}} 
  {{- template "CmdGroup" .}}{{- $group_idx = 1}}
  {{- end -}}
  {{- if $groups.Groups}} ){{end}}
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
}
