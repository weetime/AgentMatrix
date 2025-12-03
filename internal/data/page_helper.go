package data

import (
	"reflect"
	"strings"

	"nova/internal/data/ent"
	"nova/internal/kit"

	"entgo.io/ent/dialect/sql"
)

type pagination[T any] interface {
	Limit(limit int) T
	Offset(offset int) T
}

type paginationOption struct {
	NoUseDefaultOrder bool
	ValidColumns      []string
}

func applyPagination[T pagination[T]](query T, page *kit.PageRequest, validColumns []string) T {
	return applyPaginationWithOptions(query, page, paginationOption{ValidColumns: validColumns})
}

func applyPaginationWithOptions[T pagination[T]](query T, page *kit.PageRequest, options paginationOption) T {
	if page == nil {
		return query
	}
	pageSize := page.GetPageSize()
	sortField := page.GetSortField()

	if pageSize > 0 {
		query.Limit(pageSize)
	}

	if pageNo, ok := page.GetPageNo(); ok {
		offset := (pageNo - 1) * pageSize
		query.Offset(offset)
	}

	var order func(*sql.Selector)
	var orderStr string

	switch page.GetSort() {
	case kit.PageRequest_DESC:
		order = ent.Desc(sortField)
		orderStr = sql.Desc(sortField)
	case kit.PageRequest_ASC:
		order = ent.Asc(sortField)
		orderStr = sql.Asc(sortField)
	}

	noOrder := order == nil
	if len(options.ValidColumns) > 0 && !containsLike(options.ValidColumns, sortField) {
		noOrder = true
	}
	if options.NoUseDefaultOrder && page.SortField == "" {
		noOrder = true
	}
	if !noOrder {
		// = query.Order(order)
		selector, ok := any(query).(*sql.Selector)
		if ok {
			selector.OrderBy(orderStr)
		} else {
			reflect.ValueOf(query).MethodByName("Order").Call([]reflect.Value{reflect.ValueOf(order)})
		}
	}

	return query
}

func containsLike(slice []string, searchTerm string) bool {
	for _, item := range slice {
		// 按空格切分，取最后一个部分
		parts := strings.Fields(item)
		if len(parts) > 0 {
			lastPart := parts[len(parts)-1]
			// 去掉反引号，然后精准匹配
			cleanPart := strings.Trim(lastPart, "`")
			if cleanPart == searchTerm {
				return true
			}
		}
	}
	return false
}
