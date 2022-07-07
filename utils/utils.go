package utils

func Bool(i bool) *bool {
	return &i
}

func Int(i int) *int {
	return &i
}

func Int64(i int64) *int64 {
	return &i
}

func String(i string) *string {
	return &i
}

func Strings(ss []string) []*string {
	r := make([]*string, len(ss))
	for i := range ss {
		r[i] = &ss[i]
	}
	return r
}
