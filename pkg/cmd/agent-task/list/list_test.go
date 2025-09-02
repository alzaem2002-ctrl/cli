package list

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/internal/browser"
	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/pkg/cmd/agent-task/capi"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdList(t *testing.T) {
	tests := []struct {
		name     string
		args     string
		wantOpts ListOptions
		wantErr  string
	}{
		{
			name: "no arguments",
			wantOpts: ListOptions{
				Limit: defaultLimit,
			},
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
			f := &cmdutil.Factory{}
			var gotOpts *ListOptions
			cmd := NewCmdList(f, func(opts *ListOptions) error { gotOpts = opts; return nil })
			if tt.args != "" {
				cmd.SetArgs(strings.Split(tt.args, " "))
			}
			_, err := cmd.ExecuteC()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantOpts.Limit, gotOpts.Limit)
			assert.Equal(t, tt.wantOpts.Web, gotOpts.Web)
		})
	}
}

func Test_listRun(t *testing.T) {
	createdAt := time.Now().Add(-6 * time.Hour).Format(time.RFC3339) // 6h ago

	tests := []struct {
		name           string
		tty            bool
		stubs          func(*httpmock.Registry)
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
			name:    "no sessions",
			tty:     true,
			stubs:   func(reg *httpmock.Registry) { registerEmptySessionsMock(reg) },
			wantErr: cmdutil.NewNoResultsError("no agent tasks found"),
		},
		{
			name:  "limit truncates sessions",
			tty:   true,
			limit: 3,
			stubs: func(reg *httpmock.Registry) { registerManySessionsMock(reg, createdAt) },
			wantOut: heredoc.Doc(`
			SESSION ID  PULL REQUEST  REPO        SESSION STATE  CREATED
			s1          #101          OWNER/REPO  completed      about 6 hours ago
			s2          #102          OWNER/REPO  failed         about 6 hours ago
			s3          #103          OWNER/REPO  in_progress    about 6 hours ago
			`),
		},
		{
			name:  "single session (tty)",
			tty:   true,
			stubs: func(reg *httpmock.Registry) { registerSingleSessionMock(reg, createdAt) },
			wantOut: heredoc.Doc(`
			SESSION ID  PULL REQUEST  REPO        SESSION STATE  CREATED
			sess1       #42           OWNER/REPO  completed      about 6 hours ago
			`),
		},
		{
			name:    "single session (nontty)",
			tty:     false,
			stubs:   func(reg *httpmock.Registry) { registerSingleSessionMock(reg, createdAt) },
			wantOut: "sess1\t#42\tOWNER/REPO\tcompleted\t" + createdAt + "\n", // header omitted for non-tty
		},
		{
			name:  "many sessions (tty)",
			tty:   true,
			stubs: func(reg *httpmock.Registry) { registerManySessionsMock(reg, createdAt) },
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
			name:     "repo scoped single session",
			tty:      true,
			stubs:    func(reg *httpmock.Registry) { registerRepoSingleSessionMock(reg, createdAt, "OWNER", "REPO") },
			baseRepo: ghrepo.New("OWNER", "REPO"),
			wantOut: heredoc.Doc(`
			SESSION ID  PULL REQUEST  REPO        SESSION STATE  CREATED
			sessR1      #55           OWNER/REPO  completed      about 6 hours ago
			`),
		},
		{
			name:     "repo scoped no sessions",
			tty:      true,
			stubs:    func(reg *httpmock.Registry) { registerRepoEmptySessionsMock(reg, "OWNER", "REPO") },
			baseRepo: ghrepo.New("OWNER", "REPO"),
			wantErr:  cmdutil.NewNoResultsError("no agent tasks found"),
		},
		{
			name:        "repo resolution error does not surface",
			tty:         true,
			baseRepoErr: errors.New("ambiguous repo"),
			wantErr:     cmdutil.NewNoResultsError("no agent tasks found"),
			stubs:       func(reg *httpmock.Registry) { registerEmptySessionsMock(reg) },
		},
		{
			name:     "repo scoped many sessions (tty)",
			tty:      true,
			stubs:    func(reg *httpmock.Registry) { registerRepoManySessionsMock(reg, createdAt, "OWNER", "REPO") },
			baseRepo: ghrepo.New("OWNER", "REPO"),
			wantOut: heredoc.Doc(`
			SESSION ID  PULL REQUEST  REPO        SESSION STATE  CREATED
			r1          #301          OWNER/REPO  completed      about 6 hours ago
			r2          #302          OWNER/REPO  failed         about 6 hours ago
			r3          #303          OWNER/REPO  in_progress    about 6 hours ago
			r4          #304          OWNER/REPO  queued         about 6 hours ago
			r5          #305          OWNER/REPO  canceled       about 6 hours ago
			r6          #306          OWNER/REPO  mystery        about 6 hours ago
			`),
		},
		{
			name:           "web mode",
			tty:            true,
			web:            true,
			wantOut:        "",
			wantStderr:     "Opening https://github.com/copilot/agents in your browser.\n",
			wantBrowserURL: "https://github.com/copilot/agents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &httpmock.Registry{}
			if tt.stubs != nil {
				tt.stubs(reg)
			}

			cfg := config.NewBlankConfig()
			cfg.Set("github.com", "oauth_token", "OTOKEN")
			authCfg := cfg.Authentication()

			ios, _, stdout, stderr := iostreams.Test()
			ios.SetStdoutTTY(tt.tty)

			var br *browser.Stub
			if tt.web {
				br = &browser.Stub{}
			}

			httpClient := &http.Client{Transport: reg}
			capiClient := capi.NewCAPIClient(httpClient, authCfg)
			opts := &ListOptions{
				IO:      ios,
				Config:  func() (gh.Config, error) { return cfg, nil },
				Limit:   tt.limit,
				Web:     tt.web,
				Browser: br,
				CapiClient: func() (*capi.CAPIClient, error) {
					if tt.web {
						panic("CapiClient should not be invoked when --web is set")
					}
					return capiClient, nil
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
			reg.Verify(t)
		})
	}
}

