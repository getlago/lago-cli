package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
)

var decimalPattern = regexp.MustCompile(`^-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?$`)

type Decimal string

func NewDecimal(value string) (Decimal, error) {
	if !decimalPattern.MatchString(value) {
		return "", fmt.Errorf("invalid decimal %q", value)
	}
	return Decimal(value), nil
}

func (d Decimal) MarshalJSON() ([]byte, error) {
	if !decimalPattern.MatchString(string(d)) {
		return nil, fmt.Errorf("invalid decimal %q", d)
	}
	return []byte(d), nil
}

func (d *Decimal) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		return nil
	}
	if data[0] == '"' {
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		data = []byte(value)
	}
	value, err := NewDecimal(string(data))
	if err != nil {
		return err
	}
	*d = value
	return nil
}

func (d Decimal) String() string { return string(d) }

type MinorUnits int64
