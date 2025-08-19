package status

import "encoding/json"

type AuthState int

const (
	AuthStateSuccess AuthState = iota
	AuthStateTimeout
	AuthStateError
)

func (s AuthState) String() string {
	switch s {
	case AuthStateSuccess:
		return "success"
	case AuthStateTimeout:
		return "timeout"
	case AuthStateError:
		return "error"
	default:
		return "unknown"
	}
}

func (s AuthState) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
