package capi

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/api"

	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListSessionsForViewer(t *testing.T) {
	sampleDateString := "2025-08-29T00:00:00Z"
	sampleDate, err := time.Parse(time.RFC3339, sampleDateString)
	require.NoError(t, err)

	tests := []struct {
		name      string
		perPage   int
		limit     int
		httpStubs func(*testing.T, *httpmock.Registry)
		wantErr   string
		wantOut   []*Session
	}{
		{
			name:    "zero limit",
			limit:   0,
			wantOut: nil,
		},
		{
			name:  "no sessions",
			limit: 10,
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions", url.Values{
							"page_number": {"1"},
							"page_size":   {"50"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StringResponse(`{"sessions":[]}`),
				)
			},
			wantOut: nil,
		},
		{
			name:  "single session",
			limit: 10,
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions", url.Values{
							"page_number": {"1"},
							"page_size":   {"50"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StringResponse(heredoc.Docf(`
						{
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
						}`,
						sampleDateString,
					)),
				)
				// GraphQL hydration
				reg.Register(
					httpmock.GraphQL(`query FetchPRsAndUsersForAgentTaskSessions\b`),
					httpmock.GraphQLQuery(heredoc.Docf(`
						{
							"data": {
								"nodes": [
									{
										"__typename": "PullRequest",
										"id": "PR_node",
										"fullDatabaseId": "2000",
										"number": 42,
										"title": "Improve docs",
										"state": "OPEN",
										"isDraft": true,
										"url": "https://github.com/OWNER/REPO/pull/42",
										"body": "",
										"createdAt": "%[1]s",
										"updatedAt": "%[1]s",
										"repository": {
											"nameWithOwner": "OWNER/REPO"
										}
									},
									{
										"__typename": "User",
										"login": "octocat",
										"name": "Octocat",
										"databaseId": 1
									}
								]
							}
						}`,
						sampleDateString,
					), func(q string, vars map[string]interface{}) {
						assert.Equal(t, []interface{}{"PR_kwDNA-jNB9A", "U_kgAB"}, vars["ids"])
					}),
				)
			},
			wantOut: []*Session{
				{

					ID:           "sess1",
					Name:         "Build artifacts",
					UserID:       1,
					AgentID:      2,
					Logs:         "",
					State:        "completed",
					OwnerID:      10,
					RepoID:       1000,
					ResourceType: "pull",
					ResourceID:   2000,
					CreatedAt:    sampleDate,
					PullRequest: &api.PullRequest{
						ID:             "PR_node",
						FullDatabaseID: "2000",
						Number:         42,
						Title:          "Improve docs",
						State:          "OPEN",
						IsDraft:        true,
						URL:            "https://github.com/OWNER/REPO/pull/42",
						Body:           "",
						CreatedAt:      sampleDate,
						UpdatedAt:      sampleDate,
						Repository: &api.PRRepository{
							NameWithOwner: "OWNER/REPO",
						},
					},
					User: &api.GitHubUser{
						Login:      "octocat",
						Name:       "Octocat",
						DatabaseID: 1,
					},
				},
			},
		},
		{
			// This happens at the early moments of a session lifecycle, before a PR is created and associated with it.
			name:  "single session, no pull request resource",
			limit: 10,
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions", url.Values{
							"page_number": {"1"},
							"page_size":   {"50"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StringResponse(heredoc.Docf(`
						{
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
									"resource_type": "",
									"resource_id": 0,
									"created_at": "%[1]s"
								}
							]
						}`,
						sampleDateString,
					)),
				)
				// GraphQL hydration
				reg.Register(
					httpmock.GraphQL(`query FetchPRsAndUsersForAgentTaskSessions\b`),
					httpmock.GraphQLQuery(heredoc.Docf(`
						{
							"data": {
								"nodes": [
									{
										"__typename": "User",
										"login": "octocat",
										"name": "Octocat",
										"databaseId": 1
									}
								]
							}
						}`,
						sampleDateString,
					), func(q string, vars map[string]interface{}) {
						assert.Equal(t, []interface{}{"U_kgAB"}, vars["ids"])
					}),
				)
			},
			wantOut: []*Session{
				{

					ID:           "sess1",
					Name:         "Build artifacts",
					UserID:       1,
					AgentID:      2,
					Logs:         "",
					State:        "completed",
					OwnerID:      10,
					RepoID:       1000,
					ResourceType: "",
					ResourceID:   0,
					CreatedAt:    sampleDate,
					User: &api.GitHubUser{
						Login:      "octocat",
						Name:       "Octocat",
						DatabaseID: 1,
					},
				},
			},
		},
		{
			name:    "multiple sessions, paginated",
			perPage: 1, // to enforce pagination
			limit:   2,
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions", url.Values{
							"page_number": {"1"},
							"page_size":   {"1"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StringResponse(heredoc.Docf(`
						{
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
						}`,
						sampleDateString,
					)),
				)

				// Second page
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions", url.Values{
							"page_number": {"2"},
							"page_size":   {"1"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StringResponse(heredoc.Docf(`
						{
							"sessions": [
								{
									"id": "sess2",
									"name": "Build artifacts",
									"user_id": 1,
									"agent_id": 2,
									"logs": "",
									"state": "completed",
									"owner_id": 10,
									"repo_id": 1000,
									"resource_type": "pull",
									"resource_id": 2001,
									"created_at": "%[1]s"
								}
							]
						}`,
						sampleDateString,
					)),
				)
				// GraphQL hydration
				reg.Register(
					httpmock.GraphQL(`query FetchPRsAndUsersForAgentTaskSessions\b`),
					httpmock.GraphQLQuery(heredoc.Docf(`
						{
							"data": {
								"nodes": [
									{
										"__typename": "PullRequest",
										"id": "PR_node",
										"fullDatabaseId": "2000",
										"number": 42,
										"title": "Improve docs",
										"state": "OPEN",
										"isDraft": true,
										"url": "https://github.com/OWNER/REPO/pull/42",
										"body": "",
										"createdAt": "%[1]s",
										"updatedAt": "%[1]s",
										"repository": {
											"nameWithOwner": "OWNER/REPO"
										}
									},
									{
										"__typename": "PullRequest",
										"id": "PR_node",
										"fullDatabaseId": "2001",
										"number": 43,
										"title": "Improve docs",
										"state": "OPEN",
										"isDraft": true,
										"url": "https://github.com/OWNER/REPO/pull/43",
										"body": "",
										"createdAt": "%[1]s",
										"updatedAt": "%[1]s",
										"repository": {
											"nameWithOwner": "OWNER/REPO"
										}
									},
									{
										"__typename": "User",
										"login": "octocat",
										"name": "Octocat",
										"databaseId": 1
									}
								]
							}
						}`,
						sampleDateString,
					), func(q string, vars map[string]interface{}) {
						assert.Equal(t, []interface{}{"PR_kwDNA-jNB9A", "PR_kwDNA-jNB9E", "U_kgAB"}, vars["ids"])
					}),
				)
			},
			wantOut: []*Session{
				{
					ID:           "sess1",
					Name:         "Build artifacts",
					UserID:       1,
					AgentID:      2,
					Logs:         "",
					State:        "completed",
					OwnerID:      10,
					RepoID:       1000,
					ResourceType: "pull",
					ResourceID:   2000,
					CreatedAt:    sampleDate,
					PullRequest: &api.PullRequest{
						ID:             "PR_node",
						FullDatabaseID: "2000",
						Number:         42,
						Title:          "Improve docs",
						State:          "OPEN",
						IsDraft:        true,
						URL:            "https://github.com/OWNER/REPO/pull/42",
						Body:           "",
						CreatedAt:      sampleDate,
						UpdatedAt:      sampleDate,
						Repository: &api.PRRepository{
							NameWithOwner: "OWNER/REPO",
						},
					},
					User: &api.GitHubUser{
						Login:      "octocat",
						Name:       "Octocat",
						DatabaseID: 1,
					},
				},
				{
					ID:           "sess2",
					Name:         "Build artifacts",
					UserID:       1,
					AgentID:      2,
					Logs:         "",
					State:        "completed",
					OwnerID:      10,
					RepoID:       1000,
					ResourceType: "pull",
					ResourceID:   2001,
					CreatedAt:    sampleDate,
					PullRequest: &api.PullRequest{
						ID:             "PR_node",
						FullDatabaseID: "2001",
						Number:         43,
						Title:          "Improve docs",
						State:          "OPEN",
						IsDraft:        true,
						URL:            "https://github.com/OWNER/REPO/pull/43",
						Body:           "",
						CreatedAt:      sampleDate,
						UpdatedAt:      sampleDate,
						Repository: &api.PRRepository{
							NameWithOwner: "OWNER/REPO",
						},
					},
					User: &api.GitHubUser{
						Login:      "octocat",
						Name:       "Octocat",
						DatabaseID: 1,
					},
				},
			},
		},
		{
			name:  "API error",
			limit: 10,
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions", url.Values{
							"page_number": {"1"},
							"page_size":   {"50"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StatusStringResponse(500, "{}"),
				)
			},
			wantErr: "failed to list sessions:",
		}, {
			name:  "API error at hydration",
			limit: 10,
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions", url.Values{
							"page_number": {"1"},
							"page_size":   {"50"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StringResponse(heredoc.Docf(`
						{
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
						}`,
						sampleDateString,
					)),
				)
				// GraphQL hydration
				reg.Register(
					httpmock.GraphQL(`query FetchPRsAndUsersForAgentTaskSessions\b`),
					httpmock.StatusStringResponse(500, `{}`),
				)
			},
			wantErr: `failed to fetch session resources: non-200 OK status code:`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &httpmock.Registry{}
			if tt.httpStubs != nil {
				tt.httpStubs(t, reg)
			}
			defer reg.Verify(t)

			httpClient := &http.Client{Transport: reg}

			cfg := config.NewBlankConfig()
			capiClient := NewCAPIClient(httpClient, cfg.Authentication())

			if tt.perPage != 0 {
				last := defaultSessionsPerPage
				defaultSessionsPerPage = tt.perPage
				defer func() {
					defaultSessionsPerPage = last
				}()
			}

			sessions, err := capiClient.ListSessionsForViewer(context.Background(), tt.limit)

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				require.Nil(t, sessions)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantOut, sessions)
		})
	}
}

