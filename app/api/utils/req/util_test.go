// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package req

import (
	"reflect"
	"testing"

	commitModel "github.com/easysoft/gitfox-community/app/model/commit"
)

func TestSetDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name: "设置所有默认值",
			input: &struct {
				Page  int    `default:"1"`
				Size  int    `default:"20"`
				Sort  string `default:"name"`
				Order string `default:"asc"`
			}{},
			expected: &struct {
				Page  int    `default:"1"`
				Size  int    `default:"20"`
				Sort  string `default:"name"`
				Order string `default:"asc"`
			}{
				Page:  1,
				Size:  20,
				Sort:  "name",
				Order: "asc",
			},
		},
		{
			name: "部分字段已有值，不覆盖",
			input: &struct {
				Page  int    `default:"1"`
				Size  int    `default:"20"`
				Sort  string `default:"name"`
				Order string `default:"asc"`
			}{
				Page: 5,
				Sort: "date",
			},
			expected: &struct {
				Page  int    `default:"1"`
				Size  int    `default:"20"`
				Sort  string `default:"name"`
				Order string `default:"asc"`
			}{
				Page:  5,
				Size:  20,
				Sort:  "date",
				Order: "asc",
			},
		},
		{
			name: "支持 uint 类型",
			input: &struct {
				RepoID uint   `default:"100"`
				Name   string `default:"test"`
			}{},
			expected: &struct {
				RepoID uint   `default:"100"`
				Name   string `default:"test"`
			}{
				RepoID: 100,
				Name:   "test",
			},
		},
		{
			name: "支持 bool 类型",
			input: &struct {
				Enabled bool `default:"true"`
				Active  bool `default:"false"`
			}{},
			expected: &struct {
				Enabled bool `default:"true"`
				Active  bool `default:"false"`
			}{
				Enabled: true,
				Active:  false,
			},
		},
		{
			name: "无 default 标签，不设置值",
			input: &struct {
				Page int
				Size int
			}{},
			expected: &struct {
				Page int
				Size int
			}{
				Page: 0,
				Size: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDefaults(tt.input)

			// 使用反射比较结果
			if !compareStructs(tt.input, tt.expected) {
				t.Errorf("SetDefaults() failed for %s\nGot: %+v\nExpected: %+v", tt.name, tt.input, tt.expected)
			}
		})
	}
}

// TestSetDefaults_EmbeddedStruct 测试嵌入结构体的默认值设置
func TestSetDefaults_EmbeddedStruct(t *testing.T) {
	// 使用真实的 commitModel.ReqFilter 来测试嵌入结构体
	reqFilter := &commitModel.ReqFilter{}

	// 调用 SetDefaults
	SetDefaults(reqFilter)

	// 验证嵌入的 PaginationFilter 字段是否被设置了默认值
	if reqFilter.Page != 1 {
		t.Errorf("Expected Page=1, got %d", reqFilter.Page)
	}
	if reqFilter.PageSize != 20 {
		t.Errorf("Expected Limit=20, got %d", reqFilter.PageSize)
	}

	t.Logf("Successfully set defaults for embedded struct: Page=%d, Limit=%d", reqFilter.Page, reqFilter.PageSize)
}

func TestSetDefaults_InvalidInput(t *testing.T) {
	// 测试非指针类型 - 应该不会 panic
	notPointer := struct {
		Page int `default:"1"`
	}{}
	SetDefaults(notPointer) // 应该静默返回，不会崩溃

	// 测试非结构体类型
	var num int
	SetDefaults(&num) // 应该静默返回，不会崩溃
}

// compareStructs 比较两个结构体是否相等
func compareStructs(a, b interface{}) bool {
	aVal := getValue(a)
	bVal := getValue(b)

	if aVal.Kind() != bVal.Kind() {
		return false
	}

	if aVal.Kind() != reflect.Struct {
		return false
	}

	return compareStructValues(aVal, bVal)
}

// compareStructValues 递归比较两个结构体 Value
func compareStructValues(aVal, bVal reflect.Value) bool {
	for i := 0; i < aVal.NumField(); i++ {
		aField := aVal.Field(i)
		bField := bVal.Field(i)

		// 如果是嵌入的结构体，递归比较
		if aVal.Type().Field(i).Anonymous && aField.Kind() == reflect.Struct {
			if !compareStructValues(aField, bField) {
				return false
			}
			continue
		}

		switch aField.Kind() {
		case reflect.Struct:
			if !compareStructValues(aField, bField) {
				return false
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if aField.Int() != bField.Int() {
				return false
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if aField.Uint() != bField.Uint() {
				return false
			}
		case reflect.String:
			if aField.String() != bField.String() {
				return false
			}
		case reflect.Bool:
			if aField.Bool() != bField.Bool() {
				return false
			}
		case reflect.Float32, reflect.Float64:
			if aField.Float() != bField.Float() {
				return false
			}
		}
	}

	return true
}

func getValue(v interface{}) reflect.Value {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		return val.Elem()
	}
	return val
}
