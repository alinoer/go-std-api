package handlers

import (
	"net/http"
	"strconv"

	"github.com/alinoer/go-std-api/internal/models"
)

func ParsePaginationParams(r *http.Request) *models.PaginationParams {
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	pageSize := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	return models.NewPaginationParams(page, pageSize)
}