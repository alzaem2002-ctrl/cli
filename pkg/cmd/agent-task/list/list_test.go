package list

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/browser"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/pkg/cmd/agent-task/capi"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdList(t *testing.T) {
	tests := []struct {
		name         string
		args         string
		wantOpts     ListOptions
		wantBaseRepo ghrepo.Interface
		wantErr      string
	}{
		{
			name: "no arguments",
			wantOpts: ListOptions{
				Limit: defaultLimit,
			},
		},
		{
			name: "base repo specified",
			args: "--repo OWNER/REPO",
			wantOpts: ListOptions{
				Limit: defaultLimit,
			},
			wantBaseRepo: ghrepo.New("OWNER", "REPO"),
		},
		{
			name: "custom limit",
			args: "--limit 15",
			wantOpts: ListOptions{
				Limit: 15,
			},
		},
		{
			name:    "invalid limit",
			args:    "--limit 0",
			wantErr: "invalid limit: 0",
		},
		{
			name:    "negative limit",
			args:    "--limit -5",
			wantErr: "invalid limit: -5",
		},
		{
			name: "web flag",
			args: "--web",
			wantOpts: ListOptions{
				Limit: defaultLimit,
				Web:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			f := &cmdutil.Factory{
				IOStreams: ios,
			}

			var gotOpts *ListOptions
			cmd := NewCmdList(f, func(opts *ListOptions) error { gotOpts = opts; return nil })

			argv, err := shlex.Split(tt.args)
			require.NoError(t, err)
			cmd.SetArgs(argv)

			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			_, err = cmd.ExecuteC()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantOpts.Limit, gotOpts.Limit)
			assert.Equal(t, tt.wantOpts.Web, gotOpts.Web)

			if tt.wantBaseRepo != nil {
				baseRepo, err := gotOpts.BaseRepo()
				require.NoError(t, err)
				assert.True(t, ghrepo.IsSame(tt.wantBaseRepo, baseRepo))
			}
		})
	}
}

