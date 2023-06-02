package gocli


type String string
type List []string
type Bool bool
type OneOf string
type OneOfList []string

type IFlagArg interface {
	Compare(IFlagArg) int
	GetName() string
	GetUsage() string
	GetDefault() string
	GetHints() []string
	GetPlaceholder() string
	IsRequired() bool
	IsSetByUser() bool
	IsCumulative() bool
	GetValue() interface{}
	SetByUser()
	SetValue(value string) error
	SetRequired(bool)
	SetPlaceholder(string)
	Clear()
	GetValidationGroups() []string
	ValidateWrapper(a *Application) error
	// private methodds
	getDestination() interface{}
}

type ICommand interface {
	FullCommand() string
	ValidateWrapper(*Application) error
	ActionWraper(*Application) error
}

type ValidationGroup struct {
	RequiredFlags []IFlag
	OptionalFlags []IFlag
	RequiredArgs  []IArg
	OptionalArgs  []IArg
}
type GroupedFlagsArgs struct {
	Ungrouped ValidationGroup
	Groups    map[string]ValidationGroup
}
type CommandCategory struct {
	Name     string
	Order    int
	Commands []*Command
}
