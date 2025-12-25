package utils

import (
	"slices"
)

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
	if len(slice) == 0 {
		return []R{}
	}

	result := make([]R, len(slice))
	for i, item := range slice {
		result[i] = mapper(item)
	}
	return result
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
