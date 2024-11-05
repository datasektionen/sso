package templates

//go:generate templ generate

func bigIfTrue[T any](cond bool, thenVal, elseVal T) T {
	if cond {
		return thenVal
	} else {
		return elseVal
	}
}
