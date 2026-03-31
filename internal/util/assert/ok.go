package assert

func Ok[T any](value T, ok bool) T {
	if !ok {
		panic("assertion failed")
	}
	return value
}
