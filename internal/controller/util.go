package controller

import "strings"

func concat(s ...string) string {
	return strings.Join(s, "")
}
