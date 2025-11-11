package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/alinoer/go-std-api/internal/models"
)

func TestParsePaginationParams(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedPage   int
		expectedSize   int
		expectedOffset int
	}{
		{
			name:           "no query parameters",
			queryParams:    map[string]string{},
			expectedPage:   1,
			expectedSize:   10,
			expectedOffset: 0,
		},
		{
			name: "valid page and page_size",
			queryParams: map[string]string{
				"page":      "3",
				"page_size": "20",
			},
			expectedPage:   3,
			expectedSize:   20,
			expectedOffset: 40,
		},
		{
			name: "only page parameter",
			queryParams: map[string]string{
				"page": "2",
			},
			expectedPage:   2,
			expectedSize:   10,
			expectedOffset: 10,
		},
		{
			name: "only page_size parameter",
			queryParams: map[string]string{
				"page_size": "25",
			},
			expectedPage:   1,
			expectedSize:   25,
			expectedOffset: 0,
		},
		{
			name: "invalid page (not a number)",
			queryParams: map[string]string{
				"page":      "abc",
				"page_size": "15",
			},
			expectedPage:   1,
			expectedSize:   15,
			expectedOffset: 0,
		},
		{
			name: "invalid page_size (not a number)",
			queryParams: map[string]string{
				"page":      "2",
				"page_size": "xyz",
			},
			expectedPage:   2,
			expectedSize:   10,
			expectedOffset: 10,
		},
		{
			name: "negative page",
			queryParams: map[string]string{
				"page":      "-1",
				"page_size": "15",
			},
			expectedPage:   1,
			expectedSize:   15,
			expectedOffset: 0,
		},
		{
			name: "zero page",
			queryParams: map[string]string{
				"page":      "0",
				"page_size": "15",
			},
			expectedPage:   1,
			expectedSize:   15,
			expectedOffset: 0,
		},
		{
			name: "negative page_size",
			queryParams: map[string]string{
				"page":      "2",
				"page_size": "-5",
			},
			expectedPage:   2,
			expectedSize:   10,
			expectedOffset: 10,
		},
		{
			name: "zero page_size",
			queryParams: map[string]string{
				"page":      "2",
				"page_size": "0",
			},
			expectedPage:   2,
			expectedSize:   10,
			expectedOffset: 10,
		},
		{
			name: "page_size greater than 100",
			queryParams: map[string]string{
				"page":      "1",
				"page_size": "150",
			},
			expectedPage:   1,
			expectedSize:   10,
			expectedOffset: 0,
		},
		{
			name: "page_size exactly 100",
			queryParams: map[string]string{
				"page":      "1",
				"page_size": "100",
			},
			expectedPage:   1,
			expectedSize:   100,
			expectedOffset: 0,
		},
		{
			name: "page_size exactly 1",
			queryParams: map[string]string{
				"page":      "5",
				"page_size": "1",
			},
			expectedPage:   5,
			expectedSize:   1,
			expectedOffset: 4,
		},
		{
			name: "empty string values",
			queryParams: map[string]string{
				"page":      "",
				"page_size": "",
			},
			expectedPage:   1,
			expectedSize:   10,
			expectedOffset: 0,
		},
		{
			name: "very large page number",
			queryParams: map[string]string{
				"page":      "1000",
				"page_size": "50",
			},
			expectedPage:   1000,
			expectedSize:   50,
			expectedOffset: 49950,
		},
		{
			name: "decimal page number (should be ignored)",
			queryParams: map[string]string{
				"page":      "2.5",
				"page_size": "20",
			},
			expectedPage:   1,
			expectedSize:   20,
			expectedOffset: 0,
		},
		{
			name: "decimal page_size (should be ignored)",
			queryParams: map[string]string{
				"page":      "2",
				"page_size": "15.7",
			},
			expectedPage:   2,
			expectedSize:   10,
			expectedOffset: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create URL with query parameters
			u, _ := url.Parse("http://example.com/test")
			q := u.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			u.RawQuery = q.Encode()

			// Create HTTP request
			req := httptest.NewRequest(http.MethodGet, u.String(), nil)

			// Parse pagination parameters
			pagination := ParsePaginationParams(req)

			// Verify results
			if pagination.Page != tt.expectedPage {
				t.Errorf("expected page %d, got %d", tt.expectedPage, pagination.Page)
			}

			if pagination.PageSize != tt.expectedSize {
				t.Errorf("expected page size %d, got %d", tt.expectedSize, pagination.PageSize)
			}

			if pagination.Offset != tt.expectedOffset {
				t.Errorf("expected offset %d, got %d", tt.expectedOffset, pagination.Offset)
			}
		})
	}
}

