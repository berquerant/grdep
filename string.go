package grdep

import (
	"slices"
	"strings"
)

func Split(s, sep string) []string {
	return slices.DeleteFunc(strings.Split(s, sep), func(x string) bool {
		return strings.TrimSpace(x) == ""
	})
}
