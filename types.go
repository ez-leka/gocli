package gocli

import (
	"net"
	"net/mail"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
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
	IsHidden() bool
}

type String string
type Bool bool
type Int int
type Hex int
type Binary int
type Octal int
type OneOf string
type Email string
type TimeStamp time.Time
type Duration time.Duration
type IP net.IP
type File String // Represents existing file path. Will not validate if file does not exist. Use String type if you do not want validate for existance

type TArgFlag interface {
	String | []String | OneOf | Email | []Email | File | []File | TimeStamp | []TimeStamp | Duration | []Duration | Int | []Int | Hex | []Hex | Octal | []Octal | Binary | []Binary | IP | []IP
}

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

func (s *Duration) GetReturnType() reflect.Type {
	var d time.Duration
	return reflect.TypeOf(d)
}

func (s *Duration) FromString(v string, fa IFlagArg) error {
	d, err := time.ParseDuration(v)
	if err != nil {
		return err
	}
	*s = Duration(d)
	return nil
}

func (s *Duration) GetValue() interface{} {
	return time.Duration(*s)
}

func (s *IP) GetReturnType() reflect.Type {
	return reflect.TypeOf(net.IP{})
}

func (s *IP) FromString(v string, fa IFlagArg) error {

	ip := net.ParseIP(v)
	if ip == nil {
		return i18n.NewError("InvalidIPFormat", ElementTemplateContext{Element: fa, Extra: v})
	}
	*s = IP(ip)
	return nil
}

func (s *IP) GetValue() interface{} {
	return net.IP(*s)
}

func (s *Int) GetReturnType() reflect.Type {
	var i int
	return reflect.TypeOf(i)
}

func (s *Int) FromString(v string, fa IFlagArg) error {

	i, err := strconv.ParseInt(v, 10, 0)
	if err != nil {
		return i18n.NewError("InvalidIntFormat", ElementTemplateContext{Element: fa, Extra: v})
	}
	*s = Int(i)
	return nil
}

func (s *Int) GetValue() interface{} {
	return int(*s)
}

func (s *Hex) GetReturnType() reflect.Type {
	var i int
	return reflect.TypeOf(i)
}

func (s *Hex) FromString(v string, fa IFlagArg) error {

	// if string starts with 0x - remove it before parsing
	v = strings.Replace(v, "0x", "", -1)
	i, err := strconv.ParseInt(v, 16, 0)
	if err != nil {
		return i18n.NewError("InvalidHexFormat", ElementTemplateContext{Element: fa, Extra: v})
	}
	*s = Hex(i)
	return nil
}

func (s *Hex) GetValue() interface{} {
	return int(*s)
}

func (s *Binary) GetReturnType() reflect.Type {
	var i int
	return reflect.TypeOf(i)
}

func (s *Binary) FromString(v string, fa IFlagArg) error {

	i, err := strconv.ParseInt(v, 2, 0)
	if err != nil {
		return i18n.NewError("InvalidBinaryFormat", ElementTemplateContext{Element: fa, Extra: v})
	}
	*s = Binary(i)
	return nil
}

func (s *Binary) GetValue() interface{} {
	return int(*s)
}

func (s *Octal) GetReturnType() reflect.Type {
	var i int
	return reflect.TypeOf(i)
}

func (s *Octal) FromString(v string, fa IFlagArg) error {

	i, err := strconv.ParseInt(v, 8, 0)
	if err != nil {
		return i18n.NewError("InvalidOctalFormat", ElementTemplateContext{Element: fa, Extra: v})
	}
	*s = Octal(i)
	return nil
}

func (s *Octal) GetValue() interface{} {
	return int(*s)
}

type IFlagArg interface {
	IValidatable
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
	ActionWrapper(*Application, interface{}) (interface{}, error)
}
type CommandCategory struct {
	Name     string
	Order    int
	commands []*Command
}

func (cat *CommandCategory) GetCommands() []*Command {
	return cat.commands
}

// private types
type validationGroup struct {
	Command          string
	IsGenericCommand bool
	RequiredFlags    []IValidatable
	OptionalFlags    []IValidatable
	RequiredArgs     []IValidatable
	OptionalArgs     []IValidatable
}
type groupedFlagsArgs struct {
	Ungrouped validationGroup
	Groups    map[string]validationGroup
}
