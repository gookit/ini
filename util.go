package ini

// simple merge two string map
func mergeStringMap(src, dst map[string]string) map[string]string {
	for k,v := range src {
		dst[k] = v
	}

	return dst
}
