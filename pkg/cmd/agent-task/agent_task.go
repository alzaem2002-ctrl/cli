package agent

import (
	"errors"
	"fmt"
	"strings"

	cmdCreate "github.com/cli/cli/v2/pkg/cmd/agent-task/create"
	cmdList "github.com/cli/cli/v2/pkg/cmd/agent-task/list"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/spf13/cobra"
)

// NewCmdAgentTask creates the base `agent-task` command.
func NewCmdAgentTask(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agent-task",
		Aliases: []string{"agent-tasks", "agent", "agents"},
		Short:   "Manage agent tasks (preview)",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return requireOAuthToken(f)
		},
		// This is required to run this root command. We want to
		// run it to test PersistentPreRunE behavior.
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// register subcommands
	cmd.AddCommand(cmdList.NewCmdList(f, nil))
	cmd.AddCommand(cmdCreate.NewCmdCreate(f, nil))

	return cmd
}

// requireOAuthToken ensures an OAuth (device flow) token is present and valid.
// agent-task subcommands inherit this check via PersistentPreRunE.
func requireOAuthToken(f *cmdutil.Factory) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	authCfg := cfg.Authentication()
	host, _ := authCfg.DefaultHost()
	if host == "" {
		return errors.New("no default host configured; run 'gh auth login'")
	}

	if auth.IsEnterprise(host) {
		return errors.New("agent tasks are not supported on this host")
	}

	token, source := authCfg.ActiveToken(host)

	// Tokens from sources "oauth_token" and "keyring" are likely
	// minted through our device flow.
	tokenSourceIsDeviceFlow := source == "oauth_token" || source == "keyring"
	// Tokens with "gho_" prefix are OAuth tokens.
	tokenIsOAuth := strings.HasPrefix(token, "gho_")

	// Reject if the token is not from a device flow source or is not an OAuth token
	if !tokenSourceIsDeviceFlow || !tokenIsOAuth {
		return fmt.Errorf("this command requires an OAuth token. Re-authenticate with: gh auth login")
	}
	return nil
}