// registerRepoSingleSessionMock mocks repo-scoped endpoint with one session and hydration.
func registerRepoSingleSessionMock(reg *httpmock.Registry, createdAt, owner, repo string) {
	reg.Register(
		httpmock.WithHost(httpmock.REST("GET", "agents/sessions/nwo/"+owner+"/"+repo), "api.githubcopilot.com"),
		httpmock.StringResponse(heredoc.Docf(`{
			"sessions": [
				{
					"id": "sessR1",
					"name": "Repo build",
					"user_id": 1,
					"agent_id": 2,
					"logs": "",
					"state": "completed",
					"owner_id": 10,
					"repo_id": 1000,
					"resource_type": "pull",
					"resource_id": 3000,
					"created_at": "%[1]s"
				}
			]
		}`, createdAt)),
	)
	// Second page empty (pagination end)
	reg.Register(
		httpmock.WithHost(httpmock.REST("GET", "agents/sessions/nwo/"+owner+"/"+repo), "api.githubcopilot.com"),
		httpmock.StringResponse(heredoc.Doc(`{
			"sessions": []
		}`)),
	)
	// Hydration
	reg.Register(
		httpmock.GraphQL(`query FetchPRs`),
		httpmock.StringResponse(heredoc.Docf(`{
	"data": {
		"nodes": [
			{
				"id": "PR_nodeR1",
				"fullDatabaseId": "3000",
				"number": 55,
				"title": "Improve build",
				"state": "OPEN",
				"url": "https://github.com/%[2]s/%[3]s/pull/55",
				"body": "",
				"createdAt": "%[1]s",
				"updatedAt": "%[1]s",
				"repository": { "nameWithOwner": "%[2]s/%[3]s" }
			}
		]
	}
}`, createdAt, owner, repo)),
	)
}

// registerRepoEmptySessionsMock mocks repo-scoped endpoint returning no sessions.
func registerRepoEmptySessionsMock(reg *httpmock.Registry, owner, repo string) {
	reg.Register(
		httpmock.WithHost(httpmock.REST("GET", "agents/sessions/nwo/"+owner+"/"+repo), "api.githubcopilot.com"),
		httpmock.StringResponse(heredoc.Doc(`{
	"sessions": []
}`)),
	)
}

