package utils

func CheckNotNil(v interface{}) {
	if v == nil {
		panic("not nil")
	}
}

func CheckNoError(err error) {
	if err != nil {
		panic(err)
	}
}
