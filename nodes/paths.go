package nodes

import "strings"

func PathName(path string) string {
	if len(path) == 0 {
		return ""
	}
	i := strings.LastIndex(path, "/")
	for i == len(path) - 1 && i != -1 {
		path = path[:i]
		i = strings.LastIndex(path, "/")
	}
	if i == -1 {
		return path
	}

	return path[i + 1:]
}