package httpserver

import (
	"strings"
)

func parseContentDisposition(values string) string {
	sp := strings.Split(values, ";")
	for _, val := range sp {
		cleanVal := strings.TrimSpace(val)
		if strings.HasPrefix(cleanVal, "filename=") {
			return strings.Trim(strings.TrimPrefix(cleanVal, "filename="), "\"")
		}
	}
	return ""
}
