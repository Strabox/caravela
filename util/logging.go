package util

import "strings"

const maxTagSize = 12

func LogTag(tag string) string {
	tagLen := len(tag)
	indentation := strings.Repeat(" ", maxTagSize-tagLen)
	return tag + indentation
}
