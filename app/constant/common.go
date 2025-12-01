package constant

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

// 常量：和 MySQL varchar(500) 对齐
const (
	MaxDescCharCount        = 500                  // 最大字符数（核心限定）
	MaxDescByteCountUTF8MB4 = MaxDescCharCount * 4 // 2000 字节
)

type ContentEncodingType string

// 编码
const (
	ContentEncodingTypeUTF8   ContentEncodingType = "utf8"
	ContentEncodingTypeBase64 ContentEncodingType = "base64"
)

type RespCode string

const (
	RespSuccessCode RespCode = "success"
	RespFailCode    RespCode = "failure"
)

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

func toInterfaceSlice[T interface{}](vals []T) []interface{} {
	res := make([]interface{}, len(vals))
	for i := range vals {
		res[i] = vals[i]
	}
	return res
}

func sortEnum[T constraints.Ordered](slice []T) []T {
	slices.Sort(slice)
	return slice
}
