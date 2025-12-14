package gohera

import "slices"

// Ternary 三目运算通用函数
// 如果条件 a 为 true，返回 b，否则返回 c
func Ternary[T any](a bool, b, c T) T {
	if a {
		return b
	}
	return c
}

// Contains 判断切片中是否包含指定元素
func Contains[T comparable](needle T, hystack []T) bool {
	return slices.Contains(hystack, needle)
}

// Map 判断 Map 中是否存在指定的 Key
func Map[K comparable, V any](needle K, hystack map[K]V) bool {
	if _, ok := hystack[needle]; ok {
		return true
	}
	return false
}

// Set 对切片进行去重
func Set[T comparable](array []T) []T {
	result := make([]T, 0, len(array))
	temp := map[T]struct{}{}
	for _, item := range array {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
