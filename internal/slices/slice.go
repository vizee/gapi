package slices

func Shrink[T any](a []T) []T {
	if cap(a) > len(a) {
		a = append(make([]T, 0, len(a)), a...)
	}
	return a
}
