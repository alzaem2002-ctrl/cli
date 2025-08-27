package list

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/internal/gh"
	capi "github.com/cli/cli/v2/pkg/cmd/agent-task/capi"
	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

// testListOptionsWithRegistry constructs ListOptions and returns the stdout buffer for assertions
func testListOptionsWithRegistry(reg *httpmock.Registry) (*ListOptions, *bytes.Buffer) {
	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)

	opts := &ListOptions{
		IO: ios,
		HttpClient: func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		},
		Config: func() (gh.Config, error) {
			c := config.NewBlankConfig()
			c.Set("github.com", "oauth_token", "gho_OAUTH123")
			return c, nil
		},
		Limit: defaultLimit,
	}

	return opts, stdout
}

// mockCAPIClient is a small test double for the CAPI client.
type mockCAPIClient struct {
	sessions []*capi.Session
}

// Updated to match production interface which now includes a limit parameter.
func (m *mockCAPIClient) ListSessionsForViewer(ctx context.Context, limit int) ([]*capi.Session, error) {
	return m.sessions, nil
}

func TestListRun_WithSessions(t *testing.T) {
	reg := httpmock.Registry{}
	defer reg.Verify(t)

	opts, stdout := testListOptionsWithRegistry(&reg)

	createdAt := time.Date(2025, time.August, 25, 12, 0, 0, 0, time.UTC)
	s := &capi.Session{}
	s.ID = "s1"
	s.RepoID = 123
	s.ResourceType = "pull"
	s.ResourceID = 456
	s.State = "completed"
	s.CreatedAt = createdAt
	s.PullRequest = &api.PullRequest{
		Number:     456,
		State:      "OPEN",
		Repository: &api.PRRepository{NameWithOwner: "owner/repo"},
	}
	opts.CapiClient = &mockCAPIClient{sessions: []*capi.Session{s}}

	err := listRun(opts)
	require.NoError(t, err)
	out := stdout.String()
	require.Contains(t, out, "SESSION ID")
	require.Contains(t, out, "s1")
	require.Contains(t, out, "#456")
	require.Contains(t, out, "owner/repo")
}

func TestListRun_NoSessions(t *testing.T) {
	reg := httpmock.Registry{}
	defer reg.Verify(t)

	opts, stdout := testListOptionsWithRegistry(&reg)
	opts.CapiClient = &mockCAPIClient{sessions: []*capi.Session{}}

	err := listRun(opts)
	require.NoError(t, err)
	out := stdout.String()
	require.Contains(t, out, "no agent tasks found")
}