func Test_listRun(t *testing.T) {
	sampleDate := time.Now().Add(-6 * time.Hour) // 6h ago
	sampleDateString := sampleDate.Format(time.RFC3339)

	tests := []struct {
		name           string
		tty            bool
		stubs          func(*httpmock.Registry)
		capiStubs      func(*testing.T, *capi.CapiClientMock)
		baseRepo       ghrepo.Interface
		baseRepoErr    error
		limit          int
		web            bool
		wantOut        string
		wantErr        error
		wantStderr     string
		wantBrowserURL string
	}{
		{
			name: "viewer-scoped no sessions",
			tty:  true,
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.ListSessionsForViewerFunc = func(ctx context.Context, limit int) ([]*capi.Session, error) {
					return nil, nil
				}
			},
			wantErr: cmdutil.NewNoResultsError("no agent tasks found"),
		},
		{
			name:  "viewer-scoped respects --limit",
			tty:   true,
			limit: 999,
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.ListSessionsForViewerFunc = func(ctx context.Context, limit int) ([]*capi.Session, error) {
					assert.Equal(t, 999, limit)
					return nil, nil
				}
			},
			wantErr: cmdutil.NewNoResultsError("no agent tasks found"), // not important
		},
		{
			name: "viewer-scoped single session (tty)",
			tty:  true,
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.ListSessionsForViewerFunc = func(ctx context.Context, limit int) ([]*capi.Session, error) {
					return []*capi.Session{
						{
							ID:           "s1",
							State:        "completed",
							CreatedAt:    sampleDate,
							ResourceType: "pull",
							PullRequest: &api.PullRequest{
								Number: 101,
								Repository: &api.PRRepository{
									NameWithOwner: "OWNER/REPO",
								},
							},
						},
					}, nil
				}
			},
			wantOut: heredoc.Doc(`
				SESSION ID  PULL REQUEST  REPO        SESSION STATE  CREATED
				s1          #101          OWNER/REPO  completed      about 6 hours ago
			`),
		},
		{
			name: "viewer-scoped single session (nontty)",
			tty:  false,
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.ListSessionsForViewerFunc = func(ctx context.Context, limit int) ([]*capi.Session, error) {
					return []*capi.Session{
						{
							ID:           "s1",
							State:        "completed",
							ResourceType: "pull",
							CreatedAt:    sampleDate,
							PullRequest: &api.PullRequest{
								Number: 101,
								Repository: &api.PRRepository{
									NameWithOwner: "OWNER/REPO",
								},
							},
						},
					}, nil
				}
			},
			wantOut: "s1\t#101\tOWNER/REPO\tcompleted\t" + sampleDateString + "\n", // header omitted for non-tty
		},
		{
			name: "viewer-scoped many sessions (tty)",
			tty:  true,
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.ListSessionsForViewerFunc = func(ctx context.Context, limit int) ([]*capi.Session, error) {
					return []*capi.Session{
						{
							ID:           "s1",
							State:        "completed",
							CreatedAt:    sampleDate,
							ResourceType: "pull",
							PullRequest: &api.PullRequest{
								Number: 101,
								Repository: &api.PRRepository{
									NameWithOwner: "OWNER/REPO",
								},
							},
						},
						{
							ID:           "s2",
							State:        "failed",
							CreatedAt:    sampleDate,
							ResourceType: "pull",
							PullRequest: &api.PullRequest{
								Number: 102,
								Repository: &api.PRRepository{
									NameWithOwner: "OWNER/REPO",
								},
							},
						},
						{
							ID:           "s3",
							State:        "in_progress",
							CreatedAt:    sampleDate,
							ResourceType: "pull",
							PullRequest: &api.PullRequest{
								Number: 103,
								Repository: &api.PRRepository{
									NameWithOwner: "OWNER/REPO",
								},
							},
						},
						{
							ID:           "s4",
							State:        "queued",
							CreatedAt:    sampleDate,
							ResourceType: "pull",
							PullRequest: &api.PullRequest{
								Number: 104,
								Repository: &api.PRRepository{
									NameWithOwner: "OWNER/REPO",
								},
							},
						},
						{
							ID:           "s5",
							State:        "canceled",
							CreatedAt:    sampleDate,
							ResourceType: "pull",
							PullRequest: &api.PullRequest{
								Number: 105,
								Repository: &api.PRRepository{
									NameWithOwner: "OWNER/REPO",
								},
							},
						},
						{
							ID:           "s6",
							State:        "mystery",
							CreatedAt:    sampleDate,
							ResourceType: "pull",
							PullRequest: &api.PullRequest{
								Number: 106,
								Repository: &api.PRRepository{
									NameWithOwner: "OWNER/REPO",
								},
							},
						},
					}, nil
				}
			},
			wantOut: heredoc.Doc(`
				SESSION ID  PULL REQUEST  REPO        SESSION STATE  CREATED
				s1          #101          OWNER/REPO  completed      about 6 hours ago
				s2          #102          OWNER/REPO  failed         about 6 hours ago
				s3          #103          OWNER/REPO  in_progress    about 6 hours ago
				s4          #104          OWNER/REPO  queued         about 6 hours ago
				s5          #105          OWNER/REPO  canceled       about 6 hours ago
				s6          #106          OWNER/REPO  mystery        about 6 hours ago
			`),
		},
		{
			name:     "repo-scoped no sessions",
			tty:      true,
			baseRepo: ghrepo.New("OWNER", "REPO"),
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.ListSessionsForRepoFunc = func(ctx context.Context, owner, repo string, limit int) ([]*capi.Session, error) {
					return nil, nil
				}
			},
			wantErr: cmdutil.NewNoResultsError("no agent tasks found"),
		},
		{
			name:     "repo-scoped respects --limit/--repo",
			tty:      true,
			limit:    999,
			baseRepo: ghrepo.New("OWNER", "REPO"),
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.ListSessionsForRepoFunc = func(ctx context.Context, owner, repo string, limit int) ([]*capi.Session, error) {
					assert.Equal(t, 999, limit)
					assert.Equal(t, "OWNER", owner)
					assert.Equal(t, "REPO", repo)
					return nil, nil
				}
			},
			wantErr: cmdutil.NewNoResultsError("no agent tasks found"), // not important
		},
		{
			name:     "repo-scoped single session (tty)",
			tty:      true,
			baseRepo: ghrepo.New("OWNER", "REPO"),
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.ListSessionsForRepoFunc = func(ctx context.Context, owner, repo string, limit int) ([]*capi.Session, error) {
					return []*capi.Session{
						{
							ID:           "s1",
							State:        "completed",
							CreatedAt:    sampleDate,
							ResourceType: "pull",
							PullRequest: &api.PullRequest{
								Number: 101,
								Repository: &api.PRRepository{
									NameWithOwner: "OWNER/REPO",
								},
							},
						},
					}, nil
				}
			},
			wantOut: heredoc.Doc(`
				SESSION ID  PULL REQUEST  REPO        SESSION STATE  CREATED
				s1          #101          OWNER/REPO  completed      about 6 hours ago
			`),
		},
		{
			name:     "repo-scoped single session (nontty)",
			tty:      false,
			baseRepo: ghrepo.New("OWNER", "REPO"),
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.ListSessionsForRepoFunc = func(ctx context.Context, owner, repo string, limit int) ([]*capi.Session, error) {
					return []*capi.Session{
						{
							ID:           "s1",
							State:        "completed",
							ResourceType: "pull",
							CreatedAt:    sampleDate,
							PullRequest: &api.PullRequest{
								Number: 101,
								Repository: &api.PRRepository{
									NameWithOwner: "OWNER/REPO",
								},
							},
						},
					}, nil
				}
			},
			wantOut: "s1\t#101\tOWNER/REPO\tcompleted\t" + sampleDateString + "\n", // header omitted for non-tty
		},
		{
			name:     "repo-scoped many sessions (tty)",
			tty:      true,
			baseRepo: ghrepo.New("OWNER", "REPO"),
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.ListSessionsForRepoFunc = func(ctx context.Context, owner, repo string, limit int) ([]*capi.Session, error) {
					return []*capi.Session{
						{
							ID:           "s1",
							State:        "completed",
							CreatedAt:    sampleDate,
							ResourceType: "pull",
							PullRequest: &api.PullRequest{
								Number: 101,
								Repository: &api.PRRepository{
									NameWithOwner: "OWNER/REPO",
								},
							},
						},
						{
							ID:           "s2",
							State:        "failed",
							CreatedAt:    sampleDate,
							ResourceType: "pull",
							PullRequest: &api.PullRequest{
								Number: 102,
								Repository: &api.PRRepository{
									NameWithOwner: "OWNER/REPO",
								},
							},
						},
						{
							ID:           "s3",
							State:        "in_progress",
							CreatedAt:    sampleDate,
							ResourceType: "pull",
							PullRequest: &api.PullRequest{
								Number: 103,
								Repository: &api.PRRepository{
									NameWithOwner: "OWNER/REPO",
								},
							},
						},
						{
							ID:           "s4",
							State:        "queued",
							CreatedAt:    sampleDate,
							ResourceType: "pull",
							PullRequest: &api.PullRequest{
								Number: 104,
								Repository: &api.PRRepository{
									NameWithOwner: "OWNER/REPO",
								},
							},
						},
						{
							ID:           "s5",
							State:        "canceled",
							CreatedAt:    sampleDate,
							ResourceType: "pull",
							PullRequest: &api.PullRequest{
								Number: 105,
								Repository: &api.PRRepository{
									NameWithOwner: "OWNER/REPO",
								},
							},
						},
						{
							ID:           "s6",
							State:        "mystery",
							CreatedAt:    sampleDate,
							ResourceType: "pull",
							PullRequest: &api.PullRequest{
								Number: 106,
								Repository: &api.PRRepository{
									NameWithOwner: "OWNER/REPO",
								},
							},
						},
					}, nil
				}
			},
			wantOut: heredoc.Doc(`
				SESSION ID  PULL REQUEST  REPO        SESSION STATE  CREATED
				s1          #101          OWNER/REPO  completed      about 6 hours ago
				s2          #102          OWNER/REPO  failed         about 6 hours ago
				s3          #103          OWNER/REPO  in_progress    about 6 hours ago
				s4          #104          OWNER/REPO  queued         about 6 hours ago
				s5          #105          OWNER/REPO  canceled       about 6 hours ago
				s6          #106          OWNER/REPO  mystery        about 6 hours ago
			`),
		},
		{
			name:        "repo resolution error does not surface",
			tty:         true,
			baseRepoErr: errors.New("ambiguous repo"),
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				// We expect a viewer-scoped fetch request:
				m.ListSessionsForViewerFunc = func(ctx context.Context, limit int) ([]*capi.Session, error) {
					return nil, nil
				}
			},
			wantErr: cmdutil.NewNoResultsError("no agent tasks found"),
		},
		{
			name:           "web mode",
			tty:            true,
			web:            true,
			wantOut:        "",
			wantStderr:     "Opening https://github.com/copilot/agents in your browser.\n",
			wantBrowserURL: "https://github.com/copilot/agents",
		},
		{
			name:           "web mode with repo still uses global URL, even when --repo is set",
			tty:            true,
			web:            true,
			baseRepo:       ghrepo.New("OWNER", "REPO"),
			wantOut:        "",
			wantStderr:     "Opening https://github.com/copilot/agents in your browser.\n",
			wantBrowserURL: "https://github.com/copilot/agents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capiClientMock := &capi.CapiClientMock{}
			if tt.capiStubs != nil {
				tt.capiStubs(t, capiClientMock)
			}

			ios, _, stdout, stderr := iostreams.Test()
			ios.SetStdoutTTY(tt.tty)

			var br *browser.Stub
			if tt.web {
				br = &browser.Stub{}
			}

			opts := &ListOptions{
				IO:      ios,
				Limit:   tt.limit,
				Web:     tt.web,
				Browser: br,
				CapiClient: func() (capi.CapiClient, error) {
					if tt.web {
						require.FailNow(t, "CapiClient was called with --web")
					}
					return capiClientMock, nil
				},
			}
			if tt.baseRepo != nil || tt.baseRepoErr != nil {
				baseRepo := tt.baseRepo
				baseRepoErr := tt.baseRepoErr
				opts.BaseRepo = func() (ghrepo.Interface, error) { return baseRepo, baseRepoErr }
			}

			err := listRun(opts)
			if tt.wantErr != nil {
				assert.Error(t, err)
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}
			got := stdout.String()
			require.Equal(t, tt.wantOut, got)
			require.Equal(t, tt.wantStderr, stderr.String())
			if tt.web {
				br.Verify(t, tt.wantBrowserURL)
			}
		})
	}
}
