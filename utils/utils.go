package utils

func Bool(i bool) *bool {
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
