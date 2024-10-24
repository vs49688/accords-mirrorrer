package gqlclient

import (
	"encoding/json"
	"time"
)

const timeLayout = time.RFC3339Nano

type Time struct {
	time.Time
}

func (t Time) MarshalJSON() ([]byte, error) {
	var v interface{}
	if !t.IsZero() {
		v = t.Format(timeLayout)
	}
	return json.Marshal(v)
}

func (t *Time) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		t.Time = time.Time{}
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	var err error
	t.Time, err = time.Parse(timeLayout, s)
	return err
}
