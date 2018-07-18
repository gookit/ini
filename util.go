package ini

import "strings"

// simple merge two string map
func mergeStringMap(src, dst map[string]string, ignoreCase bool) map[string]string {
	for k,v := range src {
		if ignoreCase {
			k = strings.ToLower(k)
		}

		dst[k] = v
	}

	return dst
}

func mapKeyToLower(src map[string]string) map[string]string {
	newMp := make(map[string]string)

	for k,v := range src {
		k = strings.ToLower(k)
		newMp[k] = v
	}

	return newMp
}
