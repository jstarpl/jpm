package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Status int

const (
	Respawn  Status = 3
	Running  Status = 2
	Starting Status = 1
	Stopped  Status = 0
	Stopping Status = -1
	Failed   Status = -2
)

var (
	Status_name = map[int]string{
		3:  "respawn",
		2:  "running",
		1:  "starting",
		0:  "stopped",
		-1: "stopping",
		-2: "failed",
	}
	Status_value = map[string]int{
		"respawn":  3,
		"running":  2,
		"starting": 1,
		"stopped":  0,
		"stopping": -1,
		"failed":   -2,
	}
)

// String allows Status to implement fmt.Stringer
func (s Status) String() string {
	return Status_name[int(s)]
}

// Convert a string to a Status, returns an error if the string is unknown.
// NOTE: for JSON marshaling this must return a Status value not a pointer, which is
// common when using integer enumerations (or any primitive type alias).
func ParseStatus(s string) (Status, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	value, ok := Status_value[s]
	if !ok {
		return Status(0), fmt.Errorf("%q is not a valid Status", s)
	}
	return Status(value), nil
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Status) UnmarshalJSON(data []byte) (err error) {
	var suits string
	if err := json.Unmarshal(data, &suits); err != nil {
		return err
	}
	if *s, err = ParseStatus(suits); err != nil {
		return err
	}
	return nil
}
