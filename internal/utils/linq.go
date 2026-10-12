package utils

import (
	"cmp"
	"slices"
)

// Query eagerly evaluates each stage. From and ToSlice share their slice's storage;
// Where, Select, SelectMany and OrderByDescending produce new slices.
type Query[T any] []T

func From[T any](slice []T) Query[T] { return Query[T](slice) }

func (q Query[T]) ToSlice() []T { return []T(q) }

func (q Query[T]) Where(predicate func(T) bool) Query[T] {
	return Query[T](Where(q.ToSlice(), predicate))
}

func (q Query[T]) Select[R any](selector func(T) R) Query[R] {
	return Query[R](Select(q.ToSlice(), selector))
}

func (q Query[T]) SelectMany[R any](selector func(T) []R) Query[R] {
	return Query[R](SelectMany(q.ToSlice(), selector))
}

// OrderByDescending sorts a copy, leaving the input slice unchanged.
func (q Query[T]) OrderByDescending[K cmp.Ordered](key func(T) K) Query[T] {
	result := make(Query[T], len(q))
	copy(result, q)
	slices.SortFunc(result, func(a, b T) int {
		return cmp.Compare(key(b), key(a))
	})
	return result
}

// Top preserves the CLI convention that a non-positive limit means no limit.
func (q Query[T]) Top(limit int) Query[T] {
	if limit > 0 && limit < len(q) {
		return q[:limit]
	}
	return q
}

func (q Query[T]) CountBy[K comparable](key func(T) K) map[K]int {
	counts := make(map[K]int)
	for _, item := range q {
		counts[key(item)]++
	}
	return counts
}

// MaxBy and MinBy return false for an empty query and keep the first item on ties.
func (q Query[T]) MaxBy[K cmp.Ordered](key func(T) K) (T, bool) {
	if len(q) == 0 {
		var zero T
		return zero, false
	}
	return MaxBy(q.ToSlice(), func(a, b T) bool {
		return cmp.Compare(key(a), key(b)) < 0
	}), true
}

func (q Query[T]) MinBy[K cmp.Ordered](key func(T) K) (T, bool) {
	if len(q) == 0 {
		var zero T
		return zero, false
	}
	return MinBy(q.ToSlice(), func(a, b T) bool {
		return cmp.Compare(key(a), key(b)) < 0
	}), true
}

func Unique[T comparable](slice []T) []T {
	keys := make(map[T]bool)
	result := []T{}
	for _, item := range slice {
		if _, value := keys[item]; !value {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}

func Contains[T comparable](slice []T, item T) bool {
	return slices.Contains(slice, item)
}

func Intersect[T comparable](slice1, slice2 []T) []T {
	keys := make(map[T]bool)
	result := []T{}

	for _, item := range slice1 {
		keys[item] = true
	}

	for _, item := range slice2 {
		if _, found := keys[item]; found {
			result = append(result, item)
		}
	}

	return result
}

func Except[T comparable](slice1, slice2 []T) []T {
	keys := make(map[T]bool)
	result := []T{}

	for _, item := range slice2 {
		keys[item] = true
	}

	for _, item := range slice1 {
		if _, found := keys[item]; !found {
			result = append(result, item)
		}
	}

	return result
}

func Select[T any, R any](slice []T, selector func(T) R) []R {
	if len(slice) == 0 {
		return []R{}
	}

	result := make([]R, len(slice))
	for i, item := range slice {
		result[i] = selector(item)
	}
	return result
}

func Where[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0)
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

func Any[T any](slice []T, predicate func(T) bool) bool {
	for _, item := range slice {
		if predicate(item) {
			return true
		}
	}
	return false
}

func FirstOrDefault[T any](slice []T, predicate func(T) bool) T {
	var zero T
	for _, item := range slice {
		if predicate(item) {
			return item
		}
	}
	return zero
}

func All[T any](slice []T, predicate func(T) bool) bool {
	for _, item := range slice {
		if !predicate(item) {
			return false
		}
	}
	return true
}

func MapSlice[T any, R any](slice []T, mapper func(T) R) []R {
	return Select(slice, mapper)
}

func MaxBy[T any](slice []T, less func(a, b T) bool) T {
	var zero T
	if len(slice) == 0 {
		return zero
	}
	max := slice[0]
	for _, v := range slice[1:] {
		if less(max, v) {
			max = v
		}
	}
	return max
}

func MinBy[T any](slice []T, less func(a, b T) bool) T {
	var zero T
	if len(slice) == 0 {
		return zero
	}
	min := slice[0]
	for _, v := range slice[1:] {
		if less(v, min) {
			min = v
		}
	}
	return min
}

// SelectMany 展平集合，将每个元素映射为一个集合，然后将所有集合合并为一个平面集合
func SelectMany[T any, R any](slice []T, selector func(T) []R) []R {
	result := make([]R, 0)
	for _, item := range slice {
		result = append(result, selector(item)...)
	}
	return result
}
