package models

import (
	"testing"
)

func TestNewPaginationParams(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		expectedPage int
		expectedSize int
		expectedOffset int
	}{
		{
			name:           "valid parameters",
			page:           2,
			pageSize:       20,
			expectedPage:   2,
			expectedSize:   20,
			expectedOffset: 20,
		},
		{
			name:           "page less than 1",
			page:           0,
			pageSize:       10,
			expectedPage:   1,
			expectedSize:   10,
			expectedOffset: 0,
		},
		{
			name:           "negative page",
			page:           -5,
			pageSize:       10,
			expectedPage:   1,
			expectedSize:   10,
			expectedOffset: 0,
		},
		{
			name:           "page size less than 1",
			page:           1,
			pageSize:       0,
			expectedPage:   1,
			expectedSize:   10,
			expectedOffset: 0,
		},
		{
			name:           "page size greater than 100",
			page:           1,
			pageSize:       150,
			expectedPage:   1,
			expectedSize:   10,
			expectedOffset: 0,
		},
		{
			name:           "negative page size",
			page:           1,
			pageSize:       -10,
			expectedPage:   1,
			expectedSize:   10,
			expectedOffset: 0,
		},
		{
			name:           "valid boundary values",
			page:           1,
			pageSize:       100,
			expectedPage:   1,
			expectedSize:   100,
			expectedOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := NewPaginationParams(tt.page, tt.pageSize)

			if params.Page != tt.expectedPage {
				t.Errorf("expected page %d, got %d", tt.expectedPage, params.Page)
			}
			if params.PageSize != tt.expectedSize {
				t.Errorf("expected page size %d, got %d", tt.expectedSize, params.PageSize)
			}
			if params.Offset != tt.expectedOffset {
				t.Errorf("expected offset %d, got %d", tt.expectedOffset, params.Offset)
			}
		})
	}
}

func TestNewPaginationMeta(t *testing.T) {
	tests := []struct {
		name            string
		page            int
		pageSize        int
		total           int64
		expectedPages   int
		expectedHasNext bool
		expectedHasPrev bool
	}{
		{
			name:            "first page with more pages",
			page:            1,
			pageSize:        10,
			total:           25,
			expectedPages:   3,
			expectedHasNext: true,
			expectedHasPrev: false,
		},
		{
			name:            "middle page",
			page:            2,
			pageSize:        10,
			total:           25,
			expectedPages:   3,
			expectedHasNext: true,
			expectedHasPrev: true,
		},
		{
			name:            "last page",
			page:            3,
			pageSize:        10,
			total:           25,
			expectedPages:   3,
			expectedHasNext: false,
			expectedHasPrev: true,
		},
		{
			name:            "single page",
			page:            1,
			pageSize:        10,
			total:           5,
			expectedPages:   1,
			expectedHasNext: false,
			expectedHasPrev: false,
		},
		{
			name:            "exact page boundary",
			page:            1,
			pageSize:        10,
			total:           10,
			expectedPages:   1,
			expectedHasNext: false,
			expectedHasPrev: false,
		},
		{
			name:            "zero total",
			page:            1,
			pageSize:        10,
			total:           0,
			expectedPages:   0,
			expectedHasNext: false,
			expectedHasPrev: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := NewPaginationMeta(tt.page, tt.pageSize, tt.total)

			if meta.Page != tt.page {
				t.Errorf("expected page %d, got %d", tt.page, meta.Page)
			}
			if meta.PageSize != tt.pageSize {
				t.Errorf("expected page size %d, got %d", tt.pageSize, meta.PageSize)
			}
			if meta.Total != tt.total {
				t.Errorf("expected total %d, got %d", tt.total, meta.Total)
			}
			if meta.TotalPages != tt.expectedPages {
				t.Errorf("expected total pages %d, got %d", tt.expectedPages, meta.TotalPages)
			}
			if meta.HasNext != tt.expectedHasNext {
				t.Errorf("expected has next %t, got %t", tt.expectedHasNext, meta.HasNext)
			}
			if meta.HasPrevious != tt.expectedHasPrev {
				t.Errorf("expected has previous %t, got %t", tt.expectedHasPrev, meta.HasPrevious)
			}
		})
	}
}