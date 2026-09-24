package model

import "testing"

func TestNewPageRequest(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		want     PageRequest
	}{
		{name: "defaults", want: PageRequest{Page: DefaultPage, PageSize: DefaultPageSize}},
		{name: "values", page: 3, pageSize: 20, want: PageRequest{Page: 3, PageSize: 20}},
		{name: "maximum", page: 2, pageSize: MaxPageSize + 1, want: PageRequest{Page: 2, PageSize: MaxPageSize}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := NewPageRequest(test.page, test.pageSize); got != test.want {
				t.Fatalf("NewPageRequest() = %#v, want %#v", got, test.want)
			}
		})
	}
}
