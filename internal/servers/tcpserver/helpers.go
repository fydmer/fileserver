package tcpserver

import (
	"strings"
)

func parseNodeHeader(headerStr string) (string, string) {
	sp := strings.SplitN(headerStr, ":", 2)
	if len(sp) != 2 {
		return "", ""
	}
	return sp[0], strings.TrimSpace(sp[1])
}
