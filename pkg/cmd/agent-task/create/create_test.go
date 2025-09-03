package create

import (
	"net/http"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/cenkalti/backoff/v4"
	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/pkg/cmd/agent-task/capi"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

// Test basic option parsing & repository requirement
func TestNewCmdCreate_Args(t *testing.T) {
	f := &cmdutil.Factory{}
	cmd := NewCmdCreate(f, func(o *CreateOptions) error { return nil })
	// no args should error via cobra MinimumNArgs before our runF
	// TODO once we support more sources of problem statement input,
	// this will change.
	_, err := cmd.ExecuteC()
	require.Error(t, err)
}

func Test_createRun(t *testing.T) {
	createdJobSuccessResponse := heredoc.Doc(`{
		"job_id":"job123",
		"session_id":"sess1",
		"actor":{"id":1,"login":"octocat"},
		"created_at":"2025-08-29T00:00:00Z",
		"updated_at":"2025-08-29T00:00:00Z"
	}`)
	createdJobSuccessWithPRResponse := heredoc.Doc(`{
		"job_id":"job123",
		"session_id":"sess1",
		"actor":{"id":1,"login":"octocat"},
		"created_at":"2025-08-29T00:00:00Z",
		"updated_at":"2025-08-29T00:00:00Z",
		"pull_request":{"id":101,"number":42}
	}`)
	createdJobTimeoutResponse := heredoc.Doc(`{
		"job_id":"jobABC",
		"session_id":"sess1",
		"actor":{"id":1,"login":"octocat"},
		"created_at":"2025-08-29T00:00:00Z",
		"updated_at":"2025-08-29T00:00:00Z"
	}`)

	tests := []struct {
		name             string
		stubs            func(*httpmock.Registry)
		baseRepo         ghrepo.Interface
		baseRepoErr      error
		problemStatement string
		wantStdout       string
		wantErr          string
	}{
		{
			name:             "success with immediate PR",
			baseRepo:         ghrepo.New("OWNER", "REPO"),
			problemStatement: "Do the thing",
			stubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(httpmock.REST("POST", "agents/swe/v1/jobs/OWNER/REPO"), "api.githubcopilot.com"),
					httpmock.StatusStringResponse(201, createdJobSuccessWithPRResponse),
				)
			},
			wantStdout: "https://github.com/OWNER/REPO/pull/42/agent-sessions/sess1\n",
		},
		{
			name:             "success with delayed PR after polling",
			baseRepo:         ghrepo.New("OWNER", "REPO"),
			problemStatement: "Do the thing",
			stubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(httpmock.REST("POST", "agents/swe/v1/jobs/OWNER/REPO"), "api.githubcopilot.com"),
					httpmock.StatusStringResponse(201, createdJobSuccessResponse),
				)
				reg.Register(
					httpmock.WithHost(httpmock.REST("GET", "agents/swe/v1/jobs/OWNER/REPO/job123"), "api.githubcopilot.com"),
					httpmock.StringResponse(`{"job_id":"job123","pull_request":{"id":101,"number":42}}`),
				)
			},
			wantStdout: "https://github.com/OWNER/REPO/pull/42\n",
		},
		{
			name:             "fallback after timeout returns link to global agents page",
			baseRepo:         ghrepo.New("OWNER", "REPO"),
			problemStatement: "Do the thing",
			stubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(httpmock.REST("POST", "agents/swe/v1/jobs/OWNER/REPO"), "api.githubcopilot.com"),
					httpmock.StatusStringResponse(201, createdJobTimeoutResponse),
				)
				for range 3 {
					reg.Register(
						httpmock.WithHost(httpmock.REST("GET", "agents/swe/v1/jobs/OWNER/REPO/jobABC"), "api.githubcopilot.com"),
						httpmock.StringResponse(`{"job_id":"jobABC"}`),
					)
				}
			},
			wantStdout: "job jobABC queued. View progress: https://github.com/copilot/agents\n",
		},
		{
			name:             "missing repo returns error",
			problemStatement: "task",
			baseRepo:         ghrepo.New("", ""),
			wantErr:          "a repository is required; re-run in a repository or supply one with --repo owner/name",
		},
		{
			name:             "create task API failure returns error",
			baseRepo:         ghrepo.New("OWNER", "REPO"),
			problemStatement: "do the thing",
			stubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(httpmock.REST("POST", "agents/swe/v1/jobs/OWNER/REPO"), "api.githubcopilot.com"),
					httpmock.StatusStringResponse(500, `{"error":{"message":"some API error"}}`),
				)
			},
			wantErr: "failed to create job: some API error",
		},
		{
			name:             "error fetching job during polling returns error and falls back to global agents page",
			baseRepo:         ghrepo.New("OWNER", "REPO"),
			problemStatement: "Do the thing",
			stubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(httpmock.REST("POST", "agents/swe/v1/jobs/OWNER/REPO"), "api.githubcopilot.com"),
					httpmock.StatusStringResponse(201, createdJobTimeoutResponse),
				)
				reg.Register(
					httpmock.WithHost(httpmock.REST("GET", "agents/swe/v1/jobs/OWNER/REPO/jobABC"), "api.githubcopilot.com"),
					httpmock.StringResponse(`{"job_id":"jobABC"}`),
				)
				reg.Register(
					httpmock.WithHost(httpmock.REST("GET", "agents/swe/v1/jobs/OWNER/REPO/jobABC"), "api.githubcopilot.com"),
					httpmock.StatusStringResponse(500, `{"error":{"message":"something went wrong"}}`),
				)
			},
			wantStdout: "job jobABC queued. View progress: https://github.com/copilot/agents\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, stdout, _ := iostreams.Test()
			opts := &CreateOptions{
				IO:               ios,
				ProblemStatement: tt.problemStatement,
			}

			if tt.baseRepo != nil || tt.baseRepoErr != nil {
				br, bre := tt.baseRepo, tt.baseRepoErr
				opts.BaseRepo = func() (ghrepo.Interface, error) { return br, bre }
			}

			// A backoff with no internal between retries to keep tests fast,
			// and also a max number of retries so we don't infinitely poll.
			opts.BackOff = backoff.WithMaxRetries(&backoff.ZeroBackOff{}, 3)

			reg := &httpmock.Registry{}
			if tt.stubs != nil {
				tt.stubs(reg)
				cfg := config.NewBlankConfig()
				cfg.Set("github.com", "oauth_token", "OTOKEN")
				authCfg := cfg.Authentication()
				client := capi.NewCAPIClient(&http.Client{Transport: reg}, authCfg)
				opts.CapiClient = func() (capi.CapiClient, error) { return client, nil }
			}

			err := createRun(opts)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.Equal(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			if tt.wantStdout != "" {
				require.Equal(t, tt.wantStdout, stdout.String())
			}
			if tt.stubs != nil {
				reg.Verify(t)
			}
		})
	}
}
