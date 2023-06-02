package gocli

import (
	"reflect"
	"strconv"
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
			return "", i18n.NewError("UnknownArgumentValue", FlagTemplateContext{Name: fa.GetName(), Extra: value})
		} else {
			return "", i18n.NewError("UnknownArgumentValue", ArgTemplateContext{Name: fa.GetName(), Extra: value})
		}
	}

	return true_value, nil
}

func setFlagArgValue(fa IFlagArg, value string) error {
	var err error
	dest := fa.getDestination()
	rv := reflect.ValueOf(dest).Elem()

	is_flag := false
	if _, ok := fa.(IFlag); ok {
		is_flag = true
	}

	switch dest.(type) {
	case *Bool:
		if is_flag && fa.IsSetByUser() {
			return i18n.NewError("FlagAlreadySet")
		}

		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(Bool(v)))
	case *OneOf:
		// validate value against hints
		value, err = getValueFromHints(fa, value, is_flag)
		if err != nil {
			return err
		}
		if is_flag && fa.IsSetByUser() {
			return i18n.NewError("FlagAlreadySet")
		}
		rv.Set(reflect.ValueOf(OneOf(value)))

	case *String:
		if is_flag && fa.IsSetByUser() {
			return i18n.NewError("FlagAlreadySet")
		}
		rv.Set(reflect.ValueOf(String(value)))

	case *List:
		rv.Set(reflect.Append(rv, reflect.ValueOf(value)))
	case *OneOfList:
		true_value, err := getValueFromHints(fa, value, is_flag)
		if err != nil {
			return err
		}
		rv.Set(reflect.Append(rv, reflect.ValueOf(true_value)))
	}
	return nil
}

func getFlagArgValue(fa IFlagArg) interface{} {

	dest := fa.getDestination()

	switch v := dest.(type) {
	case *String:
		return string(*(v))
	case *Bool:
		return bool(*(v))
	case *List:
		cumulative := *(v)
		val := []string{}
		for _, v := range cumulative {
			val = append(val, v)
		}
		return val

	case *OneOfList:
		cumulative := *(v)
		val := []string{}
		for _, v := range cumulative {
			val = append(val, v)
		}
		return val
	case *OneOf:
		return string(*(dest.(*OneOf)))

	}
	return nil
}

func IsType[T TFlag | TArg](fa IFlagArg) bool {

	_, ok := any(fa.getDestination()).(*T)
	return ok
}