func TestListSessionForRepoRequiresRepo(t *testing.T) {
	client := &CAPIClient{}

	_, err := client.ListSessionsForRepo(context.Background(), "", "only-repo", 0)
	assert.EqualError(t, err, "owner and repo are required")
	_, err = client.ListSessionsForRepo(context.Background(), "only-owner", "", 0)
	assert.EqualError(t, err, "owner and repo are required")
	_, err = client.ListSessionsForRepo(context.Background(), "", "", 0)
	assert.EqualError(t, err, "owner and repo are required")
}

func TestListSessionsForRepo(t *testing.T) {
	sampleDateString := "2025-08-29T00:00:00Z"
	sampleDate, err := time.Parse(time.RFC3339, sampleDateString)
	require.NoError(t, err)

	tests := []struct {
		name      string
		perPage   int
		limit     int
		httpStubs func(*testing.T, *httpmock.Registry)
		wantErr   string
		wantOut   []*Session
	}{
		{
			name:    "zero limit",
			limit:   0,
			wantOut: nil,
		},
		{
			name:  "no sessions",
			limit: 10,
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions/nwo/OWNER/REPO", url.Values{
							"page_number": {"1"},
							"page_size":   {"50"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StringResponse(`{"sessions":[]}`),
				)
			},
			wantOut: nil,
		},
		{
			name:  "single session",
			limit: 10,
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions/nwo/OWNER/REPO", url.Values{
							"page_number": {"1"},
							"page_size":   {"50"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StringResponse(heredoc.Docf(`
						{
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
						}`,
						sampleDateString,
					)),
				)
				// GraphQL hydration
				reg.Register(
					httpmock.GraphQL(`query FetchPRsAndUsersForAgentTaskSessions\b`),
					httpmock.GraphQLQuery(heredoc.Docf(`
						{
							"data": {
								"nodes": [
									{
										"__typename": "PullRequest",
										"id": "PR_node",
										"fullDatabaseId": "2000",
										"number": 42,
										"title": "Improve docs",
										"state": "OPEN",
										"isDraft": true,
										"url": "https://github.com/OWNER/REPO/pull/42",
										"body": "",
										"createdAt": "%[1]s",
										"updatedAt": "%[1]s",
										"repository": {
											"nameWithOwner": "OWNER/REPO"
										}
									},
									{
										"__typename": "User",
										"login": "octocat",
										"name": "Octocat",
										"databaseId": 1
									}
								]
							}
						}`,
						sampleDateString,
					), func(q string, vars map[string]interface{}) {
						assert.Equal(t, []interface{}{"PR_kwDNA-jNB9A", "U_kgAB"}, vars["ids"])
					}),
				)
			},
			wantOut: []*Session{
				{
					ID:           "sess1",
					Name:         "Build artifacts",
					UserID:       1,
					AgentID:      2,
					Logs:         "",
					State:        "completed",
					OwnerID:      10,
					RepoID:       1000,
					ResourceType: "pull",
					ResourceID:   2000,
					CreatedAt:    sampleDate,
					PullRequest: &api.PullRequest{
						ID:             "PR_node",
						FullDatabaseID: "2000",
						Number:         42,
						Title:          "Improve docs",
						State:          "OPEN",
						IsDraft:        true,
						URL:            "https://github.com/OWNER/REPO/pull/42",
						Body:           "",
						CreatedAt:      sampleDate,
						UpdatedAt:      sampleDate,
						Repository: &api.PRRepository{
							NameWithOwner: "OWNER/REPO",
						},
					},
					User: &api.GitHubUser{
						Login:      "octocat",
						Name:       "Octocat",
						DatabaseID: 1,
					},
				},
			},
		},
		{
			// This happens at the early moments of a session lifecycle, before a PR is created and associated with it.
			name:  "single session, no pull request resource",
			limit: 10,
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions/nwo/OWNER/REPO", url.Values{
							"page_number": {"1"},
							"page_size":   {"50"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StringResponse(heredoc.Docf(`
						{
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
									"resource_type": "",
									"resource_id": 0,
									"created_at": "%[1]s"
								}
							]
						}`,
						sampleDateString,
					)),
				)
				// GraphQL hydration
				reg.Register(
					httpmock.GraphQL(`query FetchPRsAndUsersForAgentTaskSessions\b`),
					httpmock.GraphQLQuery(heredoc.Docf(`
						{
							"data": {
								"nodes": [
									{
										"__typename": "User",
										"login": "octocat",
										"name": "Octocat",
										"databaseId": 1
									}
								]
							}
						}`,
						sampleDateString,
					), func(q string, vars map[string]interface{}) {
						assert.Equal(t, []interface{}{"U_kgAB"}, vars["ids"])
					}),
				)
			},
			wantOut: []*Session{
				{

					ID:           "sess1",
					Name:         "Build artifacts",
					UserID:       1,
					AgentID:      2,
					Logs:         "",
					State:        "completed",
					OwnerID:      10,
					RepoID:       1000,
					ResourceType: "",
					ResourceID:   0,
					CreatedAt:    sampleDate,
					User: &api.GitHubUser{
						Login:      "octocat",
						Name:       "Octocat",
						DatabaseID: 1,
					},
				},
			},
		},
		{
			name:    "multiple sessions, paginated",
			perPage: 1, // to enforce pagination
			limit:   2,
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions/nwo/OWNER/REPO", url.Values{
							"page_number": {"1"},
							"page_size":   {"1"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StringResponse(heredoc.Docf(`
						{
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
						}`,
						sampleDateString,
					)),
				)

				// Second page
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions/nwo/OWNER/REPO", url.Values{
							"page_number": {"2"},
							"page_size":   {"1"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StringResponse(heredoc.Docf(`
						{
							"sessions": [
								{
									"id": "sess2",
									"name": "Build artifacts",
									"user_id": 1,
									"agent_id": 2,
									"logs": "",
									"state": "completed",
									"owner_id": 10,
									"repo_id": 1000,
									"resource_type": "pull",
									"resource_id": 2001,
									"created_at": "%[1]s"
								}
							]
						}`,
						sampleDateString,
					)),
				)
				// GraphQL hydration
				reg.Register(
					httpmock.GraphQL(`query FetchPRsAndUsersForAgentTaskSessions\b`),
					httpmock.GraphQLQuery(heredoc.Docf(`
						{
							"data": {
								"nodes": [
									{
										"__typename": "PullRequest",
										"id": "PR_node",
										"fullDatabaseId": "2000",
										"number": 42,
										"title": "Improve docs",
										"state": "OPEN",
										"isDraft": true,
										"url": "https://github.com/OWNER/REPO/pull/42",
										"body": "",
										"createdAt": "%[1]s",
										"updatedAt": "%[1]s",
										"repository": {
											"nameWithOwner": "OWNER/REPO"
										}
									},
									{
										"__typename": "PullRequest",
										"id": "PR_node",
										"fullDatabaseId": "2001",
										"number": 43,
										"title": "Improve docs",
										"state": "OPEN",
										"isDraft": true,
										"url": "https://github.com/OWNER/REPO/pull/43",
										"body": "",
										"createdAt": "%[1]s",
										"updatedAt": "%[1]s",
										"repository": {
											"nameWithOwner": "OWNER/REPO"
										}
									},
									{
										"__typename": "User",
										"login": "octocat",
										"name": "Octocat",
										"databaseId": 1
									}
								]
							}
						}`,
						sampleDateString,
					), func(q string, vars map[string]interface{}) {
						assert.Equal(t, []interface{}{"PR_kwDNA-jNB9A", "PR_kwDNA-jNB9E", "U_kgAB"}, vars["ids"])
					}),
				)
			},
			wantOut: []*Session{
				{
					ID:           "sess1",
					Name:         "Build artifacts",
					UserID:       1,
					AgentID:      2,
					Logs:         "",
					State:        "completed",
					OwnerID:      10,
					RepoID:       1000,
					ResourceType: "pull",
					ResourceID:   2000,
					CreatedAt:    sampleDate,
					PullRequest: &api.PullRequest{
						ID:             "PR_node",
						FullDatabaseID: "2000",
						Number:         42,
						Title:          "Improve docs",
						State:          "OPEN",
						IsDraft:        true,
						URL:            "https://github.com/OWNER/REPO/pull/42",
						Body:           "",
						CreatedAt:      sampleDate,
						UpdatedAt:      sampleDate,
						Repository: &api.PRRepository{
							NameWithOwner: "OWNER/REPO",
						},
					},
					User: &api.GitHubUser{
						Login:      "octocat",
						Name:       "Octocat",
						DatabaseID: 1,
					},
				},
				{
					ID:           "sess2",
					Name:         "Build artifacts",
					UserID:       1,
					AgentID:      2,
					Logs:         "",
					State:        "completed",
					OwnerID:      10,
					RepoID:       1000,
					ResourceType: "pull",
					ResourceID:   2001,
					CreatedAt:    sampleDate,
					PullRequest: &api.PullRequest{
						ID:             "PR_node",
						FullDatabaseID: "2001",
						Number:         43,
						Title:          "Improve docs",
						State:          "OPEN",
						IsDraft:        true,
						URL:            "https://github.com/OWNER/REPO/pull/43",
						Body:           "",
						CreatedAt:      sampleDate,
						UpdatedAt:      sampleDate,
						Repository: &api.PRRepository{
							NameWithOwner: "OWNER/REPO",
						},
					},
					User: &api.GitHubUser{
						Login:      "octocat",
						Name:       "Octocat",
						DatabaseID: 1,
					},
				},
			},
		},
		{
			name:  "API error",
			limit: 10,
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions/nwo/OWNER/REPO", url.Values{
							"page_number": {"1"},
							"page_size":   {"50"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StatusStringResponse(500, "{}"),
				)
			},
			wantErr: "failed to list sessions:",
		}, {
			name:  "API error at hydration",
			limit: 10,
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions/nwo/OWNER/REPO", url.Values{
							"page_number": {"1"},
							"page_size":   {"50"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StringResponse(heredoc.Docf(`
						{
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
						}`,
						sampleDateString,
					)),
				)
				// GraphQL hydration
				reg.Register(
					httpmock.GraphQL(`query FetchPRsAndUsersForAgentTaskSessions\b`),
					httpmock.StatusStringResponse(500, `{}`),
				)
			},
			wantErr: `failed to fetch session resources: non-200 OK status code:`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &httpmock.Registry{}
			if tt.httpStubs != nil {
				tt.httpStubs(t, reg)
			}
			defer reg.Verify(t)

			httpClient := &http.Client{Transport: reg}

			cfg := config.NewBlankConfig()
			capiClient := NewCAPIClient(httpClient, cfg.Authentication())

			if tt.perPage != 0 {
				last := defaultSessionsPerPage
				defaultSessionsPerPage = tt.perPage
				defer func() {
					defaultSessionsPerPage = last
				}()
			}

			sessions, err := capiClient.ListSessionsForRepo(context.Background(), "OWNER", "REPO", tt.limit)

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				require.Nil(t, sessions)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantOut, sessions)
		})
	}
}

