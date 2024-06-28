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
{{- if .GetShort}} -{{.GetShort|Rune}}{{else}} --{{.GetName}}{{end -}}
{{- if not .IsBool}}[=]<{{.GetPlaceholder}}>{{end}}{{if .IsCumulative}}...{{end -}}
{{end -}}`,
	"CmdArgTemplate": `
{{- define "CmdArg"}}<{{.GetPlaceholder}}>{{end -}}
`,
	"CmdFullCommand": `

	`,
	"CmdGroupTemplate": `
{{- define "CmdGroup"}}
{{- if .Group.Command}} {{if .Group.IsGenericCommand}}<{{end}}{{Translate .Group.Command}} {{if .Group.IsGenericCommand}}>{{end}}{{end -}}
{{- if .Group.HasGlobalFlags .Level}} [global options] {{end -}}
{{- range .Group.RequiredFlags .Level}}{{template "CmdFlag" .}}{{end -}}
{{- if .Group.OptionalFlags .Level}} [{{end}}{{range .Group.OptionalFlags .Level}}{{template "CmdFlag" .}}{{end}}{{if .Group.OptionalFlags .Level}} ]{{end -}}
{{- range .Group.RequiredArgs}} {{template "CmdArg" .}}{{end -}}
{{- if .Group.OptionalArgs}} [ {{end}}{{range .Group.OptionalArgs}}{{template "CmdArg" .}} {{end}}{{if .Group.OptionalArgs}}]{{end -}}
{{end -}}`,
	"FormatCommandCategoryTemplate": `
{{- define "FormatCommandCategory"}}
{{- if .}}
{{- range .|CommandCategories}}
{{HLevel 1}} {{.Name}}
{{.GetCommands|CommandsToTwoColumns|DefinitionList}}
{{end -}}
{{end -}}
{{end -}}`,
	"FlagListTemplate": `
{{define "FlagList"}}
{{if .Flags}}
{{HLevel 1}} {{if eq .Level 0}}{{Translate "Global"}} {{end}}{{Translate "Options"}}
{{FlagsArgsToTwoColumns .Flags .Level|DefinitionList}}
{{end}}
{{end}}`,
	"ArgListTemplate": `
{{define "ArgList"}}
{{if .Args}}
{{HLevel 1}} {{Translate "Arguments"}}:
{{FlagsArgsToTwoColumns .Args .Level|DefinitionList}}
{{end}}
{{end}}
`,
	"AppUsageTemplate": `
{{- if eq .Level  0}}
{{- HLevel 1}} {{Translate "Name"}}
{{.CurrentCommand.FullCommand -}}
{{- else -}}
{{- HLevel 0}} {{.CurrentCommand.Name}} {{- if .DocGeneration}} {{Translate "command"}}{{ else }} - {{FormatTemplate .CurrentCommand.Description .CurrentCommand}}{{end -}}
{{- end}}
{{if eq .Level  0 -}}
{{HLevel 1}} {{Translate "Synopsis"}}
{{- end}}
{{BlockBracket}}
{{- $groups := .CurrentCommand.GetGroupedFlagsAndArgs}}
{{- $group_idx := 0}}
{{.CurrentCommand.FullCommand}}
{{- if $groups.Ungrouped}}
	{{- template "CmdGroup" Dict "Group" $groups.Ungrouped "Level" .Level}}
{{end -}}
{{- if gt (len $groups.Groups) 1}} ({{end -}}
  {{- range $groups.Groups -}}
  {{- if eq $group_idx 1}} | {{end -}}
  {{- template "CmdGroup" Dict "Group" . "Level" .Level -}}{{$group_idx = 1}}
  {{end -}}
  {{- if gt (len $groups.Groups) 1}} ){{end -}}
{{BlockBracket}}
{{if and .CurrentCommand.Description .DocGeneration}}
{{- if eq .Level  0}}
{{HLevel 1}} {{Translate "Description"}}
{{end -}}
{{FormatTemplate .CurrentCommand.Description .CurrentCommand}}
{{end -}}
{{- if .CurrentCommand.Usage}}
{{FormatTemplate .CurrentCommand.Usage .CurrentCommand}}
{{end -}}
{{- template "FormatCommandCategory" .CurrentCommand.Commands}}
{{- template "FlagList" Dict "Flags" .Flags "Level" .Level}}
{{- template "ArgList" Dict "Args" .Args  "Level" .Level}}
{{if not .DocGeneration}}
Use "{{.AppName}} <command> --help" for more information about a given command.
{{if .UseOptionsCommand}}
Use "{{.AppName}} options" for a list of global command-line options (applies to all commands).
{{end}}
{{end}}
`,
	"ShellCompletionCommand":           `generate-completion`,
	"ShellCompletionCommandDesc":       `generate completion script for bash or zch shell`,
	"ShellCompletionFlagUsageTemplate": `used in dynamic bash completion`,
	"ShellCompletionArgName":           `shell`,
	"ShellCompetionArgUsage":           `type of shell for which to generate complition script`,
	"DocGenerationCommand":             `generate-documentation`,
	"DocGenerationCommandDesc":         `Generate documentation in specified format`,
	"DocGenerationFormatArgName":       `format`,
	"DocGenerationFormatArgUsage":      `Format of documenation to be generated.`,
	"DocGenerationCssFlagName":         `css`,
	"DocGenerationCssFlagUsage":        `path to CSS stylesheet (applies to HTML only).`,
	"DocGenerationIconFlagName":        `icon`,
	"DocGenerationIconFlagUsage":       `path to image to be sed as browser icon (applies to HTML only).`,
	"DocGenerationTocFlagName":         `toc`,
	"DocGenerationTocFlagUsage":        `if set, TOC will be generated (applies to HTML only)`,

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
	"command":                       `command`,
	"subCommand":                    `sub-command`,
	"FormatCommandsCategory":        "Commands",
	"FormatMisCommandsCategory":     "Miscellaneous Commands",
	"FormatFlagWithShort":           "-%c, --%s",
	"FormatFlagNoShort":             "--%s",
	"FormatFlagShort":               "-%c",
	"FormatArg":                     "%s",
	"FormatDefault":                 "(Default: %s)",
	"FormatHints":                   "One of %s",
	"FormatGlobal":                  "Global",
}
