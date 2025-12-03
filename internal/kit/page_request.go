package kit

import "math"

const (
	MIN_PAGE_ZISE      = 1
	MAX_PAGE_ZISE      = 2000
	DEFAULT_PAGE_ZISE  = 100
	DEFAULT_SORT_FIELD = "id"
)

type PageRequest struct {
	Sort      PageRequest_Sort
	SortField string
	PageSize  int
	Way       isPageRequest_Way
	Total     int
	NoMore    bool
}

type PageRequest_Sort int

const (
	PageRequest_ASC  PageRequest_Sort = 0
	PageRequest_DESC PageRequest_Sort = 1
)

func (x *PageRequest) GetSort() PageRequest_Sort {
	if x != nil {
		return x.Sort
	}
	return PageRequest_ASC
}

func (x *PageRequest) SetSortAsc() *PageRequest {
	x.Sort = PageRequest_ASC
	return x
}

func (x *PageRequest) SetSortDesc() *PageRequest {
	x.Sort = PageRequest_DESC
	return x
}

func (x *PageRequest) GetSortField() string {
	if x != nil && x.SortField != "" {
		return x.toSnakeCase(x.SortField)
	}
	return DEFAULT_SORT_FIELD
}

func (x *PageRequest) toSnakeCase(s string) string {
	const (
		lower = false
		upper = true
	)

	if s == "" {
		return s
	}

	var result []rune
	var lastCase, currCase, nextCase bool

	for i, char := range s {
		nextCase = i+1 < len(s) && s[i+1] >= 'A' && s[i+1] <= 'Z'

		if i > 0 && char >= 'A' && char <= 'Z' && (!lastCase || !nextCase) {
			result = append(result, '_')
		}

		if char >= 'A' && char <= 'Z' {
			currCase = upper
			result = append(result, char+32)
		} else {
			currCase = lower
			result = append(result, char)
		}

		lastCase = currCase
	}

	return string(result)
}

func (x *PageRequest) SetSortField(field string) *PageRequest {
	x.SortField = field
	return x
}

func (x *PageRequest) GetPageSize() int {
	if x != nil {
		if x.PageSize == 0 {
			return DEFAULT_PAGE_ZISE
		}
		if x.PageSize < MIN_PAGE_ZISE {
			return MIN_PAGE_ZISE
		}
		if x.PageSize > MAX_PAGE_ZISE {
			return MAX_PAGE_ZISE
		}
		return x.PageSize
	}
	return 0
}

func (x *PageRequest) SetPageSize(pageSize int) *PageRequest {
	x.PageSize = pageSize
	return x
}

func (m *PageRequest) GetWay() isPageRequest_Way {
	if m != nil {
		return m.Way
	}
	return nil
}

func (x *PageRequest) GetPageNo() (pageNo int, ok bool) {
	way := x.GetWay()
	if way == nil {
		return 0, false
	}
	if x, ok := way.(*PageRequest_PageNo); ok {
		return x.PageNo, true
	}
	return 0, false
}

func (x *PageRequest) SetPageNo(pageNo int) *PageRequest {
	x.Way = &PageRequest_PageNo{
		PageNo: pageNo,
	}
	return x
}

func (x *PageRequest) IncrPageNo() *PageRequest {
	if pageNo, ok := x.GetPageNo(); ok {
		x.GetWay().(*PageRequest_PageNo).PageNo = pageNo + 1
	}
	return x
}

func (x *PageRequest) GetCursorID() (cursorID int, ok bool) {
	way := x.GetWay()
	if way == nil {
		return 0, false
	}
	if x, ok := way.(*PageRequest_CursorID); ok {
		return x.CursorID, true
	}
	return 0, false
}

func (x *PageRequest) SetCursorID(cursorID int) *PageRequest {
	x.Way = &PageRequest_CursorID{
		CursorID: cursorID,
	}
	return x
}

func (x *PageRequest) GetUUID() (uuid string, ok bool) {
	way := x.GetWay()
	if way == nil {
		return "", false
	}
	if x, ok := way.(*PageRequest_UUID); ok {
		return x.UUID, true
	}
	return "", false
}

func (x *PageRequest) SetUUID(uuid string) *PageRequest {
	x.Way = &PageRequest_UUID{
		UUID: uuid,
	}
	return x
}

func (x *PageRequest) GetTotal() int {
	if x != nil {
		return x.Total
	}
	return 0
}

func (x *PageRequest) SetTotal(total int) *PageRequest {
	x.Total = total
	return x
}

func (x *PageRequest) SetNoMore() *PageRequest {
	x.NoMore = true
	return x
}

func (x *PageRequest) HasMore() bool {
	if x.NoMore {
		return false
	}
	way := x.GetWay()
	if pageNo, ok := way.(*PageRequest_PageNo); ok {
		if x.GetTotal() == 0 {
			return true
		}
		totalPage := int(math.Ceil(float64(x.GetTotal()) / float64(x.GetPageSize())))
		return pageNo.PageNo <= totalPage
	}
	return true
}

type isPageRequest_Way interface {
	isPageRequest_Way()
}

type PageRequest_PageNo struct {
	PageNo int
}

type PageRequest_CursorID struct {
	CursorID int
}

type PageRequest_UUID struct {
	UUID string
}

func (*PageRequest_PageNo) isPageRequest_Way() {}

func (*PageRequest_CursorID) isPageRequest_Way() {}

func (*PageRequest_UUID) isPageRequest_Way() {}