// registerRepoManySessionsMock mirrors registerManySessionsMock but for repo-scoped endpoint
func registerRepoManySessionsMock(reg *httpmock.Registry, createdAt, owner, repo string) {
	reg.Register(
		httpmock.WithHost(httpmock.REST("GET", "agents/sessions/nwo/"+owner+"/"+repo), "api.githubcopilot.com"),
		httpmock.StringResponse(heredoc.Docf(`{
			"sessions": [
				{
					"id": "r1",
					"name": "A",
					"user_id": 1,
					"agent_id": 2,
					"logs": "",
					"state": "completed",
					"owner_id": 10,
					"repo_id": 1000,
					"resource_type": "pull",
					"resource_id": 3001,
					"created_at": "%[1]s"
				},
				{
					"id": "r2",
					"name": "B",
					"user_id": 1,
					"agent_id": 2,
					"logs": "",
					"state": "failed",
					"owner_id": 10,
					"repo_id": 1000,
					"resource_type": "pull",
					"resource_id": 3002,
					"created_at": "%[1]s"
				},
				{
					"id": "r3",
					"name": "C",
					"user_id": 1,
					"agent_id": 2,
					"logs": "",
					"state": "in_progress",
					"owner_id": 10,
					"repo_id": 1000,
					"resource_type": "pull",
					"resource_id": 3003,
					"created_at": "%[1]s"
				},
				{
					"id": "r4",
					"name": "D",
					"user_id": 1,
					"agent_id": 2,
					"logs": "",
					"state": "queued",
					"owner_id": 10,
					"repo_id": 1000,
					"resource_type": "pull",
					"resource_id": 3004,
					"created_at": "%[1]s"
				},
				{
					"id": "r5",
					"name": "E",
					"user_id": 1,
					"agent_id": 2,
					"logs": "",
					"state": "canceled",
					"owner_id": 10,
					"repo_id": 1000,
					"resource_type": "pull",
					"resource_id": 3005,
					"created_at": "%[1]s"
				},
				{
					"id": "r6",
					"name": "F",
					"user_id": 1,
					"agent_id": 2,
					"logs": "",
					"state": "mystery",
					"owner_id": 10,
					"repo_id": 1000,
					"resource_type": "pull",
					"resource_id": 3006,
					"created_at": "%[1]s"
				}
			]
		}`, createdAt)),
	)
	reg.Register(
		httpmock.WithHost(httpmock.REST("GET", "agents/sessions/nwo/"+owner+"/"+repo), "api.githubcopilot.com"),
		httpmock.StringResponse(heredoc.Doc(`{
			"sessions": []
		}`)),
	)
	reg.Register(
		httpmock.GraphQL(`query FetchPRs`),
		httpmock.StringResponse(heredoc.Docf(`{
			"data": {
				"nodes": [
					{
						"id": "PR_r1",
						"fullDatabaseId": "3001",
						"number": 301,
						"title": "PR 301",
						"state": "OPEN",
						"url": "https://github.com/%[2]s/%[3]s/pull/301",
						"body": "",
						"createdAt": "%[1]s",
						"updatedAt": "%[1]s",
						"repository": {
							"nameWithOwner": "%[2]s/%[3]s"
						}
					},
					{
						"id": "PR_r2",
						"fullDatabaseId": "3002",
						"number": 302,
						"title": "PR 302",
						"state": "OPEN",
						"url": "https://github.com/%[2]s/%[3]s/pull/302",
						"body": "",
						"createdAt": "%[1]s",
						"updatedAt": "%[1]s",
						"repository": {
							"nameWithOwner": "%[2]s/%[3]s"
						}
					},
					{
						"id": "PR_r3",
						"fullDatabaseId": "3003",
						"number": 303,
						"title": "PR 303",
						"state": "OPEN",
						"url": "https://github.com/%[2]s/%[3]s/pull/303",
						"body": "",
						"createdAt": "%[1]s",
						"updatedAt": "%[1]s",
						"repository": {
							"nameWithOwner": "%[2]s/%[3]s"
						}
					},
					{
						"id": "PR_r4",
						"fullDatabaseId": "3004",
						"number": 304,
						"title": "PR 304",
						"state": "OPEN",
						"url": "https://github.com/%[2]s/%[3]s/pull/304",
						"body": "",
						"createdAt": "%[1]s",
						"updatedAt": "%[1]s",
						"repository": {
							"nameWithOwner": "%[2]s/%[3]s"
						}
					},
					{
						"id": "PR_r5",
						"fullDatabaseId": "3005",
						"number": 305,
						"title": "PR 305",
						"state": "OPEN",
						"url": "https://github.com/%[2]s/%[3]s/pull/305",
						"body": "",
						"createdAt": "%[1]s",
						"updatedAt": "%[1]s",
						"repository": {
							"nameWithOwner": "%[2]s/%[3]s"
						}
					},
					{
						"id": "PR_r6",
						"fullDatabaseId": "3006",
						"number": 306,
						"title": "PR 306",
						"state": "OPEN",
						"url": "https://github.com/%[2]s/%[3]s/pull/306",
						"body": "",
						"createdAt": "%[1]s",
						"updatedAt": "%[1]s",
						"repository": {
							"nameWithOwner": "%[2]s/%[3]s"
						}
					}
				]
			}
		}`, createdAt, owner, repo)),
	)
}

