package list

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/internal/tableprinter"
	"github.com/cli/cli/v2/pkg/cmd/agent-task/capi"
	prShared "github.com/cli/cli/v2/pkg/cmd/pr/shared"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

const defaultLimit = 30

// ListOptions are the options for the list command
type ListOptions struct {
	IO         *iostreams.IOStreams
	Config     func() (gh.Config, error)
	Limit      int
	CapiClient capi.CapiClient
	HttpClient func() (*http.Client, error)
}

// NewCmdList creates the list command
func NewCmdList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{
		IO:         f.IOStreams,
		Config:     f.Config,
		Limit:      defaultLimit,
		HttpClient: f.HttpClient,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List agent tasks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			httpClient, err := opts.HttpClient()
			if err != nil {
				return err
			}

			authCfg := cfg.Authentication()
			opts.CapiClient = capi.NewCAPIClient(httpClient, authCfg)

			if runF != nil {
				return runF(opts)
			}
			return listRun(opts)
		},
	}

	return cmd
}

func listRun(opts *ListOptions) error {
	if opts.Limit <= 0 {
		opts.Limit = defaultLimit
	}

	capiClient := opts.CapiClient

	opts.IO.StartProgressIndicatorWithLabel("Fetching agent tasks...")
	defer opts.IO.StopProgressIndicator()
	sessions, err := capiClient.ListSessionsForViewer(context.Background(), opts.Limit)
	if err != nil {
		return err
	}
	opts.IO.StopProgressIndicator()

	if len(sessions) == 0 {
		fmt.Fprintln(opts.IO.Out, "no agent tasks found")
		return nil
	}

	cs := opts.IO.ColorScheme()
	tp := tableprinter.New(opts.IO, tableprinter.WithHeader("Session ID", "Pull Request", "Repo", "Session State", "Created"))
	for _, s := range sessions {
		if s.ResourceType != "pull" || s.PullRequest == nil || s.PullRequest.Repository == nil {
			// Skip these sessions in case they happen, for now.
			continue
		}

		pr := fmt.Sprintf("#%d", s.PullRequest.Number)
		repo := s.PullRequest.Repository.NameWithOwner

		// ID
		tp.AddField(s.ID)
		if tp.IsTTY() {
			tp.AddField(pr, tableprinter.WithColor(cs.ColorFromString(prShared.ColorForPRState(*s.PullRequest))))
		} else {
			tp.AddField(pr)
		}

		// Repo
		tp.AddField(repo, tableprinter.WithColor(cs.Muted))

		// State
		if tp.IsTTY() {
			var stateColor func(string) string
			switch s.State {
			case "completed":
				stateColor = cs.Green
			case "canceled":
				stateColor = cs.Muted
			case "in_progress", "queued":
				stateColor = cs.Yellow
			case "failed":
				stateColor = cs.Red
			default:
				stateColor = cs.Muted
			}
			tp.AddField(s.State, tableprinter.WithColor(stateColor))
		} else {
			tp.AddField(s.State)
		}

		// Created
		if tp.IsTTY() {
			tp.AddTimeField(time.Now(), s.CreatedAt, cs.Muted)
		} else {
			tp.AddField(s.CreatedAt.Format(time.RFC3339))
		}

		tp.EndRow()
	}

	if err := tp.Render(); err != nil {
		return err
	}

	return nil
}
