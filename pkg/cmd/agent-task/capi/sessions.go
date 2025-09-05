package capi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/cli/cli/v2/api"
	"github.com/vmihailenco/msgpack/v5"
)

var defaultSessionsPerPage = 50

// session is an in-flight agent task
type session struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	UserID        int64     `json:"user_id"`
	AgentID       int64     `json:"agent_id"`
	Logs          string    `json:"logs"`
	State         string    `json:"state"`
	OwnerID       uint64    `json:"owner_id"`
	RepoID        uint64    `json:"repo_id"`
	ResourceType  string    `json:"resource_type"`
	ResourceID    int64     `json:"resource_id"`
	LastUpdatedAt time.Time `json:"last_updated_at,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	CompletedAt   time.Time `json:"completed_at,omitempty"`
	EventURL      string    `json:"event_url"`
	EventType     string    `json:"event_type"`
}

// A shim of a full pull request because looking up by node ID
// using the full api.PullRequest type fails on unions (actors)
type sessionPullRequest struct {
	ID             string
	FullDatabaseID string
	Number         int
	Title          string
	State          string
	URL            string
	Body           string
	IsDraft        bool

	CreatedAt time.Time
	UpdatedAt time.Time
	ClosedAt  *time.Time
	MergedAt  *time.Time

	Repository *api.PRRepository
}

// Session is a hydrated in-flight agent task
type Session struct {
	ID            string
	Name          string
	UserID        int64
	AgentID       int64
	Logs          string
	State         string
	OwnerID       uint64
	RepoID        uint64
	ResourceType  string
	ResourceID    int64
	LastUpdatedAt time.Time
	CreatedAt     time.Time
	CompletedAt   time.Time
	EventURL      string
	EventType     string

	PullRequest *api.PullRequest
	User        *api.GitHubUser
}

// ListSessionsForViewer lists all agent sessions for the
// authenticated user up to limit.
func (c *CAPIClient) ListSessionsForViewer(ctx context.Context, limit int) ([]*Session, error) {
	if limit == 0 {
		return nil, nil
	}

	url := baseCAPIURL + "/agents/sessions"
	pageSize := defaultSessionsPerPage

	sessions := make([]session, 0, limit+pageSize)

	for page := 1; ; page++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Set("page_size", strconv.Itoa(pageSize))
		q.Set("page_number", strconv.Itoa(page))
		req.URL.RawQuery = q.Encode()

		res, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to list sessions: %s", res.Status)
		}
		var response struct {
			Sessions []session `json:"sessions"`
		}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode sessions response: %w", err)
		}

		sessions = append(sessions, response.Sessions...)
		if len(response.Sessions) < pageSize || len(sessions) >= limit {
			break
		}
	}

	// Drop any above the limit
	if len(sessions) > limit {
		sessions = sessions[:limit]
	}

	// Hydrate the result with pull request data.
	result, err := c.hydrateSessionPullRequestsAndUsers(sessions)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch session resources: %w", err)
	}

	return result, nil
}

// ListSessionsForRepo lists agent sessions for a specific repository identified by owner/name up to limit.
func (c *CAPIClient) ListSessionsForRepo(ctx context.Context, owner string, repo string, limit int) ([]*Session, error) {
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}

	if limit == 0 {
		return nil, nil
	}

	url := fmt.Sprintf("%s/agents/sessions/nwo/%s/%s", baseCAPIURL, url.PathEscape(owner), url.PathEscape(repo))
	pageSize := defaultSessionsPerPage

	sessions := make([]session, 0, limit+pageSize)

	for page := 1; ; page++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Set("page_size", strconv.Itoa(pageSize))
		q.Set("page_number", strconv.Itoa(page))
		req.URL.RawQuery = q.Encode()

		res, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to list sessions: %s", res.Status)
		}
		var response struct {
			Sessions []session `json:"sessions"`
		}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode sessions response: %w", err)
		}

		sessions = append(sessions, response.Sessions...)
		if len(response.Sessions) < pageSize || len(sessions) >= limit {
			break
		}
	}

	// Drop any above the limit
	if len(sessions) > limit {
		sessions = sessions[:limit]
	}

	// Hydrate the result with pull request data.
	result, err := c.hydrateSessionPullRequestsAndUsers(sessions)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch session resources: %w", err)
	}
	return result, nil
}

// hydrateSessionPullRequestsAndUsers hydrates pull request and user information in sessions
func (c *CAPIClient) hydrateSessionPullRequestsAndUsers(sessions []session) ([]*Session, error) {
	if len(sessions) == 0 {
		return nil, nil
	}

	prNodeIds := make([]string, 0, len(sessions))
	userNodeIds := make([]string, 0, len(sessions))
	for _, session := range sessions {
		prNodeID := generatePullRequestNodeID(int64(session.RepoID), session.ResourceID)
		if !slices.Contains(prNodeIds, prNodeID) {
			prNodeIds = append(prNodeIds, prNodeID)
		}

		userNodeId := generateUserNodeID(session.UserID)
		if !slices.Contains(userNodeIds, userNodeId) {
			userNodeIds = append(userNodeIds, userNodeId)
		}
	}
	apiClient := api.NewClientFromHTTP(c.httpClient)

	var resp struct {
		Nodes []struct {
			TypeName    string             `graphql:"__typename"`
			PullRequest sessionPullRequest `graphql:"... on PullRequest"`
			User        api.GitHubUser     `graphql:"... on User"`
		} `graphql:"nodes(ids: $ids)"`
	}

	ids := make([]string, 0, len(prNodeIds)+len(userNodeIds))
	ids = append(ids, prNodeIds...)
	ids = append(ids, userNodeIds...)

	// TODO handle pagination
	host, _ := c.authCfg.DefaultHost()
	err := apiClient.Query(host, "FetchPRsAndUsersForAgentTaskSessions", &resp, map[string]any{
		"ids": ids,
	})

	if err != nil {
		return nil, err
	}

	prMap := make(map[string]*api.PullRequest, len(prNodeIds))
	userMap := make(map[int64]*api.GitHubUser, len(userNodeIds))
	for _, node := range resp.Nodes {
		switch node.TypeName {
		case "User":
			userMap[node.User.DatabaseID] = &node.User
		case "PullRequest":
			prMap[node.PullRequest.FullDatabaseID] = &api.PullRequest{
				ID:             node.PullRequest.ID,
				FullDatabaseID: node.PullRequest.FullDatabaseID,
				Number:         node.PullRequest.Number,
				Title:          node.PullRequest.Title,
				State:          node.PullRequest.State,
				IsDraft:        node.PullRequest.IsDraft,
				URL:            node.PullRequest.URL,
				Body:           node.PullRequest.Body,
				CreatedAt:      node.PullRequest.CreatedAt,
				UpdatedAt:      node.PullRequest.UpdatedAt,
				ClosedAt:       node.PullRequest.ClosedAt,
				MergedAt:       node.PullRequest.MergedAt,
				Repository:     node.PullRequest.Repository,
			}
		}
	}

	newSessions := make([]*Session, 0, len(sessions))
	for _, s := range sessions {
		newSession := fromAPISession(s)
		newSession.PullRequest = prMap[strconv.FormatInt(s.ResourceID, 10)]
		newSession.User = userMap[s.UserID]
		newSessions = append(newSessions, newSession)
	}

	return newSessions, nil
}

// generatePullRequestNodeID converts an int64 databaseID and repoID to a GraphQL Node ID format
// with the "PR_" prefix for pull requests
func generatePullRequestNodeID(repoID, pullRequestID int64) string {
	buf := bytes.Buffer{}
	parts := []int64{0, repoID, pullRequestID}

	encoder := msgpack.NewEncoder(&buf)
	encoder.UseCompactInts(true)

	if err := encoder.Encode(parts); err != nil {
		panic(err)
	}

	encoded := base64.RawURLEncoding.EncodeToString(buf.Bytes())

	return "PR_" + encoded
}

func generateUserNodeID(userID int64) string {
	buf := bytes.Buffer{}
	parts := []int64{0, userID}

	encoder := msgpack.NewEncoder(&buf)
	encoder.UseCompactInts(true)

	if err := encoder.Encode(parts); err != nil {
		panic(err)
	}

	encoded := base64.RawURLEncoding.EncodeToString(buf.Bytes())

	return "U_" + encoded
}

func fromAPISession(s session) *Session {
	return &Session{
		ID:            s.ID,
		Name:          s.Name,
		UserID:        s.UserID,
		AgentID:       s.AgentID,
		Logs:          s.Logs,
		State:         s.State,
		OwnerID:       s.OwnerID,
		RepoID:        s.RepoID,
		ResourceType:  s.ResourceType,
		ResourceID:    s.ResourceID,
		LastUpdatedAt: s.LastUpdatedAt,
		CreatedAt:     s.CreatedAt,
		CompletedAt:   s.CompletedAt,
		EventURL:      s.EventURL,
		EventType:     s.EventType,
	}
}
