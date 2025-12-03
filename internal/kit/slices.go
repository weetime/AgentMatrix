package kit

import (
	"strings"

	"github.com/samber/lo"
)

func MapFunc[S ~[]E, E, R any](s S, mapper func(E) R) []R {
	if s == nil {
		return nil
	}
	size := len(s)
	newslice := make([]R, size)
	for i := range s {
		newslice[i] = mapper(s[i])
	}
	return newslice
}

func FilterFunc[S ~[]E, E any](s S, filter func(E) bool) S {
	if s == nil {
		return nil
	}
	size := 0
	for i := range s {
		if filter(s[i]) {
			size += 1
		}
	}

	j := 0
	newslice := make(S, size)
	for i := range s {
		if filter(s[i]) {
			newslice[j] = s[i]
			j += 1
		}
	}
	return newslice
}

func Uniq[S ~[]E, E comparable](s S) S {
	return lo.Uniq(s)
}

func UniqBy[T any, U comparable, Slice ~[]T](collection Slice, iteratee func(item T) U) Slice {
	return lo.UniqBy(collection, iteratee)
}

// ContainsIgnoreCase checks if a string contains a substring (case-insensitive)
func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
