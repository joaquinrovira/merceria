package util

import "fmt"

func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func MustOk[T any](value T, ok bool) T {
	if !ok {
		panic(fmt.Errorf("MustOk(): fail"))
	}
	return value
}
