package device

import "strings"

func stripNull(s string) string {
	return strings.Replace(s, string(0), ``, -1)
}
