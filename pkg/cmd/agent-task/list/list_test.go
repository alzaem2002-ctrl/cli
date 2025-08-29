package list

import (
	"net/http"
	"testing"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/pkg/cmd/agent-task/capi"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdList(t *testing.T) {
	tests := []struct {
		name     string
		cli      string
		wantOpts ListOptions
	}{
		{
			name: "no arguments",
			wantOpts: ListOptions{
				Limit: defaultLimit,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &cmdutil.Factory{}
			var gotOpts *ListOptions
			cmd := NewCmdList(f, func(opts *ListOptions) error {
				gotOpts = opts
				return nil
			})
			cmd.ExecuteC()

			assert.Equal(t, tt.wantOpts.Limit, gotOpts.Limit)
		})
	}
}

func Test_listRun(t *testing.T) {
	sixHours, _ := time.ParseDuration("6h")
	sixHoursAgo := time.Now().Add(-sixHours)
	createdAt := sixHoursAgo.Format(time.RFC3339)

	tests := []struct {
		name    string
		tty     bool
		stubs   func(*httpmock.Registry)
		wantOut string
	}{
		{
			name:    "no sessions",
			tty:     true,
			stubs:   func(reg *httpmock.Registry) { registerEmptySessionsMock(reg) },
			wantOut: "no agent tasks found\n",
		},
		{
			name:  "single session (tty)",
			tty:   true,
			stubs: func(reg *httpmock.Registry) { registerSingleSessionMock(reg, createdAt) },
			wantOut: "SESSION ID  PULL REQUEST  REPO        SESSION STATE  CREATED\n" +
				"sess1       #42           OWNER/REPO  completed      about 6 hours ago\n",
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
			wantOut: "SESSION ID  PULL REQUEST  REPO        SESSION STATE  CREATED\n" +
				"s1          #101          OWNER/REPO  completed      about 6 hours ago\n" +
				"s2          #102          OWNER/REPO  failed         about 6 hours ago\n" +
				"s3          #103          OWNER/REPO  in_progress    about 6 hours ago\n" +
				"s4          #104          OWNER/REPO  queued         about 6 hours ago\n" +
				"s5          #105          OWNER/REPO  canceled       about 6 hours ago\n" +
				"s6          #106          OWNER/REPO  mystery        about 6 hours ago\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &httpmock.Registry{}
			tt.stubs(reg)

			cfg := config.NewBlankConfig()
			cfg.Set("github.com", "oauth_token", "OTOKEN")
			authCfg := cfg.Authentication()

			ios, _, stdout, _ := iostreams.Test()
			ios.SetStdoutTTY(tt.tty)

			httpClient := &http.Client{Transport: reg}
			capiClient := capi.NewCAPIClient(httpClient, authCfg)
			opts := &ListOptions{
				IO:         ios,
				Config:     func() (gh.Config, error) { return cfg, nil },
				Limit:      30,
				CapiClient: func() (*capi.CAPIClient, error) { return capiClient, nil },
			}

			err := listRun(opts)
			assert.NoError(t, err)

			got := stdout.String()
			if tt.wantOut == "" && tt.name == "single session (tty)" {
				t.Logf("Captured output for single session (tty):\n%s", got)
				t.Fatalf("fill in wantOut with the above output and re-run tests")
			}
			assert.Equal(t, tt.wantOut, got)
			reg.Verify(t)
		})
	}
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
		httpmock.StringResponse(`{"sessions": []}`),
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
				"repository": { "nameWithOwner": "OWNER/REPO" }
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
		httpmock.StringResponse(`{"sessions": []}`),
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
						"repository": { "nameWithOwner": "OWNER/REPO" }
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
						"repository": { "nameWithOwner": "OWNER/REPO" }
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
						"repository": { "nameWithOwner": "OWNER/REPO" }
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
						"repository": { "nameWithOwner": "OWNER/REPO" }
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
						"repository": { "nameWithOwner": "OWNER/REPO" }
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
						"repository": { "nameWithOwner": "OWNER/REPO" }
					}
				]
			}
		}`, createdAt)),
	)
}
