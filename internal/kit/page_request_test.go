package kit

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// PageRequestSuite defines the test suite for PageRequest
type PageRequestSuite struct {
	suite.Suite
}

// Execute the test suite
func TestPageRequestSuite(t *testing.T) {
	suite.Run(t, new(PageRequestSuite))
}

func (s *PageRequestSuite) TestGetSort() {
	tests := []struct {
		name string
		pr   *PageRequest
		want PageRequest_Sort
	}{
		{
			name: "nil page request",
			pr:   nil,
			want: PageRequest_ASC,
		},
		{
			name: "with ASC sort",
			pr:   &PageRequest{Sort: PageRequest_ASC},
			want: PageRequest_ASC,
		},
		{
			name: "with DESC sort",
			pr:   &PageRequest{Sort: PageRequest_DESC},
			want: PageRequest_DESC,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := tt.pr.GetSort()
			s.Equal(tt.want, got)
		})
	}
}

func (s *PageRequestSuite) TestSetSort() {
	s.Run("set sort asc", func() {
		pr := &PageRequest{}
		result := pr.SetSortAsc()
		s.Equal(PageRequest_ASC, pr.Sort, "PageRequest.SetSortAsc() did not set Sort to ASC")
		s.Equal(pr, result, "PageRequest.SetSortAsc() did not return itself")
	})

	s.Run("set sort desc", func() {
		pr := &PageRequest{}
		result := pr.SetSortDesc()
		s.Equal(PageRequest_DESC, pr.Sort, "PageRequest.SetSortDesc() did not set Sort to DESC")
		s.Equal(pr, result, "PageRequest.SetSortDesc() did not return itself")
	})
}

func (s *PageRequestSuite) TestGetSortField() {
	tests := []struct {
		name string
		pr   *PageRequest
		want string
	}{
		{
			name: "nil page request",
			pr:   nil,
			want: DEFAULT_SORT_FIELD,
		},
		{
			name: "empty sort field",
			pr:   &PageRequest{SortField: ""},
			want: DEFAULT_SORT_FIELD,
		},
		{
			name: "camelCase sort field",
			pr:   &PageRequest{SortField: "camelCase"},
			want: "camel_case",
		},
		{
			name: "PascalCase sort field",
			pr:   &PageRequest{SortField: "PascalCase"},
			want: "pascal_case",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := tt.pr.GetSortField()
			s.Equal(tt.want, got)
		})
	}
}

func (s *PageRequestSuite) TestSetSortField() {
	pr := &PageRequest{}
	field := "testField"
	result := pr.SetSortField(field)

	s.Equal(field, pr.SortField, "PageRequest.SetSortField() did not set SortField correctly")
	s.Equal(pr, result, "PageRequest.SetSortField() did not return itself")
}

func (s *PageRequestSuite) TestGetPageSize() {
	tests := []struct {
		name string
		pr   *PageRequest
		want int
	}{
		{
			name: "nil page request",
			pr:   nil,
			want: 0,
		},
		{
			name: "zero page size",
			pr:   &PageRequest{PageSize: 0},
			want: DEFAULT_PAGE_ZISE,
		},
		{
			name: "page size less than minimum",
			pr:   &PageRequest{PageSize: MIN_PAGE_ZISE - 1},
			want: DEFAULT_PAGE_ZISE, // When PageSize is less than MIN_PAGE_ZISE, it returns DEFAULT_PAGE_ZISE (100) based on the implementation
		},
		{
			name: "page size greater than maximum",
			pr:   &PageRequest{PageSize: MAX_PAGE_ZISE + 1},
			want: MAX_PAGE_ZISE,
		},
		{
			name: "valid page size",
			pr:   &PageRequest{PageSize: 50},
			want: 50,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := tt.pr.GetPageSize()
			s.Equal(tt.want, got, "PageRequest.GetPageSize() returned incorrect size")
		})
	}
}

func (s *PageRequestSuite) TestSetPageSize() {
	pr := &PageRequest{}
	size := 50
	result := pr.SetPageSize(size)

	s.Equal(size, pr.PageSize, "PageRequest.SetPageSize() did not set PageSize correctly")
	s.Equal(pr, result, "PageRequest.SetPageSize() did not return itself")
}

func (s *PageRequestSuite) TestGetWay() {
	tests := []struct {
		name string
		pr   *PageRequest
		want bool
	}{
		{
			name: "nil page request",
			pr:   nil,
			want: false,
		},
		{
			name: "nil way",
			pr:   &PageRequest{Way: nil},
			want: false,
		},
		{
			name: "with page number",
			pr:   &PageRequest{Way: &PageRequest_PageNo{PageNo: 1}},
			want: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := tt.pr.GetWay() != nil
			s.Equal(tt.want, got)
		})
	}
}

