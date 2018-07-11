package util

import "strings"

// Max size for a log's entry tag
const maxTagSize = 12

// Builds a tag for the log entries with indentation in order to align the log entries.
func LogTag(tag string) string {
	tag = "[" + tag + "]"
	sizeOfIndentation := maxTagSize - len(tag)
	if sizeOfIndentation <= 0 {
		tag = tag[:maxTagSize-1]
		sizeOfIndentation = maxTagSize - len(tag)
	}
	indentation := strings.Repeat(" ", sizeOfIndentation)
	return tag + indentation
}
