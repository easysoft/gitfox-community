// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package slice

import (
	"fmt"
	"strings"
)

func ExampleContain() {
	result1 := Contain([]string{"a", "b", "c"}, "a")
	result2 := Contain([]int{1, 2, 3}, 4)

	fmt.Println(result1)
	fmt.Println(result2)

	// Output:
	// true
	// false
}

func ExampleToSlice() {
	result := ToSlice("a", "b", "c")

	fmt.Println(result)

	// Output:
	// [a b c]
}

func ExampleAppendIfAbsent() {
	result1 := AppendIfAbsent([]string{"a", "b"}, "b")
	result2 := AppendIfAbsent([]string{"a", "b"}, "c")

	fmt.Println(result1)
	fmt.Println(result2)

	// Output:
	// [a b]
	// [a b c]
}

func ExampleJoinFunc() {
	result := JoinFunc([]string{"a", "b", "c"}, ", ", func(s string) string {
		return strings.ToUpper(s)
	})

	fmt.Println(result)

	// Output:
	// A, B, C
}

func ExampleSort() {
	nums := []int{1, 4, 3, 2, 5}

	Sort(nums)

	fmt.Println(nums)

	// Output:
	// [1 2 3 4 5]
}

func ExampleSortBy() {
	type User struct {
		Name string
		Age  uint
	}

	users := []User{
		{Name: "a", Age: 21},
		{Name: "b", Age: 15},
		{Name: "c", Age: 100}}

	SortBy(users, func(a, b User) bool {
		return a.Age < b.Age
	})

	fmt.Println(users)

	// Output:
	// [{b 15} {a 21} {c 100}]
}

func ExampleUnion() {
	nums1 := []int{1, 3, 4, 6}
	nums2 := []int{1, 2, 5, 6}

	result := Union(nums1, nums2)

	fmt.Println(result)

	// Output:
	// [1 3 4 6 2 5]
}

func ExampleIntersection() {
	nums1 := []int{1, 2, 3}
	nums2 := []int{2, 3, 4}

	result := Intersection(nums1, nums2)

	fmt.Println(result)

	// Output:
	// [2 3]
}

func ExampleUnique() {
	result := Unique([]string{"a", "a", "b"})
	fmt.Println(result)

	// Output:
	// [a b]
}
