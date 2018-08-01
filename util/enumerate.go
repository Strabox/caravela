package util

// Enum is the interface for all the custom enum types.
type Enum interface {
	name() string
	ordinal() int
	values() *[]string
}
