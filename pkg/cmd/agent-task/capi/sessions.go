package capi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/cli/cli/v2/api"
	"github.com/vmihailenco/msgpack/v5"
)

// session is an in-flight agent task
type session struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	UserID        uint64    `json:"user_id"`
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

	CreatedAt time.Time
	UpdatedAt time.Time
	ClosedAt  *time.Time
	MergedAt  *time.Time

	Repository *api.PRRepository
}

// Session is a hydrated in-flight agent task
type Session struct {
	session
	PullRequest *api.PullRequest `json:"-"`
}

// ListSessionsForViewer lists all agent sessions for the
// authenticated user up to limit.
func (c *CAPIClient) ListSessionsForViewer(ctx context.Context, limit int) ([]*Session, error) {
	url := baseCAPIURL + "/agents/sessions"

	var sessions []session
	page := 1
	perPage := 50

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Set("page_size", strconv.Itoa(perPage))
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
		if len(response.Sessions) == 0 || len(sessions) >= limit {
			break
		}
		sessions = append(sessions, response.Sessions...)
		page++
	}

	// Drop any above the limit
	if len(sessions) > limit {
		sessions = sessions[:limit]
	}

	// Hydrate the result with pull request data.
	result, err := c.hydrateSessionPullRequests(sessions)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// hydrateSessionPullRequests hydrates pull request information in sessions
func (c *CAPIClient) hydrateSessionPullRequests(sessions []session) ([]*Session, error) {
	if len(sessions) == 0 {
		return nil, nil
	}

	prNodeIds := make([]string, 0, len(sessions))

	for _, session := range sessions {
		prNodeID := generatePullRequestNodeID(int64(session.RepoID), session.ResourceID)
		if slices.Contains(prNodeIds, prNodeID) {
			continue
		}
		prNodeIds = append(prNodeIds, prNodeID)
	}

	apiClient := api.NewClientFromHTTP(c.httpClient)

	var resp struct {
		Nodes []struct {
			PullRequest sessionPullRequest `graphql:"... on PullRequest"`
		} `graphql:"nodes(ids: $ids)"`
	}

	host, _ := c.authCfg.DefaultHost()
	err := apiClient.Query(host, "FetchPRsForAgentTaskSessions", &resp, map[string]any{
		"ids": prNodeIds,
	})

	if err != nil {
		return nil, err
	}

	prs := make([]*api.PullRequest, 0, len(prNodeIds))
	for _, node := range resp.Nodes {
		prs = append(prs, &api.PullRequest{
			ID:             node.PullRequest.ID,
			FullDatabaseID: node.PullRequest.FullDatabaseID,
			Number:         node.PullRequest.Number,
			Title:          node.PullRequest.Title,
			State:          node.PullRequest.State,
			URL:            node.PullRequest.URL,
			Body:           node.PullRequest.Body,
			CreatedAt:      node.PullRequest.CreatedAt,
			UpdatedAt:      node.PullRequest.UpdatedAt,
			ClosedAt:       node.PullRequest.ClosedAt,
			MergedAt:       node.PullRequest.MergedAt,
			Repository:     node.PullRequest.Repository,
		})
	}

	newSessions := make([]*Session, 0, len(sessions))
	// For each session, we need to attach the Pull Request
	for _, s := range sessions {
		// For each Pull Request, check if it matches the session
		for _, pr := range prs {
			if strconv.FormatInt(s.ResourceID, 10) == pr.FullDatabaseID {
				newSessions = append(newSessions, &Session{
					session:     s,
					PullRequest: pr,
				})
			}
		}
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
