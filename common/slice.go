package common

func NotNullSlice[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
