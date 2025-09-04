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
					httpmock.GraphQL(`query FetchPRsForAgentTaskSessions\b`),
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
						}`,
						sampleDateString,
					), func(q string, vars map[string]interface{}) {
						assert.Equal(t, []interface{}{"PR_kwDNA-jNB9A"}, vars["ids"])
					}),
				)
			},
			wantOut: []*Session{
				{
					session: session{
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
					},
					PullRequest: &api.PullRequest{
						ID:             "PR_node",
						FullDatabaseID: "2000",
						Number:         42,
						Title:          "Improve docs",
						State:          "OPEN",
						URL:            "https://github.com/OWNER/REPO/pull/42",
						Body:           "",
						CreatedAt:      sampleDate,
						UpdatedAt:      sampleDate,
						Repository: &api.PRRepository{
							NameWithOwner: "OWNER/REPO",
						},
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
					httpmock.GraphQL(`query FetchPRsForAgentTaskSessions\b`),
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
										"url": "https://github.com/OWNER/REPO/pull/43",
										"body": "",
										"createdAt": "%[1]s",
										"updatedAt": "%[1]s",
										"repository": {
											"nameWithOwner": "OWNER/REPO"
										}
									}
								]
							}
						}`,
						sampleDateString,
					), func(q string, vars map[string]interface{}) {
						assert.Equal(t, []interface{}{"PR_kwDNA-jNB9A", "PR_kwDNA-jNB9E"}, vars["ids"])
					}),
				)
			},
			wantOut: []*Session{
				{
					session: session{
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
					},
					PullRequest: &api.PullRequest{
						ID:             "PR_node",
						FullDatabaseID: "2000",
						Number:         42,
						Title:          "Improve docs",
						State:          "OPEN",
						URL:            "https://github.com/OWNER/REPO/pull/42",
						Body:           "",
						CreatedAt:      sampleDate,
						UpdatedAt:      sampleDate,
						Repository: &api.PRRepository{
							NameWithOwner: "OWNER/REPO",
						},
					},
				},
				{
					session: session{
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
					},
					PullRequest: &api.PullRequest{
						ID:             "PR_node",
						FullDatabaseID: "2001",
						Number:         43,
						Title:          "Improve docs",
						State:          "OPEN",
						URL:            "https://github.com/OWNER/REPO/pull/43",
						Body:           "",
						CreatedAt:      sampleDate,
						UpdatedAt:      sampleDate,
						Repository: &api.PRRepository{
							NameWithOwner: "OWNER/REPO",
						},
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
					httpmock.GraphQL(`query FetchPRsForAgentTaskSessions\b`),
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
					httpmock.GraphQL(`query FetchPRsForAgentTaskSessions\b`),
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
						}`,
						sampleDateString,
					), func(q string, vars map[string]interface{}) {
						assert.Equal(t, []interface{}{"PR_kwDNA-jNB9A"}, vars["ids"])
					}),
				)
			},
			wantOut: []*Session{
				{
					session: session{
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
					},
					PullRequest: &api.PullRequest{
						ID:             "PR_node",
						FullDatabaseID: "2000",
						Number:         42,
						Title:          "Improve docs",
						State:          "OPEN",
						URL:            "https://github.com/OWNER/REPO/pull/42",
						Body:           "",
						CreatedAt:      sampleDate,
						UpdatedAt:      sampleDate,
						Repository: &api.PRRepository{
							NameWithOwner: "OWNER/REPO",
						},
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
					httpmock.GraphQL(`query FetchPRsForAgentTaskSessions\b`),
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
										"url": "https://github.com/OWNER/REPO/pull/43",
										"body": "",
										"createdAt": "%[1]s",
										"updatedAt": "%[1]s",
										"repository": {
											"nameWithOwner": "OWNER/REPO"
										}
									}
								]
							}
						}`,
						sampleDateString,
					), func(q string, vars map[string]interface{}) {
						assert.Equal(t, []interface{}{"PR_kwDNA-jNB9A", "PR_kwDNA-jNB9E"}, vars["ids"])
					}),
				)
			},
			wantOut: []*Session{
				{
					session: session{
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
					},
					PullRequest: &api.PullRequest{
						ID:             "PR_node",
						FullDatabaseID: "2000",
						Number:         42,
						Title:          "Improve docs",
						State:          "OPEN",
						URL:            "https://github.com/OWNER/REPO/pull/42",
						Body:           "",
						CreatedAt:      sampleDate,
						UpdatedAt:      sampleDate,
						Repository: &api.PRRepository{
							NameWithOwner: "OWNER/REPO",
						},
					},
				},
				{
					session: session{
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
					},
					PullRequest: &api.PullRequest{
						ID:             "PR_node",
						FullDatabaseID: "2001",
						Number:         43,
						Title:          "Improve docs",
						State:          "OPEN",
						URL:            "https://github.com/OWNER/REPO/pull/43",
						Body:           "",
						CreatedAt:      sampleDate,
						UpdatedAt:      sampleDate,
						Repository: &api.PRRepository{
							NameWithOwner: "OWNER/REPO",
						},
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
					httpmock.GraphQL(`query FetchPRsForAgentTaskSessions\b`),
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