func (s *PageRequestSuite) TestPageNoOperations() {
	s.Run("get page no when not set", func() {
		pr := &PageRequest{}
		pageNo, ok := pr.GetPageNo()
		s.Equal(0, pageNo)
		s.False(ok)
	})

	s.Run("set and get page no", func() {
		pr := &PageRequest{}
		pageNo := 5
		result := pr.SetPageNo(pageNo)

		s.Equal(pr, result, "PageRequest.SetPageNo() did not return itself")

		gotPageNo, ok := pr.GetPageNo()
		s.True(ok)
		s.Equal(pageNo, gotPageNo)
	})

	s.Run("increment page no", func() {
		pr := &PageRequest{}
		initialPageNo := 5
		pr.SetPageNo(initialPageNo)

		result := pr.IncrPageNo()
		s.Equal(pr, result, "PageRequest.IncrPageNo() did not return itself")

		gotPageNo, ok := pr.GetPageNo()
		s.True(ok)
		s.Equal(initialPageNo+1, gotPageNo)
	})
}

func (s *PageRequestSuite) TestCursorIDOperations() {
	s.Run("get cursor ID when not set", func() {
		pr := &PageRequest{}
		cursorID, ok := pr.GetCursorID()
		s.Equal(0, cursorID)
		s.False(ok)
	})

	s.Run("set and get cursor ID", func() {
		pr := &PageRequest{}
		cursorID := 100
		result := pr.SetCursorID(cursorID)

		s.Equal(pr, result, "PageRequest.SetCursorID() did not return itself")

		gotCursorID, ok := pr.GetCursorID()
		s.True(ok)
		s.Equal(cursorID, gotCursorID)
	})
}

func (s *PageRequestSuite) TestUUIDOperations() {
	s.Run("get UUID when not set", func() {
		pr := &PageRequest{}
		uuid, ok := pr.GetUUID()
		s.Equal("", uuid)
		s.False(ok)
	})

	s.Run("set and get UUID", func() {
		pr := &PageRequest{}
		uuid := "test-uuid-123"
		result := pr.SetUUID(uuid)

		s.Equal(pr, result, "PageRequest.SetUUID() did not return itself")

		gotUUID, ok := pr.GetUUID()
		s.True(ok)
		s.Equal(uuid, gotUUID)
	})
}

func (s *PageRequestSuite) TestTotalOperations() {
	s.Run("get total when not set", func() {
		pr := &PageRequest{}
		got := pr.GetTotal()
		s.Equal(0, got)
	})

	s.Run("set and get total", func() {
		pr := &PageRequest{}
		total := 100
		result := pr.SetTotal(total)

		s.Equal(pr, result, "PageRequest.SetTotal() did not return itself")
		s.Equal(total, pr.GetTotal())
	})
}

func (s *PageRequestSuite) TestHasMore() {
	tests := []struct {
		name string
		pr   *PageRequest
		want bool
	}{
		{
			name: "NoMore flag set",
			pr:   &PageRequest{NoMore: true},
			want: false,
		},
		{
			name: "no way set",
			pr:   &PageRequest{},
			want: true,
		},
		{
			name: "page number with zero total",
			pr:   &PageRequest{Way: &PageRequest_PageNo{PageNo: 1}, Total: 0},
			want: true,
		},
		{
			name: "page number within total pages",
			pr:   &PageRequest{Way: &PageRequest_PageNo{PageNo: 2}, Total: 30, PageSize: 10},
			want: true,
		},
		{
			name: "page number equal to total pages",
			pr:   &PageRequest{Way: &PageRequest_PageNo{PageNo: 3}, Total: 30, PageSize: 10},
			want: true,
		},
		{
			name: "page number exceeds total pages",
			pr:   &PageRequest{Way: &PageRequest_PageNo{PageNo: 4}, Total: 30, PageSize: 10},
			want: false,
		},
		{
			name: "cursor ID way always has more",
			pr:   &PageRequest{Way: &PageRequest_CursorID{CursorID: 100}},
			want: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := tt.pr.HasMore()
			s.Equal(tt.want, got)
		})
	}
}

func (s *PageRequestSuite) TestSetNoMore() {
	pr := &PageRequest{NoMore: false}
	result := pr.SetNoMore()

	s.True(pr.NoMore, "PageRequest.SetNoMore() did not set NoMore to true")
	s.Equal(pr, result, "PageRequest.SetNoMore() did not return itself")
}

func (s *PageRequestSuite) TestToSnakeCase() {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "already snake case",
			input: "snake_case",
			want:  "snake_case",
		},
		{
			name:  "camel case",
			input: "camelCase",
			want:  "camel_case",
		},
		{
			name:  "pascal case",
			input: "PascalCase",
			want:  "pascal_case",
		},
		{
			name:  "multiple uppercase letters in sequence",
			input: "HTTPRequest",
			want:  "http_request",
		},
		{
			name:  "single letter",
			input: "a",
			want:  "a",
		},
	}

	pr := &PageRequest{}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := pr.toSnakeCase(tt.input)
			s.Equal(tt.want, got)
		})
	}
}
