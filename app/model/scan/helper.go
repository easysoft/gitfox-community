package scan

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

// Sanitize 验证元素是否在允许的值列表中，如果不在则返回默认值
func Sanitize[E constraints.Ordered](element E, all func() ([]E, E)) (E, bool) {
	allValues, defValue := all()
	var empty E
	if element == empty && defValue != empty {
		return defValue, true
	}
	idx, exists := slices.BinarySearch(allValues, element)
	if exists {
		return allValues[idx], true
	}
	return defValue, false
}

// toInterfaceSlice 将类型化切片转换为 interface{} 切片
func toInterfaceSlice[T interface{}](vals []T) []interface{} {
	res := make([]interface{}, len(vals))
	for i := range vals {
		res[i] = vals[i]
	}
	return res
}

// sortEnum 对枚举切片进行排序
func sortEnum[T constraints.Ordered](slice []T) []T {
	slices.Sort(slice)
	return slice
}
