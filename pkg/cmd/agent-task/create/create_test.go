package create

import (
	"net/http"
	"os"
	"path/filepath"
	"slices"
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
	type tc struct {
		name        string
		args        []string
		fileContent string         // if non-empty, create temp file and substitute {{FILE}} token in args
		wantOpts    *CreateOptions // nil when expecting error
		expectedErr string
	}

	tests := []tc{
		{
			name:        "no args nor file",
			args:        []string{},
			expectedErr: "a task description is required",
		},
		{
			name: "arg only success",
			args: []string{"task description from args"},
			wantOpts: &CreateOptions{
				ProblemStatement: "task description from args",
			},
		},
		{
			name:        "from-file success",
			args:        []string{"-F", "{{FILE}}"},
			fileContent: "task description from file",
			wantOpts: &CreateOptions{
				ProblemStatement: "task description from file",
			},
		},
		{
			name:        "file content from stdin success",
			args:        []string{"-F", "-"},
			fileContent: "task from stdin",
			wantOpts:    &CreateOptions{ProblemStatement: "task from stdin"},
		},
		{
			name:        "mutually exclusive arg and file",
			args:        []string{"Some task inline", "-F", "{{FILE}}"},
			fileContent: "Some task",
			expectedErr: "only one of -F or arg can be provided",
		},
		{
			name:        "missing file path",
			args:        []string{"-F", "does-not-exist.md"},
			expectedErr: "could not read task description file: open does-not-exist.md: no such file or directory",
		},
		{
			name:        "empty file",
			args:        []string{"-F", "{{FILE}}"},
			fileContent: "   \n\n",
			expectedErr: "task description file is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, stdinBuf, _, _ := iostreams.Test()

			// Provide file content either via stdin ( -F - ) or by creating a temp file
			if tt.fileContent != "" {
				isStdin := len(tt.args) == 2 && tt.args[0] == "-F" && tt.args[1] == "-"
				hasFileToken := slices.Contains(tt.args, "{{FILE}}")

				switch {
				case isStdin:
					stdinBuf.WriteString(tt.fileContent)
				case hasFileToken:
					dir := t.TempDir()
					path := filepath.Join(dir, "task.md")
					if err := os.WriteFile(path, []byte(tt.fileContent), 0o600); err != nil {
						t.Fatalf("failed to write temp file: %v", err)
					}
					for i, a := range tt.args {
						if a == "{{FILE}}" {
							tt.args[i] = path
						}
					}
				}
			}

			f := &cmdutil.Factory{IOStreams: ios}
			var gotOpts *CreateOptions
			cmd := NewCmdCreate(f, func(o *CreateOptions) error {
				gotOpts = o
				return nil
			})
			cmd.SetArgs(tt.args)
			_, err := cmd.ExecuteC()

			if tt.expectedErr != "" {
				require.Error(t, err)
				require.Equal(t, tt.expectedErr, err.Error())
				return
			}
			require.NoError(t, err)
			if tt.wantOpts != nil {
				require.Equal(t, tt.wantOpts.ProblemStatement, gotOpts.ProblemStatement)
			}
		})
	}
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
		baseRepoFunc     func() (ghrepo.Interface, error)
		problemStatement string
		wantStdout       string
		wantStdErr       string
		wantErr          string
	}{
		{
			name:             "get job API failure surfaces error",
			baseRepoFunc:     func() (ghrepo.Interface, error) { return ghrepo.New("OWNER", "REPO"), nil },
			problemStatement: "Do the thing",
			stubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(httpmock.REST("POST", "agents/swe/v1/jobs/OWNER/REPO"), "api.githubcopilot.com"),
					httpmock.StatusStringResponse(201, createdJobTimeoutResponse),
				)
				reg.Register(
					httpmock.WithHost(httpmock.REST("GET", "agents/swe/v1/jobs/OWNER/REPO/jobABC"), "api.githubcopilot.com"),
					httpmock.StatusStringResponse(500, `{"error":{"message":"internal server error"}}`),
				)
			},
			wantStdErr: "failed to get job: 500 Internal Server Error\n",
			wantStdout: "job jobABC queued. View progress: https://github.com/copilot/agents\n",
		},
		{
			name:             "success with immediate PR",
			baseRepoFunc:     func() (ghrepo.Interface, error) { return ghrepo.New("OWNER", "REPO"), nil },
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
			baseRepoFunc:     func() (ghrepo.Interface, error) { return ghrepo.New("OWNER", "REPO"), nil },
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
			baseRepoFunc:     func() (ghrepo.Interface, error) { return ghrepo.New("OWNER", "REPO"), nil },
			problemStatement: "Do the thing",
			stubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(httpmock.REST("POST", "agents/swe/v1/jobs/OWNER/REPO"), "api.githubcopilot.com"),
					httpmock.StatusStringResponse(201, createdJobTimeoutResponse),
				)
				// 4 attempts: initial + 3 retries
				for range 4 {
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
			baseRepoFunc:     func() (ghrepo.Interface, error) { return nil, nil },
			wantErr:          "a repository is required; re-run in a repository or supply one with --repo owner/name",
		},
		{
			name:             "create task API failure returns error",
			baseRepoFunc:     func() (ghrepo.Interface, error) { return ghrepo.New("OWNER", "REPO"), nil },
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
			name:             "missing task description returns error",
			baseRepoFunc:     func() (ghrepo.Interface, error) { return ghrepo.New("OWNER", "REPO"), nil },
			problemStatement: "",
			wantErr:          "a task description is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, stdout, stderr := iostreams.Test()
			opts := &CreateOptions{
				IO:               ios,
				ProblemStatement: tt.problemStatement,
				BaseRepo:         tt.baseRepoFunc,
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
				require.Equal(t, tt.wantErr, err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.wantStdout, stdout.String())
			require.Equal(t, tt.wantStdErr, stderr.String())

			if tt.stubs != nil {
				reg.Verify(t)
			}
		})
	}
}
