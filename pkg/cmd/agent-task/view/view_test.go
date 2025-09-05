package view

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/pkg/cmd/agent-task/capi"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdList(t *testing.T) {
	tests := []struct {
		name         string
		tty          bool
		args         string
		wantOpts     ViewOptions
		wantBaseRepo ghrepo.Interface
		wantErr      string
	}{
		{
			name:     "no arg tty",
			tty:      true,
			args:     "",
			wantOpts: ViewOptions{},
		},
		{
			name: "session ID arg tty",
			tty:  true,
			args: "00000000-0000-0000-0000-000000000000",
			wantOpts: ViewOptions{
				SelectorArg: "00000000-0000-0000-0000-000000000000",
				SessionID:   "00000000-0000-0000-0000-000000000000",
			},
		},
		{
			name: "non-session ID arg tty",
			tty:  true,
			args: "some-arg",
			wantOpts: ViewOptions{
				SelectorArg: "some-arg",
			},
		},
		{
			name:    "session ID required if non-tty",
			tty:     false,
			args:    "some-arg",
			wantErr: "session ID is required when not running interactively",
		},
		{
			name:         "repo override",
			tty:          true,
			args:         "some-arg -R OWNER/REPO",
			wantBaseRepo: ghrepo.New("OWNER", "REPO"),
			wantOpts: ViewOptions{
				SelectorArg: "some-arg",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			ios.SetStdinTTY(tt.tty)
			ios.SetStdoutTTY(tt.tty)
			ios.SetStderrTTY(tt.tty)

			f := &cmdutil.Factory{
				IOStreams: ios,
			}

			var gotOpts *ViewOptions
			cmd := NewCmdView(f, func(opts *ViewOptions) error { gotOpts = opts; return nil })

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
			assert.Equal(t, tt.wantOpts.SelectorArg, gotOpts.SelectorArg)
		})
	}
}

func Test_viewRun(t *testing.T) {
	sampleDate := time.Now().Add(-6 * time.Hour) // 6h ago

	tests := []struct {
		name        string
		selectorArg string
		tty         bool
		capiStubs   func(*testing.T, *capi.CapiClientMock)
		wantOut     string
		wantErr     error
		wantStderr  string
	}{
		{
			name:        "not found (tty)",
			tty:         true,
			selectorArg: "some-session-id",
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.GetSessionFunc = func(ctx context.Context, selector string) (*capi.Session, error) {
					return nil, capi.ErrSessionNotFound
				}
			},
			wantStderr: "session not found\n",
			wantErr:    cmdutil.SilentError,
		},
		{
			name:        "not found (nontty)",
			selectorArg: "some-session-id",
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.GetSessionFunc = func(ctx context.Context, selector string) (*capi.Session, error) {
					return nil, capi.ErrSessionNotFound
				}
			},
			wantStderr: "session not found\n",
			wantErr:    cmdutil.SilentError,
		},
		{
			name:        "API error (tty)",
			tty:         true,
			selectorArg: "some-session-id",
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.GetSessionFunc = func(ctx context.Context, selector string) (*capi.Session, error) {
					return nil, errors.New("some error")
				}
			},
			wantErr: errors.New("some error"),
		},
		{
			name:        "API error (nontty)",
			selectorArg: "some-session-id",
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.GetSessionFunc = func(ctx context.Context, selector string) (*capi.Session, error) {
					return nil, errors.New("some error")
				}
			},
			wantErr: errors.New("some error"),
		},
		{
			name:        "success, with PR and user data (tty)",
			tty:         true,
			selectorArg: "some-session-id",
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.GetSessionFunc = func(ctx context.Context, selector string) (*capi.Session, error) {
					return &capi.Session{
						ID:        "some-session-id",
						State:     "completed",
						CreatedAt: sampleDate,
						PullRequest: &api.PullRequest{
							Title:  "fix something",
							Number: 101,
							URL:    "https://github.com/OWNER/REPO/pull/101",
							Repository: &api.PRRepository{
								NameWithOwner: "OWNER/REPO",
							},
						},
						User: &api.GitHubUser{
							Login: "octocat",
						},
					}, nil
				}
			},
			wantOut: heredoc.Doc(`
				Completed • fix something • OWNER/REPO#101
				Started on behalf of octocat about 6 hours ago
				
				View this session on GitHub:
				https://github.com/OWNER/REPO/pull/101/agent-sessions/some-session-id
			`),
		},
		{
			name:        "success, without user data (tty)",
			tty:         true,
			selectorArg: "some-session-id",
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.GetSessionFunc = func(ctx context.Context, selector string) (*capi.Session, error) {
					return &capi.Session{
						ID:        "some-session-id",
						State:     "completed",
						CreatedAt: sampleDate,
						PullRequest: &api.PullRequest{
							Title:  "fix something",
							Number: 101,
							URL:    "https://github.com/OWNER/REPO/pull/101",
							Repository: &api.PRRepository{
								NameWithOwner: "OWNER/REPO",
							},
						},
					}, nil
				}
			},
			wantOut: heredoc.Doc(`
				Completed • fix something • OWNER/REPO#101
				Started about 6 hours ago
				
				View this session on GitHub:
				https://github.com/OWNER/REPO/pull/101/agent-sessions/some-session-id
			`),
		},
		{
			name:        "success, without PR data (tty)",
			tty:         true,
			selectorArg: "some-session-id",
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.GetSessionFunc = func(ctx context.Context, selector string) (*capi.Session, error) {
					return &capi.Session{
						ID:        "some-session-id",
						State:     "completed",
						CreatedAt: sampleDate,
						User: &api.GitHubUser{
							Login: "octocat",
						},
					}, nil
				}
			},
			wantOut: heredoc.Doc(`
				Completed
				Started on behalf of octocat about 6 hours ago
			`),
		},
		{
			name:        "success, without PR nor user data (tty)",
			tty:         true,
			selectorArg: "some-session-id",
			capiStubs: func(t *testing.T, m *capi.CapiClientMock) {
				m.GetSessionFunc = func(ctx context.Context, selector string) (*capi.Session, error) {
					return &capi.Session{
						ID:        "some-session-id",
						State:     "completed",
						CreatedAt: sampleDate,
					}, nil
				}
			},
			wantOut: heredoc.Doc(`
				Completed
				Started about 6 hours ago
			`),
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

			opts := &ViewOptions{
				IO: ios,
				CapiClient: func() (capi.CapiClient, error) {
					return capiClientMock, nil
				},
				SelectorArg: tt.selectorArg,
			}

			err := viewRun(opts)
			if tt.wantErr != nil {
				assert.Error(t, err)
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}

			got := stdout.String()
			require.Equal(t, tt.wantOut, got)
			require.Equal(t, tt.wantStderr, stderr.String())
		})
	}
}
