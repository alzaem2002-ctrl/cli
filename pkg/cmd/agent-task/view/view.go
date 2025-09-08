package view

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/internal/ghinstance"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/internal/prompter"
	"github.com/cli/cli/v2/internal/text"
	"github.com/cli/cli/v2/pkg/cmd/agent-task/capi"
	"github.com/cli/cli/v2/pkg/cmd/agent-task/shared"
	prShared "github.com/cli/cli/v2/pkg/cmd/pr/shared"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

const defaultLimit = 40

type ViewOptions struct {
	IO         *iostreams.IOStreams
	BaseRepo   func() (ghrepo.Interface, error)
	CapiClient func() (capi.CapiClient, error)
	HttpClient func() (*http.Client, error)
	Finder     prShared.PRFinder
	Prompter   prompter.Prompter

	SelectorArg string
	PRNumber    int
	SessionID   string
}

func NewCmdView(f *cmdutil.Factory, runF func(*ViewOptions) error) *cobra.Command {
	opts := &ViewOptions{
		IO:         f.IOStreams,
		HttpClient: f.HttpClient,
		CapiClient: shared.CapiClientFunc(f),
		Prompter:   f.Prompter,
	}

	cmd := &cobra.Command{
		Use:   "view [<session-id> | <pr-number> | <pr-url> | <pr-branch>]",
		Short: "View an agent task session",
		Long: heredoc.Doc(`
			View an agent task session.
		`),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Support -R/--repo override
			opts.BaseRepo = f.BaseRepo

			if len(args) > 0 {
				opts.SelectorArg = args[0]
				if shared.IsSessionID(opts.SelectorArg) {
					opts.SessionID = opts.SelectorArg
				}
			}

			if opts.SessionID == "" && !opts.IO.CanPrompt() {
				return fmt.Errorf("session ID is required when not running interactively")
			}

			if opts.Finder == nil {
				opts.Finder = prShared.NewFinder(f)
			}

			if runF != nil {
				return runF(opts)
			}
			return viewRun(opts)
		},
	}

	cmdutil.EnableRepoOverride(cmd, f)

	return cmd
}

func viewRun(opts *ViewOptions) error {
	capiClient, err := opts.CapiClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	cs := opts.IO.ColorScheme()

	opts.IO.StartProgressIndicatorWithLabel("Fetching agent session...")
	defer opts.IO.StopProgressIndicator()

	var session *capi.Session

	if opts.SessionID != "" {
		if sess, err := capiClient.GetSession(ctx, opts.SessionID); err != nil {
			if errors.Is(err, capi.ErrSessionNotFound) {
				fmt.Fprintln(opts.IO.ErrOut, "session not found")
				return cmdutil.SilentError
			}
			return err
		} else {
			session = sess
		}
	} else {
		var resourceID int64

		if opts.SelectorArg != "" {
			// Finder does not support the PR/issue reference format (e.g. owner/repo#123)
			// so we need to check if the selector arg is a reference and fetch the PR
			// directly.
			if repo, num, err := prShared.ParseFullReference(opts.SelectorArg); err == nil {
				// Since the selector was a reference (i.e. without hostname data), we need to
				// check the base repo to get the hostname.
				baseRepo, err := opts.BaseRepo()
				if err != nil {
					return err
				}

				hostname := baseRepo.RepoHost()
				if hostname != ghinstance.Default() {
					return fmt.Errorf("agent tasks are not supported on this host: %s", hostname)
				}

				resourceID, err = capiClient.GetPullRequestDatabaseID(ctx, hostname, repo.RepoOwner(), repo.RepoName(), num)
				if err != nil {
					return fmt.Errorf("failed to fetch pull request: %w", err)
				}
			}
		}

		if resourceID == 0 {
			findOptions := prShared.FindOptions{
				Selector: opts.SelectorArg,
				Fields:   []string{"id", "url", "fullDatabaseId"},
			}

			pr, repo, err := opts.Finder.Find(findOptions)
			if err != nil {
				return err
			}

			if repo.RepoHost() != ghinstance.Default() {
				return fmt.Errorf("agent tasks are not supported on this host: %s", repo.RepoHost())
			}

			databaseID, err := strconv.ParseInt(pr.FullDatabaseID, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse pull request: %w", err)
			}

			resourceID = databaseID
		}

		// TODO(babakks): currently we just fetch a pre-defined number of
		// matching sessions to avoid hitting the API too many times, but it's
		// technically possible for a PR to be associated with lots of sessions
		// (i.e. above our selected limit).
		sessions, err := capiClient.ListSessionsByResourceID(ctx, "pull", resourceID, defaultLimit)
		if err != nil {
			return fmt.Errorf("failed to list sessions for pull request: %w", err)
		}

		if len(sessions) == 0 {
			fmt.Fprintln(opts.IO.ErrOut, "no session found for pull request")
			return cmdutil.SilentError
		}

		session = sessions[0]
		if len(sessions) > 1 {
			now := time.Now()
			options := make([]string, 0, len(sessions))
			for _, session := range sessions {
				options = append(options, fmt.Sprintf(
					"%s %s • %s",
					shared.SessionSymbol(cs, session.State),
					session.Name,
					text.FuzzyAgo(now, session.CreatedAt),
				))
			}

			opts.IO.StopProgressIndicator()
			selected, err := opts.Prompter.Select("Select a session", options[0], options)
			if err != nil {
				return err
			}

			session = sessions[selected]
		}
	}

	opts.IO.StopProgressIndicator()

	out := opts.IO.Out

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
