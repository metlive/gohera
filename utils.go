package gohera

// 三目运算的函数
func Ternary[T any](a bool, b, c T) T {
	if a {
		return b
	}
	return c
}

// 查找包含
func Contains[T comparable](needle T, hystack []T) bool {
	for _, item := range hystack {
		if needle == item {
			return true
		}
	}
	return false
}

func Map[K comparable, V any](needle K, hystack map[K]V) bool {
	if _, ok := hystack[needle]; ok {
		return true
	}
	return false
}

func Set[T comparable](array []T) []T {
	result := make([]T, 0)
	temp := map[T]struct{}{}
	for _, item := range array {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