// registerEmptySessionsMock registers a single empty page of sessions
func registerEmptySessionsMock(reg *httpmock.Registry) {
	reg.Register(
		httpmock.WithHost(httpmock.REST("GET", "agents/sessions"), "api.githubcopilot.com"),
		httpmock.StringResponse(heredoc.Doc(`{
			"sessions": []
		}`)),
	)
}

// registerSingleSessionMock registers two REST pages (one with a session, one empty) and GraphQL hydration for that session's PR
func registerSingleSessionMock(reg *httpmock.Registry, createdAt string) {
	// First page with one session
	reg.Register(
		httpmock.WithHost(httpmock.REST("GET", "agents/sessions"), "api.githubcopilot.com"),
		httpmock.StringResponse(heredoc.Docf(`{
	"sessions": [
		{
			"id": "sess1",
			"name": "Build artifacts",
			"user_id": 1,
			"agent_id": 2,
			"logs": "",
			"state": "completed",
			"owner_id": 10,
			"repo_id": 1000,
			"resource_type": "pull",
			"resource_id": 2000,
			"created_at": "%[1]s"
		}
	]
}`, createdAt)),
	)
	// Second page empty to terminate pagination
	reg.Register(
		httpmock.WithHost(httpmock.REST("GET", "agents/sessions"), "api.githubcopilot.com"),
		httpmock.StringResponse(heredoc.Doc(`{
			"sessions": []
		}`)),
	)
	// GraphQL hydration
	reg.Register(
		httpmock.GraphQL(`query FetchPRs`),
		httpmock.StringResponse(heredoc.Docf(`{
			"data": {
				"nodes": [
					{
						"id": "PR_node",
						"fullDatabaseId": "2000",
						"number": 42,
						"title": "Improve docs",
						"state": "OPEN",
						"url": "https://github.com/OWNER/REPO/pull/42",
						"body": "",
						"createdAt": "%[1]s",
						"updatedAt": "%[1]s",
						"repository": {
							"nameWithOwner": "OWNER/REPO"
						}
					}
				]
			}
		}`, createdAt)),
	)
}

