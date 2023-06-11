package gocli

type FlagValidator func(a *Application, f IFlag) error

type TFlag interface {
	String | []String | Bool | OneOf | Email | []Email | File | []File
}

type IFlag interface {
	IFlagArg
	IsBool() bool
	SetShort(c rune)
	GetShort() rune
}

type Flag[T TFlag] struct {
	Name             string
	Short            rune
	Usage            string
	Default          string
	Hints            []string
	Destination      *T
	Required         bool
	Placeholder      string
	isSetByUser      bool
	ValidationGroups []string
	Validator        FlagValidator
}

func (f *Flag[T]) Compare(ff IFlagArg) int {
	if f == ff {
		return 0
	}
	return 1
}

func (f *Flag[T]) IsBool() bool {
	_, ok := any(f.Destination).(*Bool)
	return ok
}

func (f *Flag[T]) GetName() string {
	return f.Name
}
func (f *Flag[T]) GetShort() rune {
	return f.Short
}

func (f *Flag[T]) GetUsage() string {
	return f.Usage
}
func (f *Flag[T]) GetDefault() string {
	return f.Default
}
func (f *Flag[T]) GetHints() []string {
	return f.Hints
}
func (f *Flag[T]) GetPlaceholder() string {
	return f.Placeholder
}
func (f *Flag[T]) SetPlaceholder(placeholder string) {
	f.Placeholder = placeholder
}
func (f *Flag[T]) IsRequired() bool {
	return f.Required
}
func (f *Flag[T]) SetName(name string) {
	f.Name = name
}
func (f *Flag[T]) SetShort(c rune) {
	f.Short = c
}

func (f *Flag[T]) SetRequired(is_required bool) {
	f.Required = is_required
}

func (f *Flag[T]) IsSetByUser() bool {
	return f.isSetByUser
}

func (f *Flag[T]) IsCumulative() bool {
	return IsCumulative(f)
}

func (f *Flag[T]) SetByUser() {
	f.isSetByUser = true
}

func (f *Flag[T]) getDestination() interface{} {
	if f.Destination == nil {
		f.Destination = new(T)
	}
	return f.Destination
}

func (f *Flag[T]) SetValue(value string) error {
	err := setFlagArgValue(f, value)
	if err != nil {
		return err
	}
	f.SetByUser()
	return nil

}

func (f *Flag[T]) Clear() {
	f.isSetByUser = false
	f.Destination = new(T)
}

func (f *Flag[T]) ValidateWrapper(a *Application) error {
	if f.Validator != nil {
		return f.Validator(a, f)
	}
	return nil
}

func (f *Flag[T]) GetValue() interface{} {
	return getFlagArgValue(f)
}

func (f *Flag[T]) GetValidationGroups() []string {
	return f.ValidationGroups
}