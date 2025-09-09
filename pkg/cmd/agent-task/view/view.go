package view

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/internal/text"
	"github.com/cli/cli/v2/pkg/cmd/agent-task/capi"
	"github.com/cli/cli/v2/pkg/cmd/agent-task/shared"
	prShared "github.com/cli/cli/v2/pkg/cmd/pr/shared"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

type ViewOptions struct {
	IO         *iostreams.IOStreams
	CapiClient func() (capi.CapiClient, error)

	SelectorArg string
}

func NewCmdView(f *cmdutil.Factory, runF func(*ViewOptions) error) *cobra.Command {
	opts := &ViewOptions{
		IO:         f.IOStreams,
		CapiClient: shared.CapiClientFunc(f),
	}

	cmd := &cobra.Command{
		Use:   "view <session-id>",
		Short: "View an agent task session",
		Long: heredoc.Doc(`
			View an agent task session.
		`),
		Args: cmdutil.ExactArgs(1, "a session ID is required"),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.SelectorArg = args[0]

			if runF != nil {
				return runF(opts)
			}
			return viewRun(opts)
		},
	}

	return cmd
}

func viewRun(opts *ViewOptions) error {
	capiClient, err := opts.CapiClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	opts.IO.StartProgressIndicatorWithLabel("Fetching agent session...")
	defer opts.IO.StopProgressIndicator()

	session, err := capiClient.GetSession(ctx, opts.SelectorArg)
	opts.IO.StopProgressIndicator()

	if err != nil {
		if errors.Is(err, capi.ErrSessionNotFound) {
			fmt.Fprintln(opts.IO.ErrOut, "session not found")
			return cmdutil.SilentError
		}
		return err
	}

	out := opts.IO.Out
	cs := opts.IO.ColorScheme()

	if session.PullRequest != nil {
		fmt.Fprintf(out, "%s • %s • %s%s\n",
			shared.ColorFuncForSessionState(*session, cs)(shared.SessionStateString(session.State)),
			cs.Bold(session.PullRequest.Title),
			session.PullRequest.Repository.NameWithOwner,
			cs.ColorFromString(prShared.ColorForPRState(*session.PullRequest))(fmt.Sprintf("#%d", session.PullRequest.Number)),
		)
	} else {
		// Should never happen, but we need to cover the path
		fmt.Fprintf(out, "%s\n", shared.ColorFuncForSessionState(*session, cs)(shared.SessionStateString(session.State)))
	}

	if session.User != nil {
		fmt.Fprintf(out, "Started on behalf of %s %s\n", session.User.Login, text.FuzzyAgo(time.Now(), session.CreatedAt))
	} else {
		// Should never happen, but we need to cover the path
		fmt.Fprintf(out, "Started %s\n", text.FuzzyAgo(time.Now(), session.CreatedAt))
	}

	// TODO(babakks): uncomment when we have the --logs option ready
	// fmt.Fprintln(out, "")
	// fmt.Fprintf(out, "For the detailed session logs, try: gh agent-task view '%s' --logs\n", opts.SelectorArg)

	if session.PullRequest != nil {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, cs.Muted("View this session on GitHub:"))
		fmt.Fprintln(out, cs.Muted(fmt.Sprintf("%s/agent-sessions/%s", session.PullRequest.URL, url.PathEscape(session.ID))))
	}

	return nil
}
