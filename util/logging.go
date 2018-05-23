package util

import "strings"

const maxTagSize = 12

/*
Builds a tag for the log statments with identation in order to align the log.
*/
func LogTag(tag string) string {
	tagLen := len(tag)
	indentation := strings.Repeat(" ", maxTagSize-tagLen)
	return tag + indentation
}
