package util

import (
	"strconv"
	"strings"
)

func UnQuoteJson(s string) (string, error) {
	s = strings.TrimSpace(s)
	if len(s) > 1 {
		quote := s[0]
		switch quote {
		case '{':
			return s, nil
		default:
			return strconv.Unquote(s)
		}
	}
	return s, nil
}
