package main

import (
	"strings"
)

func cleantagstring(inp string) string {
	tmp := strings.TrimSuffix(inp, ",")
	tmp = strings.Replace(tmp, "/", ".", -1)
	return tmp
}

func min(a, b uint32) uint32 {
	if a < b {
		return a
	} else {
		return b
	}
}
