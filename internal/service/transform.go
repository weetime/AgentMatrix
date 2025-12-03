package service

import (
	"nova/internal/kit"

	v1 "nova/protos/nova/v1"
)

func apiToPageRequest(to *kit.PageRequest, from *v1.PageRequest) {
	to.PageSize = int(from.GetPageSize())
	to.Sort = kit.PageRequest_Sort(from.GetSort())
	to.SortField = from.GetSortField()
	if x, ok := from.GetWay().(*v1.PageRequest_CursorID); ok {
		to.Way = &kit.PageRequest_CursorID{CursorID: int(x.CursorID)}
	}
	if x, ok := from.GetWay().(*v1.PageRequest_PageNo); ok {
		to.Way = &kit.PageRequest_PageNo{PageNo: int(x.PageNo)}
	}
}
