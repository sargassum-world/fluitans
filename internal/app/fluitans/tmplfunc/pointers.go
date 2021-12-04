package tmplfunc

func derefBool(b *bool) bool {
	return b != nil && *b
}

func derefInt(i *int, nilValue int) int {
	if i == nil {
		return nilValue
	}

	return *i
}

func derefFloat32(i *float32, nilValue float32) float32 {
	if i == nil {
		return nilValue
	}

	return *i
}

func derefString(s *string, nilValue string) string {
	if s == nil {
		return nilValue
	}

	return *s
}
