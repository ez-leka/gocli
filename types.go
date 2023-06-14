package gocli

import (
	"net/mail"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	"github.com/ez-leka/gocli/i18n"
)

type ISetable interface {
	FromString(string, IFlagArg) error
	GetValue() interface{}
	GetReturnType() reflect.Type
}

type IValidatable interface {
	GetName() string
	GetType() string
	GetValidationGroups() []string
	IsSetByUser() bool
	IsRequired() bool
	ValidateWrapper(a *Application) error
	GetPlaceholder() string
}

type String string
type Bool bool
type OneOf string
type Email string
type TimeStamp time.Time
type File String // Represents existing file path. Will not validate if file does not exist. Use String type if you do not want validate for existance

func (s *String) GetReturnType() reflect.Type {
	return reflect.TypeOf("")
}

func (s *String) FromString(v string, fa IFlagArg) error {
	*s = String(v)
	return nil
}
func (s *String) GetValue() interface{} {
	return string(*s)
}

func (s *Bool) GetReturnType() reflect.Type {
	return reflect.TypeOf(true)
}

func (s *Bool) FromString(v string, fa IFlagArg) error {
	b, err := strconv.ParseBool(v)
	if err != nil {
		return err
	}
	*s = Bool(b)
	return nil
}
func (s *Bool) GetValue() interface{} {
	return bool(*s)
}

func (s *OneOf) GetReturnType() reflect.Type {
	return reflect.TypeOf("")
}

func (s *OneOf) FromString(v string, fa IFlagArg) error {

	is_flag := false
	if _, ok := fa.(IFlag); ok {
		is_flag = true
	}

	hints := fa.GetHints()
	if len(hints) == 0 {
		if is_flag {
			return i18n.NewError("NoHintsForOneOf", fa)
		} else {
			return i18n.NewError("NoHintsForOneOf", fa)
		}
	}

	true_value := inHints(hints, v)
	if true_value == "" {
		return i18n.NewError("UnknownOneOfValue", ElementTemplateContext{Element: fa, Extra: v})
	}

	*s = OneOf(true_value)
	return nil
}

func (s *OneOf) GetValue() interface{} {
	return string(*s)
}

func (s *Email) GetReturnType() reflect.Type {
	return reflect.TypeOf("")
}

func (s *Email) FromString(v string, fa IFlagArg) error {
	_, err := mail.ParseAddress(v)
	if err != nil {
		return err
	}

	*s = Email(v)
	return nil
}
func (s *Email) GetValue() interface{} {
	return string(*s)
}

func (s *File) GetReturnType() reflect.Type {
	return reflect.TypeOf("")
}

func (s *File) FromString(v string, fa IFlagArg) error {
	matches, err := filepath.Glob(v)

	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return os.ErrNotExist
	}
	*s = File(v)
	return nil
}
func (s *File) GetValue() interface{} {
	return string(*s)
}

func (s *TimeStamp) GetReturnType() reflect.Type {
	return reflect.TypeOf(time.Time{})
}

func (s *TimeStamp) FromString(v string, fa IFlagArg) error {
	// try to parse it any known format
	layouts := []string{
		time.RFC822,      //	= "02 Jan 06 15:04 MST"
		time.RFC822Z,     //    = "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
		time.RFC850,      //    = "Monday, 02-Jan-06 15:04:05 MST"
		time.RFC1123,     //	= "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC1123Z,    //	= "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
		time.RFC3339,     //  	= "2006-01-02T15:04:05Z07:00"
		time.RFC3339Nano, // 	= "2006-01-02T15:04:05.999999999Z07:00",
		"01/01/2006 03:04:05 PM",
		"01/01/2006 03:04:05PM",
		"03:04 PM",
		"03:04PM",
	}
	parsed := false
	for _, l := range layouts {
		t, err := time.Parse(l, v)
		if err == nil {
			*s = TimeStamp(t)
			parsed = true
			break
		}
	}
	if !parsed {
		return i18n.NewError("InvalidTimeFormat", ElementTemplateContext{Element: fa, Extra: v})
	}
	return nil
}
func (s *TimeStamp) GetValue() interface{} {
	return time.Time(*s)
}

type IFlagArg interface {
	IValidatable
	Compare(IFlagArg) int
	GetUsage() string
	GetDefault() string
	GetHints() []string
	IsCumulative() bool
	GetValue() interface{}
	SetByUser()
	SetValue(value string) error
	SetRequired(bool)
	SetPlaceholder(string)
	Clear()
	// private methodds
	getDestination() interface{}
}

type ICommand interface {
	IValidatable
	FullCommand() string
	ActionWraper(*Application) error
}

type ValidationGroup struct {
	Command       string
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
