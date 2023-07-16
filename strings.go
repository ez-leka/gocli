package gocli

import "github.com/ez-leka/gocli/i18n"

var GoCliStrings = i18n.Entries{
	"BashCompletionTemplate": `
	#/usr/bin/env bash
	_{{.Name}}_bash_autocomplete() {
		local cur prev opts base
		COMPREPLY=()
		cur="{{printf "\x24"}}{{printf "\x7B"}}COMP_WORDS[COMP_CWORD]{{printf "\x7D"}}"
		opts=$( {{printf "\x24"}}{{printf "\x7B"}}COMP_WORDS[0]{{printf "\x7D"}} --bash-completions "{{printf "\x24"}}{{printf "\x7B"}}COMP_WORDS[@]:1:$COMP_CWORD{{printf "\x7D"}}" )
		COMPREPLY=( $(compgen -W "{{printf "\x24"}}{{printf "\x7B"}}opts{{printf "\x7D"}}" -- {{printf "\x24"}}{{printf "\x7B"}}cur{{printf "\x7D"}}) )
		return 0
	}
	complete -F _{{.Name}}_bash_autocomplete -o default {{.Path}}
	
	`,

	"ZshCompletionTemplate": `#compdef {{.Name}}
	
	_{{.Name}}() {{printf "\x7B"}}
		local matches=($({{printf "\x24"}}{{printf "\x7B"}}words[1]{{printf "\x7D"}} --completion-bash "{{printf "\x24"}}{{printf "\x7B"}}(@)words[2,$CURRENT]{{printf "\x7D"}}"))
		compadd -a matches
	
		if [[ $compstate[nmatches] -eq 0 && $words[$CURRENT] != -* ]]; then
			_files
		fi
	{{printf "\x7D"}}
	
	if [[ "{{printf "\x24"}}{{printf "\x28"}}basename -- {{printf "\x24"}}{{printf "\x7B"}}(%%):-%%x{{printf "\x7D"}})" != "_{{.Name}}" ]]; then
		compdef _{{.Name}} {{.Name}}
	fi`,
	"CmdFlagTemplate": `
{{- define "CmdFlag"}}
{{- if .GetShort}} -{{.GetShort|Rune}}{{else}} --{{.GetName}}{{- end}}
{{- if not .IsBool}}[=]<{{.GetPlaceholder}}>{{- end}}{{if .IsCumulative}}...{{end}}
{{- end -}}`,
	"CmdArgTemplate": `
{{define "CmdArg"}}<{{.GetPlaceholder}}>{{end}}
`,
	"CmdGroupTemplate": `
{{- define "CmdGroup"}}
{{- if .Command}} {{if .IsGenericCommand}}<{{end}}{{Translate .Command}} {{if .IsGenericCommand}}>{{end}}{{- end}}
{{- range .RequiredFlags}}{{template "CmdFlag" .}}{{- end}}
{{- if .OptionalFlags}} [{{end -}}{{- range .OptionalFlags}}{{template "CmdFlag" .}}{{end -}}{{- if .OptionalFlags}} ]{{- end}}
{{- range .RequiredArgs}} {{template "CmdArg" .}}{{- end}}
{{- if .OptionalArgs}} [{{end -}}{{- range .OptionalArgs}}{{template "CmdArg" .}}{{- end}}{{- if .OptionalArgs}} ]{{end}}
{{- end -}}`,
	"FormatCommandCategoryTemplate": `
{{- define "FormatCommandCategory"}}
{{- if .}}
{{- range .|CommandCategories}}
{{.Name}}
{{.GetCommands|CommandsToTwoColumns|TwoColumns}}
{{- end}}
{{- end}}
{{- end -}}`,
	"FlagListTemplate": `
{{- define "FlagList"}}
{{- if .}}
{{Translate "Flags"}}:
{{.|FlagsArgsToTwoColumns|TwoColumns}}
{{- end}}
{{- end -}}`,
	"ArgListTemplate": `
{{- define "ArgList"}}
{{- if .}}
{{Translate "Arguments"}}:
{{.|FlagsArgsToTwoColumns|TwoColumns}}
{{- end -}}	
{{- end -}}
`,
	"AppUsageTemplate": `
{{- if .CurrentCommand.Description}}
{{FormatTemplate .CurrentCommand.Description .CurrentCommand 4}}
{{end}}
{{- template "FormatCommandCategory" .CurrentCommand.Commands}}
{{- template "FlagList" .Flags}}
{{- template "ArgList" .Args}}
{{Translate "Usage"}}: 
{{- $groups := .CurrentCommand.GetGroupedFlagsAndArgs -}}
{{- $group_idx := 0}}
{{Indent 4}}{{.CurrentCommand.FullCommand}} 
{{- if $groups.Ungrouped -}}
	{{- template "CmdGroup" $groups.Ungrouped -}}
{{- end -}}
{{- if gt (len $groups.Groups) 1}} ({{end -}}
  {{- range $groups.Groups}}
  {{- if eq $group_idx 1}} | {{end}} 
  {{- template "CmdGroup" .}}{{- $group_idx = 1}}
  {{- end -}}
  {{- if gt (len $groups.Groups) 1}} ){{end}}

{{if .CurrentCommand.Usage}}
{{FormatTemplate .CurrentCommand.Usage .CurrentCommand 4}}
{{end}}
`,
	"BashCompletionFlagName":          `bash-completions`,
	"BashCompletionFlagUsageTemplate": `used in dynamic bash completion`,
	"HelpCommandAndFlagName":          `help`,
	"HelpFlagShort":                   `h`,
	"HelpCommandUsageTemplate":        `show help`,
	"HelpCommandArgUsageTemplate":     `show help for <command>`,
	"CommandArgName":                  `command`,
	"VersionFlagName":                 `version`,
	"VersionFlagShort":                `v`,
	"VersionFlagUsageTemplate":        `show version`,
	"HelpFlagUsageTemplate":           `Show context-sensitive help`,

	// Errors
	"Error":                         "Error: %s",
	"FlagLongExistsTemplate":        `flag --{{.Name}} already exists`,
	"FlagShortExistsTemplate":       `flag -{{.Short|Rune}} already exists`,
	"UnknownElementTemplate":        `unknown {{.Element.GetType}} {{.Element.GetPlaceholder}}`,
	"ExtraArgument":                 `unexpected argument {{.Extra}}`,
	"UnexpectedFlagValueTemplate":   `expected argument for flag --{{.Element.Name}} {{if .Element.Short}}(-{{.Element.Short|Rune}}{{end}}) {{if .Extra}got '{{.Extra}}'{{end}}}}`,
	"UnexpectedTokenTemplate":       `expected {{.Extra}} but got {{.Name}}`,
	"WrongElementTypeTemplate":      `wrong {{.Element.GetType}} type for {{.Element.Name}}`,
	"FlagAlreadySet":                `flag {{.GetName}} already have been set. This flag is not cumulative and can only appear once on command line`,
	"NoHintsForOneOf":               `no hints speciffied for {{.GetType}} {{.GetName}}`,
	"UnknownOneOfValue":             `unsupported value {{.Extra}} for {{.Element.GetType}} {{.Element.GetPlaceholder}}`,
	"InvalidTimeFormat":             `invalid timestamp string {{.Extra}} for {{.Element.GetType}} {{.Element.GetPlaceholder}}`,
	"InvalidIPFormat":               `invalid IP string {{.Extra}} for {{.Element.GetType}} {{.Element.GetPlaceholder}}`,
	"InvalidIntFormat":              `invalid int string {{.Extra}} for {{.Element.GetType}} {{.Element.GetPlaceholder}}`,
	"InvalidHexFormat":              `invalid hex string {{.Extra}} for {{.Element.GetType}} {{.Element.GetPlaceholder}}`,
	"InvalidBinaryFormat":           `invalid binary string {{.Extra}} for {{.Element.GetType}} {{.Element.GetPlaceholder}}`,
	"InvalidOctalFormat":            `invalid octal string {{.Extra}} for {{.Element.GetType}} {{.Element.GetPlaceholder}}`,
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
	"FormatDefault":                 "(Default: %s)",
	"FormatHints":                   "One of: %s",
}
