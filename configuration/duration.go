package configuration

import (
	"encoding/json"
	"fmt"
	"time"
)

// Auxiliary data that represents the golang time.Duration
type duration struct {
	time.Duration
}

// Implementation of encoding.TextUnmarshal interface for the duration type.
// Used to deserialize durations from the configuration file (TOML).
func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

// Implementation of json.encoding interface for the duration type.
// Used to serialize/deserialize durations during HTTP requests.
func (d *duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// Implementation of json.decoding interface for the duration type.
// Used to serialize/deserialize durations during HTTP requests.
func (d *duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("invalid duration")
	}
}