// registerManySessionsMock registers multiple sessions covering various states
// States covered: completed, failed, in_progress, queued, canceled, (unknown -> treated as muted)
func registerManySessionsMock(reg *httpmock.Registry, createdAt string) {
	// First page returns six sessions
	reg.Register(
		httpmock.WithHost(httpmock.REST("GET", "agents/sessions"), "api.githubcopilot.com"),
		httpmock.StringResponse(heredoc.Docf(`{
	"sessions": [
		{
			"id": "s1",
			"name": "A",
			"user_id": 1,
			"agent_id": 2,
			"logs": "",
			"state": "completed",
			"owner_id": 10,
			"repo_id": 1000,
			"resource_type": "pull",
			"resource_id": 2000,
			"created_at": "%[1]s"
		},
		{
			"id": "s2",
			"name": "B",
			"user_id": 1,
			"agent_id": 2,
			"logs": "",
			"state": "failed",
			"owner_id": 10,
			"repo_id": 1000,
			"resource_type": "pull",
			"resource_id": 2001,
			"created_at": "%[1]s"
		},
		{
			"id": "s3",
			"name": "C",
			"user_id": 1,
			"agent_id": 2,
			"logs": "",
			"state": "in_progress",
			"owner_id": 10,
			"repo_id": 1000,
			"resource_type": "pull",
			"resource_id": 2002,
			"created_at": "%[1]s"
		},
		{
			"id": "s4",
			"name": "D",
			"user_id": 1,
			"agent_id": 2,
			"logs": "",
			"state": "queued",
			"owner_id": 10,
			"repo_id": 1000,
			"resource_type": "pull",
			"resource_id": 2003,
			"created_at": "%[1]s"
		},
		{
			"id": "s5",
			"name": "E",
			"user_id": 1,
			"agent_id": 2,
			"logs": "",
			"state": "canceled",
			"owner_id": 10,
			"repo_id": 1000,
			"resource_type": "pull",
			"resource_id": 2004,
			"created_at": "%[1]s"
		},
		{
			"id": "s6",
			"name": "F",
			"user_id": 1,
			"agent_id": 2,
			"logs": "",
			"state": "mystery",
			"owner_id": 10,
			"repo_id": 1000,
			"resource_type": "pull",
			"resource_id": 2005,
			"created_at": "%[1]s"
		}
	]
}`, createdAt)),
	)
	// Second page empty
	reg.Register(
		httpmock.WithHost(httpmock.REST("GET", "agents/sessions"), "api.githubcopilot.com"),
		httpmock.StringResponse(heredoc.Doc(`{
			"sessions": []
		}`)),
	)
	// GraphQL hydration for 6 PRs
	reg.Register(
		httpmock.GraphQL(`query FetchPRs`),
		httpmock.StringResponse(heredoc.Docf(`{
			"data": {
				"nodes": [
					{
						"id": "PR_node1",
						"fullDatabaseId": "2000",
						"number": 101,
						"title": "PR 101",
						"state": "OPEN",
						"url": "https://github.com/OWNER/REPO/pull/101",
						"body": "",
						"createdAt": "%[1]s",
						"updatedAt": "%[1]s",
						"repository": {
							"nameWithOwner": "OWNER/REPO"
						}
					},
					{
						"id": "PR_node2",
						"fullDatabaseId": "2001",
						"number": 102,
						"title": "PR 102",
						"state": "OPEN",
						"url": "https://github.com/OWNER/REPO/pull/102",
						"body": "",
						"createdAt": "%[1]s",
						"updatedAt": "%[1]s",
						"repository": {
							"nameWithOwner": "OWNER/REPO"
						}
					},
					{
						"id": "PR_node3",
						"fullDatabaseId": "2002",
						"number": 103,
						"title": "PR 103",
						"state": "OPEN",
						"url": "https://github.com/OWNER/REPO/pull/103",
						"body": "",
						"createdAt": "%[1]s",
						"updatedAt": "%[1]s",
						"repository": {
							"nameWithOwner": "OWNER/REPO"
						}
					},
					{
						"id": "PR_node4",
						"fullDatabaseId": "2003",
						"number": 104,
						"title": "PR 104",
						"state": "OPEN",
						"url": "https://github.com/OWNER/REPO/pull/104",
						"body": "",
						"createdAt": "%[1]s",
						"updatedAt": "%[1]s",
						"repository": {
							"nameWithOwner": "OWNER/REPO"
						}
					},
					{
						"id": "PR_node5",
						"fullDatabaseId": "2004",
						"number": 105,
						"title": "PR 105",
						"state": "OPEN",
						"url": "https://github.com/OWNER/REPO/pull/105",
						"body": "",
						"createdAt": "%[1]s",
						"updatedAt": "%[1]s",
						"repository": {
							"nameWithOwner": "OWNER/REPO"
						}
					},
					{
						"id": "PR_node6",
						"fullDatabaseId": "2005",
						"number": 106,
						"title": "PR 106",
						"state": "OPEN",
						"url": "https://github.com/OWNER/REPO/pull/106",
						"body": "",
						"createdAt": "%[1]s",
						"updatedAt": "%[1]s",
						"repository": {
							"nameWithOwner": "OWNER/REPO"
						}
					}
				]
			}
		}`, createdAt)),
	)
}
