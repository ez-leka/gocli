package gocli

import (
	"reflect"
	"strings"

	"github.com/ez-leka/gocli/i18n"
)

func lookupFlagsForUsage(m map[string]IFlag, show_up_to_level int, show_hidden_flags bool) []IFlagArg {
	// use every flag once
	uchecker := make(map[IFlag]bool)
	ret := make([]IFlagArg, 0)
	for _, f := range m {
		if f.IsInternal() {
			// never use internal flags in usage
			continue
		}
		if f.GetLevel() < show_up_to_level {
			// do not show commands below show level
			continue
		}
		if f.IsHidden() && !show_hidden_flags {
			continue
		}
		if !uchecker[f] {
			ret = append(ret, f)
			uchecker[f] = true
		}
	}

	return ret
}
func lookupArgsForUsage(m []IArg) []IFlagArg {
	ret := make([]IFlagArg, 0)
	for _, f := range m {
		ret = append(ret, f)
	}
	return ret
}

func commandFlagsForUsage(m []IFlag) []IFlagArg {
	ret := make([]IFlagArg, 0)
	for _, f := range m {
		ret = append(ret, f)
	}
	return ret
}
func commandArgsForUsage(m []IArg) []IFlagArg {
	return lookupArgsForUsage(m)
}

func flagSorter(a IFlagArg, b IFlagArg) bool {

	fa, oka := a.(IFlag)
	fb, okb := b.(IFlag)
	if oka && okb {
		return fa.GetLevel() < fb.GetLevel()
	}
	return false

}

func validatableSorter(a IValidatable, b IValidatable) bool {

	fa, oka := a.(IFlag)
	fb, okb := b.(IFlag)
	if oka && okb {
		res := (fa.GetLevel() < fb.GetLevel())
		return res
	}
	return false

}

func mergeValidatables(g validationGroup) []IValidatable {

	return append(append(append(g.requiredFlags, g.optionalFlags...), g.requiredArgs...), g.optionalArgs...)
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
