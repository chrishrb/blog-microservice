package api

func getPaginationWithDefaults(paramOffset, paramLimit *int) (int, int) {
	offset := 0
	limit := 20

	if paramOffset != nil {
		offset = *paramOffset
	}
	if paramLimit != nil {
		limit = *paramLimit
	}
	if limit > 100 {
		limit = 100
	}

	return offset, limit
}
