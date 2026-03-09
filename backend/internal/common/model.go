package common

type PageResult[T any] struct {
	Records []T `json:"records"` // 当前页的记录数据
	Total   int `json:"total"`   // 总记录数
	Pages   int `json:"pages"`   // 总页数
	Size    int `json:"size"`    // 每页的记录数
	Current int `json:"current"` // 当前页
}
