// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package slice

import (
	"fmt"
	"strings"

	"golang.org/x/exp/constraints"
)

// Contain check if the target value is in the slice or not.
// 检查目标值是否在切片中。
func Contain[T comparable](slice []T, target T) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}

	return false
}

// ToSlice returns a slices of a variable parameter transformation.
// 将可变参数转换为切片。
func ToSlice[T any](items ...T) []T {
	result := make([]T, len(items))
	copy(result, items)

	return result
}

// AppendIfAbsent only absent append the item.
// 如果切片中不存在目标值，则追加该值。
func AppendIfAbsent[T comparable](slice []T, item T) []T {
	if !Contain(slice, item) {
		slice = append(slice, item)
	}
	return slice
}

// Join the slice item with specify separator.
func Join[T any](slice []T, separator string) string {
	str := Map(slice, func(_ int, item T) string {
		return fmt.Sprint(item)
	})

	return strings.Join(str, separator)
}

// Map creates an slice of values by running each element of slice thru iteratee function.
func Map[T any, U any](slice []T, iteratee func(index int, item T) U) []U {
	result := make([]U, len(slice), cap(slice))

	for i := 0; i < len(slice); i++ {
		result[i] = iteratee(i, slice[i])
	}

	return result
}

// JoinFunc joins the slice elements into a single string with the given separator.
// 将切片元素连接为单个字符串，使用给定的分隔符。
func JoinFunc[T any](slice []T, sep string, transform func(T) T) string {
	var buf strings.Builder
	for i, v := range slice {
		if i > 0 {
			buf.WriteString(sep)
		}
		buf.WriteString(fmt.Sprint(transform(v)))
	}
	return buf.String()
}

// Sort sorts a slice of any ordered type(number or string), use quick sort algrithm.
// default sort order is ascending (asc), if want descending order, set param `sortOrder` to `desc`.
// 使用快速排序算法对任何有序类型（数字或字符串）的切片进行排序。
// 排序顺序可以是升序（asc）或降序（desc），默认是升序。
func Sort[T constraints.Ordered](slice []T, sortOrder ...string) {
	if len(sortOrder) > 0 && sortOrder[0] == "desc" {
		quickSort(slice, 0, len(slice)-1, "desc")
	} else {
		quickSort(slice, 0, len(slice)-1, "asc")
	}
}

// SortBy sorts the slice in ascending order as determined by the less function.
// This sort is not guaranteed to be stable.
// SortBy根据less函数确定的顺序对切片进行排序。
// 这种类型并不能保证稳定。
func SortBy[T any](slice []T, less func(a, b T) bool) {
	quickSortBy(slice, 0, len(slice)-1, less)
}

// Union creates a slice of unique elements, in order, from all given slices.
// 求并集
func Union[T comparable](slices ...[]T) []T {
	result := []T{}
	contain := map[T]struct{}{}

	for _, slice := range slices {
		for _, item := range slice {
			if _, ok := contain[item]; !ok {
				contain[item] = struct{}{}
				result = append(result, item)
			}
		}
	}

	return result
}

// Intersection creates a slice of unique elements that included by all slices.
// 求交集
func Intersection[T comparable](slices ...[]T) []T {
	result := []T{}
	elementCount := make(map[T]int)

	for _, slice := range slices {
		seen := make(map[T]bool)

		for _, item := range slice {
			if !seen[item] {
				seen[item] = true
				elementCount[item]++
			}
		}
	}

	for _, item := range slices[0] {
		if elementCount[item] == len(slices) {
			result = append(result, item)
			elementCount[item] = 0
		}
	}

	return result
}

// Unique remove duplicate elements in slice.
// 简单去重
func Unique[T comparable](slice []T) []T {
	if len(slice) == 0 {
		return slice
	}

	seen := make(map[T]struct{}, len(slice))
	result := slice[:0]

	for _, item := range slice {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}
