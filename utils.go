package gocli

import (
	"reflect"
	"strings"

	"github.com/ez-leka/gocli/i18n"
)

func mapIFlag(m map[string]IFlag) []IFlagArg {
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
func mapIArg(m []IArg) []IFlagArg {
	ret := make([]IFlagArg, 0)
	for _, f := range m {
		ret = append(ret, f)
	}
	return ret
}

func mergeValidatables(g validationGroup) []IValidatable {

	return append(append(append(g.RequiredFlags, g.OptionalFlags...), g.RequiredArgs...), g.OptionalArgs...)
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

func setFlagArgValue(fa IFlagArg, value string) error {
	dest := fa.getDestination()
	rv := reflect.ValueOf(dest).Elem()

	is_flag := false
	if _, ok := fa.(IFlag); ok {
		is_flag = true
	}

	cumulative := fa.IsCumulative()
	if is_flag && !cumulative && fa.IsSetByUser() {
		return i18n.NewError("FlagAlreadySet", fa)
	}

	if cumulative {
		t := reflect.TypeOf(dest).Elem().Elem()
		new_value := reflect.New(t).Interface()
		err := new_value.(ISetable).FromString(value, fa)
		if err != nil {
			return err
		}
		rv.Set(reflect.Append(rv, reflect.ValueOf(new_value).Elem()))
	} else {
		setable := dest.(ISetable)
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

	if fa.IsCumulative() {
		tp := reflect.New(rt.Elem()).Interface().(ISetable).GetReturnType()
		elemSlice := reflect.MakeSlice(reflect.SliceOf(tp), 0, 0)

		for i := 0; i < rv.Len(); i++ {
			item := rv.Index(i).Addr().Interface()
			elemSlice = reflect.Append(elemSlice, reflect.ValueOf(item.(ISetable).GetValue()))
		}
		return elemSlice.Interface()
	} else {
		return dest.(ISetable).GetValue()
	}
}

func isType[T TFlag | TArg](fa IFlagArg) bool {

	_, ok := any(fa.getDestination()).(*T)
	return ok
}

func isCumulative(fa IFlagArg) bool {
	rt := reflect.TypeOf(fa.getDestination()).Elem()
	switch rt.Kind() {
	case reflect.Slice:
		if _, ok := reflect.New(rt.Elem()).Interface().(ISetable); ok {
			return true
		} else {
			// thisis array of some primitive type - rnamed type
			return false
		}
	default:
		return false
	}
}
