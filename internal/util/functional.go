package util

func Map[T1 any, T2 any](input []T1, fn func(T1) T2) []T2 {
	output := make([]T2, len(input))
	for i, v := range input {
		output[i] = fn(v)
	}
	return output
}
