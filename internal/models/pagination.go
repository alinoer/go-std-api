package models

type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Offset   int `json:"offset"`
}

type PaginationMeta struct {
	Page         int   `json:"page"`
	PageSize     int   `json:"page_size"`
	Total        int64 `json:"total"`
	TotalPages   int   `json:"total_pages"`
	HasNext      bool  `json:"has_next"`
	HasPrevious  bool  `json:"has_previous"`
}

type PaginatedResponse struct {
	Data       interface{}     `json:"data"`
	Pagination *PaginationMeta `json:"pagination"`
}

func NewPaginationParams(page, pageSize int) *PaginationParams {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10 // default page size
	}

	offset := (page - 1) * pageSize

	return &PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Offset:   offset,
	}
}

func NewPaginationMeta(page, pageSize int, total int64) *PaginationMeta {
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize)) // ceiling division
	
	return &PaginationMeta{
		Page:         page,
		PageSize:     pageSize,
		Total:        total,
		TotalPages:   totalPages,
		HasNext:      page < totalPages,
		HasPrevious:  page > 1,
	}
}