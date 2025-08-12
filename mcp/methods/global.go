package methods

func MethodIn(m map[string]struct{}, method string) bool {
	_, ok := m[method]
	return ok
}
