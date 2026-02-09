package utils

import "strings"

func StringContainsCaseInsensitive(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
