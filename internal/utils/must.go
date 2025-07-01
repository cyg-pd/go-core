package utils

func Must0(err error) {
	if err != nil {
		panic(err)
	}
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
