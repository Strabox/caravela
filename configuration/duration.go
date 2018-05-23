package configuration

import "time"

/*
Auxiliary data that represents the golang time.Duration
*/
type duration struct {
	time.Duration
}

/*
Implementation of encoding.TextUnmarshal interface for the duration type.
*/
func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
