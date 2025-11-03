package database

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

// PageParam 分页查询参数（入参）
// 前端传递时只需指定 Page/PageSize，Sort/Order 可选（有默认值）
type PageParam struct {
	Page     int    `json:"page"`      // 页码（默认1）
	PageSize int    `json:"pageSize"`  // 每页条数（默认20，最大100）
	Sort     string `json:"sort"`      // 排序字段（默认"createdDate"）
	Order    string `json:"order"`     // 排序方向（desc/asc，默认"desc"）
	NeedTotal bool  `json:"needTotal"`// 是否需要总条数（默认true，某些场景可关闭提升性能）
}

// PageResult 分页查询结果（出参）
// 包含数据列表+分页元信息，直接返回给前端
type PageResult[T any] struct {
	Data       []T   `json:"data"`       // 当前页数据列表
	Total      int64 `json:"total"`      // 总条数（NeedTotal=true时返回）
	Page       int   `json:"page"`       // 当前页码
	PageSize   int   `json:"pageSize"`   // 每页条数
	TotalPage  int64 `json:"totalPage"`  // 总页数（Total/PageSize 向上取整）
}

// 默认配置
const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
	defaultSort     = "createdDate"
	defaultOrder    = "desc"
)

// Paginate 通用分页查询函数
// 参数说明：
// - ctx: 上下文（传递超时、用户信息等）
// - db: GORM查询对象（已拼接过滤条件的 tx，如 db.Where("deleted=0")）
// - param: 分页参数（PageParam）
// - dataPtr: 数据接收指针（必须是 []*Model 或 []Model 类型，如 &[]*spaceModel.Space{}）
// 返回值：
// - PageResult[T]: 分页结果（包含数据和分页元信息）
// - error: 错误信息
func Paginate[T any](ctx context.Context, db *gorm.DB, param PageParam, dataPtr *[]T) (*PageResult[T], error) {
	// 1. 处理分页参数默认值
	if param.Page <= 0 {
		param.Page = defaultPage
	}
	if param.PageSize <= 0 || param.PageSize > maxPageSize {
		param.PageSize = defaultPageSize
	}
	if param.Sort == "" {
		param.Sort = defaultSort
	}
	if param.Order == "" {
		param.Order = defaultOrder
	}
	// 标准化排序方向（仅支持 desc/asc）
	param.Order = strings.ToLower(param.Order)
	if param.Order != "desc" && param.Order != "asc" {
		param.Order = defaultOrder
	}

	// 2. 校验数据接收指针类型（必须是切片指针）
	if !isSlicePtr(dataPtr) {
		return nil, errors.New("dataPtr must be a pointer to slice (e.g. &[]*Model{})")
	}

	// 3. 构建查询对象（绑定上下文）
	tx := db.WithContext(ctx)

	// 4. 统计总条数（如果需要）
	var total int64
	if param.NeedTotal {
		if err := tx.Count(&total).Error; err != nil {
			return nil, fmt.Errorf("count total failed: %w", err)
		}
	}

	// 5. 计算分页偏移量
	offset := (param.Page - 1) * param.PageSize

	// 6. 拼接排序条件（支持多字段排序，用逗号分隔，如 "createdDate desc, name asc"）
	sortStr := buildSortString(param.Sort, param.Order)
	tx = tx.Order(sortStr)

	// 7. 执行分页查询（Limit + Offset）
	if err := tx.Limit(param.PageSize).Offset(offset).Find(dataPtr).Error; err != nil {
		return nil, fmt.Errorf("query page data failed: %w", err)
	}

	// 8. 计算总页数
	totalPage := int64(0)
	if param.NeedTotal && param.PageSize > 0 {
		totalPage = (total + int64(param.PageSize) - 1) / int64(param.PageSize) // 向上取整
	}

	// 9. 构建并返回分页结果
	return &PageResult[T]{
		Data:      *dataPtr,
		Total:     total,
		Page:      param.Page,
		PageSize:  param.PageSize,
		TotalPage: totalPage,
	}, nil
}

// isSlicePtr 判断是否为切片指针（如 &[]*Model{} 或 &[]Model{}）
func isSlicePtr(v interface{}) bool {
	if v == nil {
		return false
	}
	typ := reflect.TypeOf(v)
	return typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Slice
}

// buildSortString 构建排序字符串（支持多字段排序）
// 示例：
// - sort="createdDate", order="desc" → "createdDate desc"
// - sort="createdDate,name", order="desc" → "createdDate desc, name desc"
// - sort="createdDate,name", order="desc,asc" → "createdDate desc, name asc"
func buildSortString(sort, order string) string {
	sortFields := strings.Split(sort, ",")
	orderDirs := strings.Split(order, ",")

	// 处理排序方向与字段数量不一致的情况
	if len(orderDirs) != len(sortFields) {
		// 用默认方向填充
		defaultDir := defaultOrder
		if len(orderDirs) > 0 && (orderDirs[0] == "desc" || orderDirs[0] == "asc") {
			defaultDir = orderDirs[0]
		}
		orderDirs = make([]string, len(sortFields))
		for i := range orderDirs {
			orderDirs[i] = defaultDir
		}
	}

	// 拼接每个字段的排序规则
	var sortParts []string
	for i, field := range sortFields {
		field = strings.TrimSpace(field)
		dir := strings.TrimSpace(orderDirs[i])
		if field == "" {
			continue
		}
		// 确保排序方向合法
		if dir != "desc" && dir != "asc" {
			dir = defaultOrder
		}
		sortParts = append(sortParts, fmt.Sprintf("%s %s", field, dir))
	}

	// 若没有有效排序字段，使用默认排序
	if len(sortParts) == 0 {
		return fmt.Sprintf("%s %s", defaultSort, defaultOrder)
	}

	return strings.Join(sortParts, ", ")
}

// ------------- 可选：快捷函数（简化常用场景调用）-------------

// PaginateDefault 简化分页查询（使用默认排序：createdDate desc）
func PaginateDefault[T any](ctx context.Context, db *gorm.DB, page, pageSize int, dataPtr *[]T) (*PageResult[T], error) {
	return Paginate(ctx, db, PageParam{
		Page:     page,
		PageSize: pageSize,
		NeedTotal: true,
	}, dataPtr)
}

// PaginateNoTotal 分页查询（不统计总条数，提升性能）
func PaginateNoTotal[T any](ctx context.Context, db *gorm.DB, page, pageSize int, dataPtr *[]T) (*PageResult[T], error) {
	return Paginate(ctx, db, PageParam{
		Page:     page,
		PageSize: pageSize,
		NeedTotal: false,
	}, dataPtr)
}