func TestParsePaginationParams_Integration(t *testing.T) {
	// Test that the function properly integrates with models.NewPaginationParams
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test?page=3&page_size=25", nil)
	
	pagination := ParsePaginationParams(req)
	
	// Verify that it returns a proper PaginationParams object
	if pagination == nil {
		t.Error("expected non-nil pagination params")
		return
	}

	// Test that it has all required fields
	expectedPage := 3
	expectedPageSize := 25
	expectedOffset := 50 // (3-1) * 25

	if pagination.Page != expectedPage {
		t.Errorf("expected page %d, got %d", expectedPage, pagination.Page)
	}

	if pagination.PageSize != expectedPageSize {
		t.Errorf("expected page size %d, got %d", expectedPageSize, pagination.PageSize)
	}

	if pagination.Offset != expectedOffset {
		t.Errorf("expected offset %d, got %d", expectedOffset, pagination.Offset)
	}
}

func TestParsePaginationParams_DefaultBehavior(t *testing.T) {
	// Test default behavior when no query parameters are provided
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	
	pagination := ParsePaginationParams(req)
	
	// Should return default pagination params
	defaultPagination := models.NewPaginationParams(1, 10)
	
	if pagination.Page != defaultPagination.Page {
		t.Errorf("expected default page %d, got %d", defaultPagination.Page, pagination.Page)
	}

	if pagination.PageSize != defaultPagination.PageSize {
		t.Errorf("expected default page size %d, got %d", defaultPagination.PageSize, pagination.PageSize)
	}

	if pagination.Offset != defaultPagination.Offset {
		t.Errorf("expected default offset %d, got %d", defaultPagination.Offset, pagination.Offset)
	}
}

func TestParsePaginationParams_BoundaryValues(t *testing.T) {
	tests := []struct {
		name        string
		page        string
		pageSize    string
		expectValid bool
	}{
		{
			name:        "minimum valid values",
			page:        "1",
			pageSize:    "1",
			expectValid: true,
		},
		{
			name:        "maximum valid page size",
			page:        "1",
			pageSize:    "100",
			expectValid: true,
		},
		{
			name:        "page size over limit",
			page:        "1",
			pageSize:    "101",
			expectValid: false,
		},
		{
			name:        "page at boundary",
			page:        "1",
			pageSize:    "10",
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse("http://example.com/test")
			q := u.Query()
			q.Add("page", tt.page)
			q.Add("page_size", tt.pageSize)
			u.RawQuery = q.Encode()

			req := httptest.NewRequest(http.MethodGet, u.String(), nil)
			pagination := ParsePaginationParams(req)

			if tt.expectValid {
				// Should use the provided values
				if pagination.PageSize > 100 || pagination.PageSize < 1 {
					t.Errorf("expected valid page size, got %d", pagination.PageSize)
				}
				if pagination.Page < 1 {
					t.Errorf("expected valid page, got %d", pagination.Page)
				}
			} else {
				// Should fall back to defaults for invalid values
				if pagination.PageSize != 10 {
					t.Errorf("expected default page size 10 for invalid input, got %d", pagination.PageSize)
				}
			}
		})
	}
}