package gocli

import (
	"net/mail"
	"os"
	"strconv"

	"github.com/ez-leka/gocli/i18n"
)

type Setable interface {
	FromString(string, IFlagArg) error
}
type String string
type Bool bool
type OneOf string
type Email String
type EmailList []string

// Represents existing file. Will not validate if file does not exist. Use String type if you do not want validate for existance
type File String

// List of existing files. Will not validate if file does not exist
type FileList []File

func (s *String) FromString(v string, fa IFlagArg) error {
	*s = String(v)
	return nil
}

func (s *Bool) FromString(v string, fa IFlagArg) error {
	b, err := strconv.ParseBool(v)
	if err != nil {
		return err
	}
	*s = Bool(b)
	return nil
}

func (s *OneOf) FromString(v string, fa IFlagArg) error {

	is_flag := false
	if _, ok := fa.(IFlag); ok {
		is_flag = true
	}

	hints := fa.GetHints()
	if len(hints) == 0 {
		if is_flag {
			return i18n.NewError("NoHintsForEnumArg", FlagTemplateContext{Name: fa.GetName()})
		} else {
			return i18n.NewError("NoHintsForEnumArg", ArgTemplateContext{Name: fa.GetName()})
		}
	}

	true_value := inHints(hints, v)
	if true_value == "" {
		if is_flag {
			return i18n.NewError("UnknownFlagValue", FlagTemplateContext{Name: fa.GetName(), Extra: v})
		} else {
			return i18n.NewError("UnknownArgumentValue", ArgTemplateContext{Name: fa.GetName(), Extra: v})
		}
	}

	*s = OneOf(true_value)
	return nil
}

func (s *Email) FromString(v string, fa IFlagArg) error {
	_, err := mail.ParseAddress(v)
	if err != nil {
		return err
	}

	*s = Email(v)
	return nil
}

func (s *File) FromString(v string, fa IFlagArg) error {
	_, err := os.Stat(v)
	if err != nil {
		return err
	}

	*s = File(v)
	return nil
}

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
