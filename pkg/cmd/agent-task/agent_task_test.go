package agent

import (
	"testing"

	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/internal/gh"
	ghmock "github.com/cli/cli/v2/internal/gh/mock"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

// setupMockOAuthConfig configures a blank config with a default host and optional token behavior.
func setupMockOAuthConfig(t *testing.T, tokenSource string) gh.Config {
	t.Helper()
	c := config.NewBlankConfig()
	switch tokenSource {
	case "oauth_token":
		// valid OAuth device flow token stored in config
		c.Set("github.com", "oauth_token", "gho_OAUTH123")
	case "keyring":
		// valid OAuth device flow token stored in keyring
		c.Set("github.com", "oauth_token", "gho_OAUTH123")
	case "GH_TOKEN":
		// classic style token stored in config (will fail prefix check)
		c.Set("github.com", "oauth_token", "ghp_CLASSIC123")
	case "GH_ENTERPRISE_TOKEN":
		// enterprise style token stored in config (will fail prefix check)
		c.Set("something.ghes.com", "oauth_token", "ghe_ENTERPRISE123")
	}
	return c
}

func TestOAuthTokenAccepted(t *testing.T) {
	f := &cmdutil.Factory{}
	ios, _, stdout, _ := iostreams.Test()
	f.IOStreams = ios
	f.Config = func() (gh.Config, error) { return setupMockOAuthConfig(t, "oauth_token"), nil }

	cmd := NewCmdAgentTask(f)
	err := cmd.Execute()
	require.NoError(t, err)
	require.Equal(t, "", stdout.String())
}

func TestKeyringOAuthTokenAccepted(t *testing.T) {
	f := &cmdutil.Factory{}
	ios, _, stdout, _ := iostreams.Test()
	f.IOStreams = ios
	f.Config = func() (gh.Config, error) { return setupMockOAuthConfig(t, "keyring"), nil }

	cmd := NewCmdAgentTask(f)
	err := cmd.Execute()
	require.NoError(t, err)
	require.Equal(t, "", stdout.String())
}

func TestEnvVarTokenRejected(t *testing.T) {
	f := &cmdutil.Factory{}
	ios, _, _, _ := iostreams.Test()
	f.IOStreams = ios
	f.Config = func() (gh.Config, error) { return setupMockOAuthConfig(t, "GH_TOKEN"), nil }
	cmd := NewCmdAgentTask(f)
	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "requires an OAuth token")
}

func TestEnterpriseTokenIgnored(t *testing.T) {
	// This test ignores the test helper because we want to test a specific config state
	t.Run("enterprise token alone is ignored and rejected", func(t *testing.T) {
		f := &cmdutil.Factory{}
		ios, _, _, _ := iostreams.Test()
		f.IOStreams = ios
		f.Config = func() (gh.Config, error) {
			return func() gh.Config {
				c := config.NewBlankConfig()
				c.Set("something.ghes.com", "oauth_token", "ghe_ENTERPRISE123")
				return c
			}(), nil
		}

		cmd := NewCmdAgentTask(f)
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("github.com oauth is accepted and enterprise token ignored", func(t *testing.T) {
		f := &cmdutil.Factory{}
		ios, _, _, _ := iostreams.Test()
		f.IOStreams = ios
		f.Config = func() (gh.Config, error) {
			return func() gh.Config {
				c := config.NewBlankConfig()
				c.Set("something.ghes.com", "oauth_token", "ghe_ENTERPRISE123")
				c.Set("github.com", "oauth_token", "gho_OAUTH123")
				return c
			}(), nil
		}

		cmd := NewCmdAgentTask(f)
		err := cmd.Execute()
		require.NoError(t, err)
	})

}

func TestEnterpriseHostRejected(t *testing.T) {
	f := &cmdutil.Factory{}
	ios, _, _, _ := iostreams.Test()
	f.IOStreams = ios

	f.Config = func() (gh.Config, error) {
		return &ghmock.ConfigMock{
			AuthenticationFunc: func() gh.AuthConfig {
				c := &config.AuthConfig{}
				c.SetDefaultHost("something.ghes.com", "GH_HOST")
				return c
			},
		}, nil
	}

	cmd := NewCmdAgentTask(f)
	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "not supported on this host")
}

func TestEmptyHostRejected(t *testing.T) {
	f := &cmdutil.Factory{}
	ios, _, _, _ := iostreams.Test()
	f.IOStreams = ios

	f.Config = func() (gh.Config, error) {
		return &ghmock.ConfigMock{
			AuthenticationFunc: func() gh.AuthConfig {
				c := &config.AuthConfig{}
				c.SetDefaultHost("", "GH_HOST")
				return c
			},
		}, nil
	}

	cmd := NewCmdAgentTask(f)
	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "no default host configured")
}

func TestNoAuthRejected(t *testing.T) {
	f := &cmdutil.Factory{}
	ios, _, _, _ := iostreams.Test()
	f.IOStreams = ios
	// No token configured
	f.Config = func() (gh.Config, error) { return setupMockOAuthConfig(t, ""), nil }

	cmd := NewCmdAgentTask(f)
	err := cmd.Execute()
	require.Error(t, err)
}

func TestAliasAreSet(t *testing.T) {
	f := &cmdutil.Factory{}
	ios, _, _, _ := iostreams.Test()
	f.IOStreams = ios
	f.Config = func() (gh.Config, error) { return setupMockOAuthConfig(t, "oauth_token"), nil }

	cmd := NewCmdAgentTask(f)

	require.ElementsMatch(t, []string{"agent-tasks", "agent", "agents"}, cmd.Aliases)
}
