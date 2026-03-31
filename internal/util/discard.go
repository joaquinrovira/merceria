package util

func DiscardMap[T1, T2 any](fn func(T1) (T2, error)) func(T1) T2 {
	return func(t T1) T2 {
		v, _ := fn(t)
		return v
	}
}