func TestListSessionsByResourceIDRequiresResource(t *testing.T) {
	client := &CAPIClient{}

	_, err := client.ListSessionsByResourceID(context.Background(), "", 999, 0)
	assert.EqualError(t, err, "missing resource type/ID")
	_, err = client.ListSessionsByResourceID(context.Background(), "only-resource-type", 0, 0)
	assert.EqualError(t, err, "missing resource type/ID")
	_, err = client.ListSessionsByResourceID(context.Background(), "", 0, 0)
	assert.EqualError(t, err, "missing resource type/ID")
}

func TestListSessionsByResourceID(t *testing.T) {
	sampleDateString := "2025-08-29T00:00:00Z"
	sampleDate, err := time.Parse(time.RFC3339, sampleDateString)
	require.NoError(t, err)

	resourceID := int64(999)
	resourceType := "pull"

	tests := []struct {
		name      string
		perPage   int
		limit     int
		httpStubs func(*testing.T, *httpmock.Registry)
		wantErr   string
		wantOut   []*Session
	}{
		{
			name:    "zero limit",
			limit:   0,
			wantOut: nil,
		},
		{
			name:  "no sessions",
			limit: 10,
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions/resource/pull/999", url.Values{
							"page_number": {"1"},
							"page_size":   {"50"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StringResponse(`{"sessions":[]}`),
				)
			},
			wantOut: nil,
		},
		{
			name:  "single session",
			limit: 10,
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions/resource/pull/999", url.Values{
							"page_number": {"1"},
							"page_size":   {"50"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StringResponse(heredoc.Docf(`
						{
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
						}`,
						sampleDateString,
					)),
				)
				// GraphQL hydration
				reg.Register(
					httpmock.GraphQL(`query FetchPRsAndUsersForAgentTaskSessions\b`),
					httpmock.GraphQLQuery(heredoc.Docf(`
						{
							"data": {
								"nodes": [
									{
										"__typename": "PullRequest",
										"id": "PR_node",
										"fullDatabaseId": "2000",
										"number": 42,
										"title": "Improve docs",
										"state": "OPEN",
										"isDraft": true,
										"url": "https://github.com/OWNER/REPO/pull/42",
										"body": "",
										"createdAt": "%[1]s",
										"updatedAt": "%[1]s",
										"repository": {
											"nameWithOwner": "OWNER/REPO"
										}
									},
									{
										"__typename": "User",
										"login": "octocat",
										"name": "Octocat",
										"databaseId": 1
									}
								]
							}
						}`,
						sampleDateString,
					), func(q string, vars map[string]interface{}) {
						assert.Equal(t, []interface{}{"PR_kwDNA-jNB9A", "U_kgAB"}, vars["ids"])
					}),
				)
			},
			wantOut: []*Session{
				{

					ID:           "sess1",
					Name:         "Build artifacts",
					UserID:       1,
					AgentID:      2,
					Logs:         "",
					State:        "completed",
					OwnerID:      10,
					RepoID:       1000,
					ResourceType: "pull",
					ResourceID:   2000,
					CreatedAt:    sampleDate,
					PullRequest: &api.PullRequest{
						ID:             "PR_node",
						FullDatabaseID: "2000",
						Number:         42,
						Title:          "Improve docs",
						State:          "OPEN",
						IsDraft:        true,
						URL:            "https://github.com/OWNER/REPO/pull/42",
						Body:           "",
						CreatedAt:      sampleDate,
						UpdatedAt:      sampleDate,
						Repository: &api.PRRepository{
							NameWithOwner: "OWNER/REPO",
						},
					},
					User: &api.GitHubUser{
						Login:      "octocat",
						Name:       "Octocat",
						DatabaseID: 1,
					},
				},
			},
		},
		{
			name:    "multiple sessions, paginated",
			perPage: 1, // to enforce pagination
			limit:   2,
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions/resource/pull/999", url.Values{
							"page_number": {"1"},
							"page_size":   {"1"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StringResponse(heredoc.Docf(`
						{
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
						}`,
						sampleDateString,
					)),
				)

				// Second page
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions/resource/pull/999", url.Values{
							"page_number": {"2"},
							"page_size":   {"1"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StringResponse(heredoc.Docf(`
						{
							"sessions": [
								{
									"id": "sess2",
									"name": "Build artifacts",
									"user_id": 1,
									"agent_id": 2,
									"logs": "",
									"state": "completed",
									"owner_id": 10,
									"repo_id": 1000,
									"resource_type": "pull",
									"resource_id": 2001,
									"created_at": "%[1]s"
								}
							]
						}`,
						sampleDateString,
					)),
				)
				// GraphQL hydration
				reg.Register(
					httpmock.GraphQL(`query FetchPRsAndUsersForAgentTaskSessions\b`),
					httpmock.GraphQLQuery(heredoc.Docf(`
						{
							"data": {
								"nodes": [
									{
										"__typename": "PullRequest",
										"id": "PR_node",
										"fullDatabaseId": "2000",
										"number": 42,
										"title": "Improve docs",
										"state": "OPEN",
										"isDraft": true,
										"url": "https://github.com/OWNER/REPO/pull/42",
										"body": "",
										"createdAt": "%[1]s",
										"updatedAt": "%[1]s",
										"repository": {
											"nameWithOwner": "OWNER/REPO"
										}
									},
									{
										"__typename": "PullRequest",
										"id": "PR_node",
										"fullDatabaseId": "2001",
										"number": 43,
										"title": "Improve docs",
										"state": "OPEN",
										"isDraft": true,
										"url": "https://github.com/OWNER/REPO/pull/43",
										"body": "",
										"createdAt": "%[1]s",
										"updatedAt": "%[1]s",
										"repository": {
											"nameWithOwner": "OWNER/REPO"
										}
									},
									{
										"__typename": "User",
										"login": "octocat",
										"name": "Octocat",
										"databaseId": 1
									}
								]
							}
						}`,
						sampleDateString,
					), func(q string, vars map[string]interface{}) {
						assert.Equal(t, []interface{}{"PR_kwDNA-jNB9A", "PR_kwDNA-jNB9E", "U_kgAB"}, vars["ids"])
					}),
				)
			},
			wantOut: []*Session{
				{
					ID:           "sess1",
					Name:         "Build artifacts",
					UserID:       1,
					AgentID:      2,
					Logs:         "",
					State:        "completed",
					OwnerID:      10,
					RepoID:       1000,
					ResourceType: "pull",
					ResourceID:   2000,
					CreatedAt:    sampleDate,
					PullRequest: &api.PullRequest{
						ID:             "PR_node",
						FullDatabaseID: "2000",
						Number:         42,
						Title:          "Improve docs",
						State:          "OPEN",
						IsDraft:        true,
						URL:            "https://github.com/OWNER/REPO/pull/42",
						Body:           "",
						CreatedAt:      sampleDate,
						UpdatedAt:      sampleDate,
						Repository: &api.PRRepository{
							NameWithOwner: "OWNER/REPO",
						},
					},
					User: &api.GitHubUser{
						Login:      "octocat",
						Name:       "Octocat",
						DatabaseID: 1,
					},
				},
				{
					ID:           "sess2",
					Name:         "Build artifacts",
					UserID:       1,
					AgentID:      2,
					Logs:         "",
					State:        "completed",
					OwnerID:      10,
					RepoID:       1000,
					ResourceType: "pull",
					ResourceID:   2001,
					CreatedAt:    sampleDate,
					PullRequest: &api.PullRequest{
						ID:             "PR_node",
						FullDatabaseID: "2001",
						Number:         43,
						Title:          "Improve docs",
						State:          "OPEN",
						IsDraft:        true,
						URL:            "https://github.com/OWNER/REPO/pull/43",
						Body:           "",
						CreatedAt:      sampleDate,
						UpdatedAt:      sampleDate,
						Repository: &api.PRRepository{
							NameWithOwner: "OWNER/REPO",
						},
					},
					User: &api.GitHubUser{
						Login:      "octocat",
						Name:       "Octocat",
						DatabaseID: 1,
					},
				},
			},
		},
		{
			name:  "API error",
			limit: 10,
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions/resource/pull/999", url.Values{
							"page_number": {"1"},
							"page_size":   {"50"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StatusStringResponse(500, "{}"),
				)
			},
			wantErr: "failed to list sessions:",
		}, {
			name:  "API error at hydration",
			limit: 10,
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(
						httpmock.QueryMatcher("GET", "agents/sessions/resource/pull/999", url.Values{
							"page_number": {"1"},
							"page_size":   {"50"},
						}),
						"api.githubcopilot.com",
					),
					httpmock.StringResponse(heredoc.Docf(`
						{
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
						}`,
						sampleDateString,
					)),
				)
				// GraphQL hydration
				reg.Register(
					httpmock.GraphQL(`query FetchPRsAndUsersForAgentTaskSessions\b`),
					httpmock.StatusStringResponse(500, `{}`),
				)
			},
			wantErr: `failed to fetch session resources: non-200 OK status code:`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &httpmock.Registry{}
			if tt.httpStubs != nil {
				tt.httpStubs(t, reg)
			}
			defer reg.Verify(t)

			httpClient := &http.Client{Transport: reg}

			cfg := config.NewBlankConfig()
			capiClient := NewCAPIClient(httpClient, cfg.Authentication())

			if tt.perPage != 0 {
				last := defaultSessionsPerPage
				defaultSessionsPerPage = tt.perPage
				defer func() {
					defaultSessionsPerPage = last
				}()
			}

			sessions, err := capiClient.ListSessionsByResourceID(context.Background(), resourceType, resourceID, tt.limit)

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				require.Nil(t, sessions)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantOut, sessions)
		})
	}
}

