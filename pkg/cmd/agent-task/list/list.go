package list

import (
	"context"
	"fmt"
	"time"

	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/internal/tableprinter"
	"github.com/cli/cli/v2/pkg/cmd/agent-task/capi"
	"github.com/cli/cli/v2/pkg/cmd/agent-task/shared"
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
	CapiClient func() (*capi.CAPIClient, error)
	BaseRepo   func() (ghrepo.Interface, error)
}

// NewCmdList creates the list command
func NewCmdList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{
		IO:     f.IOStreams,
		Config: f.Config,
		Limit:  defaultLimit,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List agent tasks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Support -R/--repo override
			if f != nil {
				opts.BaseRepo = f.BaseRepo
			}
			if runF != nil {
				return runF(opts)
			}
			return listRun(opts)
		},
	}

	if f != nil {
		cmdutil.EnableRepoOverride(cmd, f)
	}

	opts.CapiClient = func() (*capi.CAPIClient, error) {
		cfg, err := opts.Config()
		if err != nil {
			return nil, err
		}
		httpClient, err := f.HttpClient()
		if err != nil {
			return nil, err
		}
		authCfg := cfg.Authentication()
		return capi.NewCAPIClient(httpClient, authCfg), nil
	}

	return cmd
}

func listRun(opts *ListOptions) error {
	if opts.Limit <= 0 {
		opts.Limit = defaultLimit
	}

	capiClient, err := opts.CapiClient()
	if err != nil {
		return err
	}

	opts.IO.StartProgressIndicatorWithLabel("Fetching agent tasks...")
	defer opts.IO.StopProgressIndicator()
	var sessions []*capi.Session
	ctx := context.Background()

	var repo ghrepo.Interface
	if opts.BaseRepo != nil {
		repo, _ = opts.BaseRepo()
	}

	if repo != nil && repo.RepoOwner() != "" && repo.RepoName() != "" {
		sessions, err = capiClient.ListSessionsForRepo(ctx, repo.RepoOwner(), repo.RepoName(), opts.Limit)
		if err != nil {
			return err
		}
	} else {
		sessions, err = capiClient.ListSessionsForViewer(ctx, opts.Limit)
		if err != nil {
			return err
		}
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
			tp.AddField(s.State, tableprinter.WithColor(shared.ColorFuncForSessionState(*s, cs)))
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
