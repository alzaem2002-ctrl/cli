package shared

import (
	"regexp"

	"github.com/cli/cli/v2/pkg/cmd/agent-task/capi"
	"github.com/cli/cli/v2/pkg/cmdutil"
)

var uuidRE = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

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
