package shared

import (
	"github.com/cli/cli/v2/pkg/cmd/agent-task/capi"
	"github.com/cli/cli/v2/pkg/iostreams"
)

// ColorFuncForSessionState returns a function that colors the session state
func ColorFuncForSessionState(s capi.Session, cs *iostreams.ColorScheme) func(string) string {
	var stateColor func(string) string
	switch s.State {
	case "completed":
		stateColor = cs.Green
	case "canceled":
		stateColor = cs.Muted
	case "in_progress", "queued":
		stateColor = cs.Yellow
	case "failed":
		stateColor = cs.Red
	default:
		stateColor = cs.Muted
	}

	return stateColor
}
