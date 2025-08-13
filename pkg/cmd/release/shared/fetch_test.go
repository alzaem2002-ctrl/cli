package shared

import (
	"context"
	"net/http"
	"testing"

	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchRefSHA(t *testing.T) {
	tests := []struct {
		name           string
		tagName        string
		responseStatus int
		responseBody   string
		expectedSHA    string
		errorMessage   string
	}{
		{
			name:           "full semver tag",
			tagName:        "v1.2.3",
			responseStatus: 200,
			responseBody:   `{"object": {"sha": "1234567890abcdef1234567890abcdef12345678"}}`,
			expectedSHA:    "1234567890abcdef1234567890abcdef12345678",
		},
		{
			name:           "partial semver - major only",
			tagName:        "v1",
			responseStatus: 200,
			responseBody:   `{"object": {"sha": "abcdef1234567890abcdef1234567890abcdef12"}}`,
			expectedSHA:    "abcdef1234567890abcdef1234567890abcdef12",
		},
		{
			name:           "partial semver - major.minor",
			tagName:        "v1.2",
			responseStatus: 200,
			responseBody:   `{"object": {"sha": "fedcba0987654321fedcba0987654321fedcba09"}}`,
			expectedSHA:    "fedcba0987654321fedcba0987654321fedcba09",
		},
		{
			name:           "prerelease tag",
			tagName:        "v1.2.3-alpha.1",
			responseStatus: 200,
			responseBody:   `{"object": {"sha": "9876543210fedcba9876543210fedcba98765432"}}`,
			expectedSHA:    "9876543210fedcba9876543210fedcba98765432",
		},
		{
			name:           "tag not found",
			tagName:        "v99.99.99",
			responseStatus: 404,
			responseBody:   ``,
			errorMessage:   "release not found",
		},
		{
			name:           "empty response body with 200 status",
			tagName:        "v1.0.0",
			responseStatus: 200,
			responseBody:   `{}`,
			errorMessage:   "release not found",
		},
		{
			name:           "malformed JSON response",
			tagName:        "v1.0.0",
			responseStatus: 200,
			responseBody:   `{"object": {"sha":`,
			errorMessage:   "failed to parse Git ref response: unexpected EOF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeHTTP := &httpmock.Registry{}
			defer fakeHTTP.Verify(t)

			repo, err := ghrepo.FromFullName("owner/repo")
			require.NoError(t, err)

			path := "repos/owner/repo/git/ref/tags/" + tt.tagName
			if tt.responseStatus == 404 {
				fakeHTTP.Register(httpmock.REST("GET", path), httpmock.StatusStringResponse(404, "Not Found"))
			} else {
				fakeHTTP.Register(httpmock.REST("GET", path), httpmock.StringResponse(tt.responseBody))
			}

			httpClient := &http.Client{Transport: fakeHTTP}
			ctx := context.Background()

			sha, err := FetchRefSHA(ctx, httpClient, repo, tt.tagName)

			if tt.errorMessage != "" {
				require.Error(t, err)
				assert.EqualError(t, err, tt.errorMessage)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedSHA, sha)
			}
		})
	}
}
