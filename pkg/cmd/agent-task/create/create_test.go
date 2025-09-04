package create

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/pkg/cmd/agent-task/capi"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/google/shlex"
	"github.com/stretchr/testify/require"
)

func TestNewCmdCreate(t *testing.T) {
	tmpDir := t.TempDir()

	tmpEmptyFile := filepath.Join(tmpDir, "empty-task-description.md")
	err := os.WriteFile(tmpEmptyFile, []byte("  \n\n"), 0600)
	require.NoError(t, err)

	tmpFile := filepath.Join(tmpDir, "task-description.md")
	err = os.WriteFile(tmpFile, []byte("task description from file"), 0600)
	require.NoError(t, err)

	tests := []struct {
		name     string
		args     string
		stdin    string
		wantOpts *CreateOptions // nil when expecting error
		wantErr  string
	}{
		{
			name:    "no args nor file",
			wantErr: "a task description is required",
		},
		{
			name: "arg only success",
			args: "'task description from args'",
			wantOpts: &CreateOptions{
				ProblemStatement: "task description from args",
			},
		},
		{
			name: "from-file success",
			args: fmt.Sprintf("-F %s", tmpFile),
			wantOpts: &CreateOptions{
				ProblemStatement: "task description from file",
			},
		},
		{
			name:  "file content from stdin success",
			args:  "-F -",
			stdin: "task description from stdin",
			wantOpts: &CreateOptions{
				ProblemStatement: "task description from stdin",
			},
		},
		{
			name:    "mutually exclusive arg and file",
			args:    "'some task inline' -F foo.md",
			wantErr: "only one of -F or arg can be provided",
		},
		{
			name:    "missing file path",
			args:    "-F does-not-exist.md",
			wantErr: "could not read task description file: open does-not-exist.md:",
		},
		{
			name:    "empty file",
			args:    fmt.Sprintf("-F %s", tmpEmptyFile),
			wantErr: "task description file is empty",
		},
		{
			name:    "empty from stdin",
			args:    "-F -",
			stdin:   "   \n\n",
			wantErr: "task description file is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, stdin, _, _ := iostreams.Test()
			f := &cmdutil.Factory{
				IOStreams: ios,
			}

			var gotOpts *CreateOptions
			cmd := NewCmdCreate(f, func(o *CreateOptions) error {
				gotOpts = o
				return nil
			})

			argv, err := shlex.Split(tt.args)
			require.NoError(t, err)
			cmd.SetArgs(argv)

			cmd.SetIn(stdin)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			if tt.stdin != "" {
				stdin.WriteString(tt.stdin)
			}

			_, err = cmd.ExecuteC()

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
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
	sampleDateString := "2025-08-29T00:00:00Z"
	sampleDate, err := time.Parse(time.RFC3339, sampleDateString)
	require.NoError(t, err)

	createdJobSuccess := capi.Job{
		ID:        "job123",
		SessionID: "sess1",
		Actor: &capi.JobActor{
			ID:    1,
			Login: "octocat",
		},
		CreatedAt: sampleDate,
		UpdatedAt: sampleDate,
	}
	createdJobSuccessWithPR := capi.Job{
		ID:        "job123",
		SessionID: "sess1",
		Actor: &capi.JobActor{
			ID:    1,
			Login: "octocat",
		},
		CreatedAt: sampleDate,
		UpdatedAt: sampleDate,
		PullRequest: &capi.JobPullRequest{
			ID:     101,
			Number: 42,
		},
	}

	tests := []struct {
		name         string
		capiStubs    func(*testing.T, *capi.CapiClientMock)
		baseRepoFunc func() (ghrepo.Interface, error)
		baseBranch   string
		wantStdout   string
		wantStdErr   string
		wantErr      string
	}{
		{
			name:         "missing repo returns error",
			baseRepoFunc: func() (ghrepo.Interface, error) { return nil, nil },
			wantErr:      "a repository is required; re-run in a repository or supply one with --repo owner/name",
		},
		{
			name:         "base branch included in create payload",
			baseRepoFunc: func() (ghrepo.Interface, error) { return ghrepo.New("OWNER", "REPO"), nil },
			baseBranch:   "feature",
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.CreateJobFunc = func(ctx context.Context, owner, repo, problemStatement, baseBranch string) (*capi.Job, error) {
					require.Equal(t, "OWNER", owner)
					require.Equal(t, "REPO", repo)
					require.Equal(t, "Do the thing", problemStatement)
					require.Equal(t, "feature", baseBranch)
					return &createdJobSuccess, nil
				}
				m.GetJobFunc = func(ctx context.Context, owner, repo, jobID string) (*capi.Job, error) {
					require.Equal(t, "OWNER", owner)
					require.Equal(t, "REPO", repo)
					require.Equal(t, "job123", jobID)
					return &createdJobSuccessWithPR, nil
				}
			},
			wantStdout: "https://github.com/OWNER/REPO/pull/42/agent-sessions/sess1\n",
		},
		{
			name:         "create task API failure returns error",
			baseRepoFunc: func() (ghrepo.Interface, error) { return ghrepo.New("OWNER", "REPO"), nil },
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.CreateJobFunc = func(ctx context.Context, owner, repo, problemStatement, baseBranch string) (*capi.Job, error) {
					require.Equal(t, "OWNER", owner)
					require.Equal(t, "REPO", repo)
					require.Equal(t, "Do the thing", problemStatement)
					require.Equal(t, "", baseBranch)
					return nil, errors.New("some error")
				}
			},
			wantErr: "some error",
		},
		{
			name:         "get job API failure surfaces error",
			baseRepoFunc: func() (ghrepo.Interface, error) { return ghrepo.New("OWNER", "REPO"), nil },
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.CreateJobFunc = func(ctx context.Context, owner, repo, problemStatement, baseBranch string) (*capi.Job, error) {
					require.Equal(t, "OWNER", owner)
					require.Equal(t, "REPO", repo)
					require.Equal(t, "Do the thing", problemStatement)
					require.Equal(t, "", baseBranch)
					return &createdJobSuccess, nil
				}
				m.GetJobFunc = func(ctx context.Context, owner, repo, jobID string) (*capi.Job, error) {
					return nil, errors.New("some error")
				}
			},
			wantStdErr: "some error\n",
			wantStdout: "job job123 queued. View progress: https://github.com/copilot/agents\n",
		},
		{
			name:         "success with immediate PR",
			baseRepoFunc: func() (ghrepo.Interface, error) { return ghrepo.New("OWNER", "REPO"), nil },
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.CreateJobFunc = func(ctx context.Context, owner, repo, problemStatement, baseBranch string) (*capi.Job, error) {
					require.Equal(t, "OWNER", owner)
					require.Equal(t, "REPO", repo)
					require.Equal(t, "Do the thing", problemStatement)
					require.Equal(t, "", baseBranch)
					return &createdJobSuccessWithPR, nil
				}
			},
			wantStdout: "https://github.com/OWNER/REPO/pull/42/agent-sessions/sess1\n",
		},
		{
			name:         "success with delayed PR after polling",
			baseRepoFunc: func() (ghrepo.Interface, error) { return ghrepo.New("OWNER", "REPO"), nil },
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.CreateJobFunc = func(ctx context.Context, owner, repo, problemStatement, baseBranch string) (*capi.Job, error) {
					require.Equal(t, "OWNER", owner)
					require.Equal(t, "REPO", repo)
					require.Equal(t, "Do the thing", problemStatement)
					require.Equal(t, "", baseBranch)
					return &createdJobSuccess, nil
				}
				m.GetJobFunc = func(ctx context.Context, owner, repo, jobID string) (*capi.Job, error) {
					require.Equal(t, "OWNER", owner)
					require.Equal(t, "REPO", repo)
					require.Equal(t, "job123", jobID)
					return &createdJobSuccessWithPR, nil
				}
			},
			wantStdout: "https://github.com/OWNER/REPO/pull/42/agent-sessions/sess1\n",
		},
		{
			name:         "fallback after timeout returns link to global agents page",
			baseRepoFunc: func() (ghrepo.Interface, error) { return ghrepo.New("OWNER", "REPO"), nil },
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.CreateJobFunc = func(ctx context.Context, owner, repo, problemStatement, baseBranch string) (*capi.Job, error) {
					require.Equal(t, "OWNER", owner)
					require.Equal(t, "REPO", repo)
					require.Equal(t, "Do the thing", problemStatement)
					require.Equal(t, "", baseBranch)
					return &createdJobSuccess, nil
				}

				count := 0
				m.GetJobFunc = func(ctx context.Context, owner, repo, jobID string) (*capi.Job, error) {
					if count++; count > 4 {
						require.FailNow(t, "too many get calls")
					}
					return &createdJobSuccess, nil
				}
			},
			wantStdout: "job job123 queued. View progress: https://github.com/copilot/agents\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capiClientMock := &capi.CapiClientMock{}
			if tt.capiStubs != nil {
				tt.capiStubs(t, capiClientMock)
			}

			ios, _, stdout, stderr := iostreams.Test()
			opts := &CreateOptions{
				IO:               ios,
				ProblemStatement: "Do the thing",
				BaseRepo:         tt.baseRepoFunc,
				BaseBranch:       tt.baseBranch,
				CapiClient: func() (capi.CapiClient, error) {
					return capiClientMock, nil
				},
			}

			// A backoff with no internal between retries to keep tests fast,
			// and also a max number of retries so we don't infinitely poll.
			opts.BackOff = backoff.WithMaxRetries(&backoff.ZeroBackOff{}, 3)

			err := createRun(opts)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.wantStdout, stdout.String())
			require.Equal(t, tt.wantStdErr, stderr.String())
		})
	}
}
