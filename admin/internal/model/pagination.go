package model

const (
	// DefaultPage 默认页码。
	DefaultPage = 1
	// DefaultPageSize 默认每页记录数。
	DefaultPageSize = 10
	// MaxPageSize 单次查询允许的最大记录数。
	MaxPageSize = 1000
)

// PageRequest 分页查询参数。
type PageRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// NewPageRequest 创建经过边界修正的分页参数。
func NewPageRequest(page int, pageSize int) PageRequest {
	if page < 1 {
		page = DefaultPage
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return PageRequest{Page: page, PageSize: pageSize}
}

// PageResult 分页查询结果。
type PageResult[T any] struct {
	List     []T `json:"list"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// NewPageResult 创建分页查询结果。
func NewPageResult[T any](list []T, total int, page PageRequest) *PageResult[T] {
	return &PageResult[T]{
		List:     list,
		Total:    total,
		Page:     page.Page,
		PageSize: page.PageSize,
	}
}