func TestGetSessionRequiresID(t *testing.T) {
	client := &CAPIClient{}

	_, err := client.GetSession(context.Background(), "")
	assert.EqualError(t, err, "missing session ID")
}

func TestGetSession(t *testing.T) {
	sampleDateString := "2025-08-29T00:00:00Z"
	sampleDate, err := time.Parse(time.RFC3339, sampleDateString)
	require.NoError(t, err)

	tests := []struct {
		name      string
		httpStubs func(*testing.T, *httpmock.Registry)
		wantErr   string
		wantErrIs error
		wantOut   *Session
	}{
		{
			name: "session not found",
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(httpmock.REST("GET", "agents/sessions/some-uuid"), "api.githubcopilot.com"),
					httpmock.StatusStringResponse(404, "{}"),
				)
			},
			wantErrIs: ErrSessionNotFound,
			wantErr:   "not found",
		},
		{
			name: "API error",
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(httpmock.REST("GET", "agents/sessions/some-uuid"), "api.githubcopilot.com"),
					httpmock.StatusStringResponse(500, "some error"),
				)
			},
			wantErr: "failed to get session:",
		},
		{
			name: "invalid JSON response",
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(httpmock.REST("GET", "agents/sessions/some-uuid"), "api.githubcopilot.com"),
					httpmock.StatusStringResponse(200, ""),
				)
			},
			wantErr: "failed to decode session response: EOF",
		},
		{
			name: "success",
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(httpmock.REST("GET", "agents/sessions/some-uuid"), "api.githubcopilot.com"),
					httpmock.StringResponse(heredoc.Docf(`
						{
							"id": "some-uuid",
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
						}`,
						sampleDateString,
					)),
				)
				// GraphQL hydration
				reg.Register(
					httpmock.GraphQL(`query FetchPRsAndUsersForAgentTaskSessions\b`),
					httpmock.GraphQLQuery(heredoc.Docf(`
						{
							"data": {
								"nodes": [
									{
										"__typename": "PullRequest",
										"id": "PR_node",
										"fullDatabaseId": "2000",
										"number": 42,
										"title": "Improve docs",
										"state": "OPEN",
										"isDraft": true,
										"url": "https://github.com/OWNER/REPO/pull/42",
										"body": "",
										"createdAt": "%[1]s",
										"updatedAt": "%[1]s",
										"repository": {
											"nameWithOwner": "OWNER/REPO"
										}
									},
									{
										"__typename": "User",
										"login": "octocat",
										"name": "Octocat",
										"databaseId": 1
									}
								]
							}
						}`,
						sampleDateString,
					), func(q string, vars map[string]interface{}) {
						assert.Equal(t, []interface{}{"PR_kwDNA-jNB9A", "U_kgAB"}, vars["ids"])
					}),
				)
			},
			wantOut: &Session{
				ID:           "some-uuid",
				Name:         "Build artifacts",
				UserID:       1,
				AgentID:      2,
				Logs:         "",
				State:        "completed",
				OwnerID:      10,
				RepoID:       1000,
				ResourceType: "pull",
				ResourceID:   2000,
				CreatedAt:    sampleDate,
				PullRequest: &api.PullRequest{
					ID:             "PR_node",
					FullDatabaseID: "2000",
					Number:         42,
					Title:          "Improve docs",
					State:          "OPEN",
					IsDraft:        true,
					URL:            "https://github.com/OWNER/REPO/pull/42",
					Body:           "",
					CreatedAt:      sampleDate,
					UpdatedAt:      sampleDate,
					Repository: &api.PRRepository{
						NameWithOwner: "OWNER/REPO",
					},
				},
				User: &api.GitHubUser{
					Login:      "octocat",
					Name:       "Octocat",
					DatabaseID: 1,
				},
			},
		},
		{
			// This happens at the early moments of a session lifecycle, before a PR is created and associated with it.
			name: "success, but no pull request resource",
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(httpmock.REST("GET", "agents/sessions/some-uuid"), "api.githubcopilot.com"),
					httpmock.StringResponse(heredoc.Docf(`
						{
							"id": "some-uuid",
							"name": "Build artifacts",
							"user_id": 1,
							"agent_id": 2,
							"logs": "",
							"state": "completed",
							"owner_id": 10,
							"repo_id": 1000,
							"resource_type": "",
							"resource_id": 0,
							"created_at": "%[1]s"
						}`,
						sampleDateString,
					)),
				)
				// GraphQL hydration
				reg.Register(
					httpmock.GraphQL(`query FetchPRsAndUsersForAgentTaskSessions\b`),
					httpmock.GraphQLQuery(heredoc.Docf(`
						{
							"data": {
								"nodes": [
									{
										"__typename": "User",
										"login": "octocat",
										"name": "Octocat",
										"databaseId": 1
									}
								]
							}
						}`,
						sampleDateString,
					), func(q string, vars map[string]interface{}) {
						assert.Equal(t, []interface{}{"U_kgAB"}, vars["ids"])
					}),
				)
			},
			wantOut: &Session{
				ID:           "some-uuid",
				Name:         "Build artifacts",
				UserID:       1,
				AgentID:      2,
				Logs:         "",
				State:        "completed",
				OwnerID:      10,
				RepoID:       1000,
				ResourceType: "",
				ResourceID:   0,
				CreatedAt:    sampleDate,
				User: &api.GitHubUser{
					Login:      "octocat",
					Name:       "Octocat",
					DatabaseID: 1,
				},
			},
		},
		{
			name: "API error at hydration",
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(httpmock.REST("GET", "agents/sessions/some-uuid"), "api.githubcopilot.com"),
					httpmock.StringResponse(heredoc.Docf(`
						{
							"id": "some-uuid",
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
						}`,
						sampleDateString,
					)),
				)
				// GraphQL hydration
				reg.Register(
					httpmock.GraphQL(`query FetchPRsAndUsersForAgentTaskSessions\b`),
					httpmock.StatusStringResponse(500, `{}`),
				)
			},
			wantErr: `failed to fetch session resources: non-200 OK status code:`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &httpmock.Registry{}
			if tt.httpStubs != nil {
				tt.httpStubs(t, reg)
			}
			defer reg.Verify(t)

			httpClient := &http.Client{Transport: reg}

			cfg := config.NewBlankConfig()
			capiClient := NewCAPIClient(httpClient, cfg.Authentication())

			session, err := capiClient.GetSession(context.Background(), "some-uuid")

			if tt.wantErrIs != nil {
				require.ErrorIs(t, err, tt.wantErrIs)
			}

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				require.Nil(t, session)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantOut, session)
		})
	}
}
func TestGetPullRequestDatabaseID(t *testing.T) {
	tests := []struct {
		name      string
		httpStubs func(*testing.T, *httpmock.Registry)
		wantErr   string
		wantOut   int64
	}{
		{
			name: "graphql error",
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(httpmock.GraphQL(`query GetPullRequestFullDatabaseID\b`), "api.github.com"),
					httpmock.StringResponse(`{"data":{}, "errors": [{"message": "some gql error"}]}`),
				)
			},
			wantErr: "some gql error",
		},
		{
			// This never happens in practice and it's just to cover more code path
			name: "non-int database ID",
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(httpmock.GraphQL(`query GetPullRequestFullDatabaseID\b`), "api.github.com"),
					httpmock.StringResponse(`{"data": {"repository": {"pullRequest": {"fullDatabaseId": "non-int"}}}}`),
				)
			},
			wantErr: `strconv.ParseInt: parsing "non-int": invalid syntax`,
		},
		{
			name: "success",
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.WithHost(httpmock.GraphQL(`query GetPullRequestFullDatabaseID\b`), "api.github.com"),
					httpmock.GraphQLQuery(`{"data": {"repository": {"pullRequest": {"fullDatabaseId": "999"}}}}`, func(s string, m map[string]interface{}) {
						assert.Equal(t, "OWNER", m["owner"])
						assert.Equal(t, "REPO", m["repo"])
						assert.Equal(t, float64(42), m["number"])
					}),
				)
			},
			wantOut: 999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &httpmock.Registry{}
			if tt.httpStubs != nil {
				tt.httpStubs(t, reg)
			}
			defer reg.Verify(t)

			httpClient := &http.Client{Transport: reg}

			cfg := config.NewBlankConfig()
			capiClient := NewCAPIClient(httpClient, cfg.Authentication())

			databaseID, err := capiClient.GetPullRequestDatabaseID(context.Background(), "github.com", "OWNER", "REPO", 42)

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				require.Zero(t, databaseID)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantOut, databaseID)
		})
	}
}
