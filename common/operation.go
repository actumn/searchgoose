package common

func GetIntersection(first, second []string) (result []string) {
	m := make(map[string]bool)
	for _, item := range first {
		m[item] = true
	}

	for _, item := range second {
		if _, ok := m[item]; ok {
			result = append(result, item)
		}
	}
	return
}

func GetMaxInt(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}
