package status

import "encoding/json"

type authState int

const (
	authStateSuccess authState = iota
	authStateTimeout
	authStateError
)

func (s authState) String() string {
	switch s {
	case authStateSuccess:
		return "success"
	case authStateTimeout:
		return "timeout"
	case authStateError:
		return "error"
	default:
		return "unknown"
	}
}

func (s authState) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
