package shared

import (
	"errors"
	"regexp"

	"github.com/cli/cli/v2/pkg/cmd/agent-task/capi"
	prShared "github.com/cli/cli/v2/pkg/cmd/pr/shared"
	"github.com/cli/cli/v2/pkg/cmdutil"
)

var uuidRE = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
var agentSessionsPathRE = regexp.MustCompile(`^/agent-sessions/([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12})$`)

func CapiClientFunc(f *cmdutil.Factory) func() (capi.CapiClient, error) {
	return func() (capi.CapiClient, error) {
		cfg, err := f.Config()
		if err != nil {
			return nil, err
		}

		httpClient, err := f.HttpClient()
		if err != nil {
			return nil, err
		}

		authCfg := cfg.Authentication()
		return capi.NewCAPIClient(httpClient, authCfg), nil
	}
}

func IsSessionID(s string) bool {
	return uuidRE.MatchString(s)
}

// ParsePullRequestAgentSessionURL parses session ID from a pull request's agent
// session URL, which is of the form:
//
//	https://github.com/OWNER/REPO/pull/NUMBER/agent-sessions/SESSION-ID
func ParsePullRequestAgentSessionURL(u string) (string, error) {
	_, _, rest, err := prShared.ParseURL(u)
	if err != nil {
		return "", err
	}

	match := agentSessionsPathRE.FindStringSubmatch(rest)
	if match == nil {
		return "", errors.New("not a valid agent session URL")
	}
	return match[1], nil
}
