package artifact

import (
	"reflect"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
)

func validatorEmpty(args ...any) error {
	for _, arg := range args {
		if isEmpty(arg) {
			return model.ErrArgsValueEmpty
		}
	}
	return nil
}

func isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	switch reflect.ValueOf(value).Kind() {
	case reflect.String:
		return reflect.ValueOf(value).String() == ""
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return reflect.ValueOf(value).Int() == 0
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return reflect.ValueOf(value).Uint() == 0
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(value).Float() == 0.0
	case reflect.Slice, reflect.Array, reflect.Map:
		return reflect.ValueOf(value).Len() == 0
	default:
		return false
	}
}
