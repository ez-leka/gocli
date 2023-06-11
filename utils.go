package gocli

import (
	"reflect"
	"strings"

	"github.com/ez-leka/gocli/i18n"
)

func MapIFlag(m map[string]IFlag) []IFlagArg {
	uchecker := make(map[IFlag]bool)
	ret := make([]IFlagArg, 0)
	for _, f := range m {
		if !uchecker[f] {
			ret = append(ret, f)
			uchecker[f] = true
		}
	}
	return ret
}
func MapIArg(m []IArg) []IFlagArg {
	ret := make([]IFlagArg, 0)
	for _, f := range m {
		ret = append(ret, f)
	}
	return ret
}

// set value that works for flags and arguments
func inHints(hints []string, value string) string {

	for _, h := range hints {
		if strings.HasSuffix(h, "(s)") {
			single := h[0 : len(h)-3]
			plural := single + "s"
			if value == single || value == plural {
				return single
			}
		} else {
			if value == h {
				return value
			}
		}
	}
	return ""
}

func getValueFromHints(fa IFlagArg, value string, is_flag bool) (string, error) {
	hints := fa.GetHints()

	if len(hints) == 0 {
		if is_flag {
			return "", i18n.NewError("NoHintsForEnumArg", FlagTemplateContext{Name: fa.GetName()})
		} else {
			return "", i18n.NewError("NoHintsForEnumArg", ArgTemplateContext{Name: fa.GetName()})
		}
	}

	true_value := inHints(hints, value)
	if true_value == "" {
		if is_flag {
			return "", i18n.NewError("UnknownFlagValue", FlagTemplateContext{Name: fa.GetName(), Extra: value})
		} else {
			return "", i18n.NewError("UnknownArgumentValue", ArgTemplateContext{Name: fa.GetName(), Extra: value})
		}
	}

	return true_value, nil
}

func setFlagArgValue(fa IFlagArg, value string) error {
	dest := fa.getDestination()
	rv := reflect.ValueOf(dest).Elem()

	is_flag := false
	if _, ok := fa.(IFlag); ok {
		is_flag = true
	}

	if is_flag && !fa.IsCumulative() && fa.IsSetByUser() {
		return i18n.NewError("FlagAlreadySet")
	}

	if fa.IsCumulative() {
		t := reflect.TypeOf(dest).Elem().Elem()
		new_value := reflect.New(t).Interface()
		err := new_value.(Setable).FromString(value, fa)
		if err != nil {
			return err
		}
		rv.Set(reflect.Append(rv, reflect.ValueOf(new_value).Elem()))
	} else {
		setable := dest.(Setable)
		err := setable.FromString(value, fa)
		if err != nil {
			return err
		}
	}

	return nil
}

func getFlagArgValue(fa IFlagArg) interface{} {

	dest := fa.getDestination()

	rt := reflect.TypeOf(dest).Elem()
	rv := reflect.ValueOf(dest).Elem()
	switch rt.Kind() {
	case reflect.Slice:
		val := []string{}
		for i := 0; i < rv.Len(); i++ {
			val = append(val, rv.Index(i).String())
		}
		return val
	case reflect.Bool:
		return bool(rv.Bool())
	case reflect.String:
		return rv.String()
	}

	return nil
}

func IsType[T TFlag | TArg](fa IFlagArg) bool {

	_, ok := any(fa.getDestination()).(*T)
	return ok
}

func IsCumulative(fa IFlagArg) bool {
	rt := reflect.TypeOf(fa.getDestination()).Elem()
	switch rt.Kind() {
	case reflect.Slice:
		return true
	default:
		return false
	}
}
