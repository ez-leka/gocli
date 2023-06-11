package gocli

import "strings"

type IArg interface {
	IFlagArg
}

type TArg interface {
	String | []String | Bool | OneOf | []OneOf | Email | []Email | File | []File
}

type ArgValidator func(a *Application, arg IArg) error

type Arg[T TArg] struct {
	Name             string
	Usage            string
	Hints            []string
	Default          string
	Placeholder      string
	Required         bool
	Destination      *T
	isSetByUser      bool
	ValidationGroups []string
	Validator        ArgValidator
}

func (a *Arg[T]) Compare(aa IFlagArg) int {
	if a == aa {
		return 0
	}
	return 1
}

func (a *Arg[T]) GetName() string {
	return a.Name
}
func (a *Arg[T]) GetShort() rune {
	return 0
}

func (a *Arg[T]) GetUsage() string {
	return a.Usage
}
func (a *Arg[T]) GetHints() []string {
	return a.Hints
}
func (a *Arg[T]) IsRequired() bool {
	return a.Required
}

func (a *Arg[T]) IsSetByUser() bool {
	return a.isSetByUser
}

func (a *Arg[T]) IsCumulative() bool {
	return IsCumulative(a)
}

func (a *Arg[T]) GetDefault() string {
	return a.Default
}

func (a *Arg[T]) getDestination() interface{} {
	if a.Destination == nil {
		a.Destination = new(T)
	}
	return a.Destination
}

func (a *Arg[T]) GetPlaceholder() string {
	return a.Placeholder
}
func (a *Arg[T]) SetPlaceholder(placeholder string) {
	a.Placeholder = placeholder
}

func (a *Arg[T]) SetRequired(is_required bool) {
	a.Required = is_required
}

func (a *Arg[T]) SetValue(value string) error {

	var vals []string
	if a.IsCumulative() {
		// could be comma-separated value
		vals = strings.Split(value, ",")
	} else {
		vals = []string{value}
	}

	for _, v := range vals {
		err := setFlagArgValue(a, v)
		if err != nil {
			return err
		}
	}
	a.SetByUser()
	return nil
}

func (a *Arg[T]) Clear() {
	a.isSetByUser = false
	a.Destination = new(T)
}

func (a *Arg[T]) SetByUser() {
	a.isSetByUser = true
}

func (a *Arg[T]) GetValue() interface{} {
	return getFlagArgValue(a)
}

func (a *Arg[T]) ValidateWrapper(app *Application) error {
	if a.Validator != nil {
		return a.Validator(app, a)
	}
	return nil
}

func (a *Arg[T]) GetValidationGroups() []string {
	return a.ValidationGroups
}
