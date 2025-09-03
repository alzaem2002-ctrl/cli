package create

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/pkg/cmd/agent-task/capi"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

// CreateOptions holds options for create command
type CreateOptions struct {
	IO               *iostreams.IOStreams
	BaseRepo         func() (ghrepo.Interface, error)
	CapiClient       func() (capi.CapiClient, error)
	Config           func() (gh.Config, error)
	ProblemStatement string
	BackOff          backoff.BackOff
}

func NewCmdCreate(f *cmdutil.Factory, runF func(*CreateOptions) error) *cobra.Command {
	opts := &CreateOptions{
		IO: f.IOStreams,
	}
	cmd := &cobra.Command{
		Use:   "create \"<task description>\"",
		Short: "Create an agent task (preview)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: We'll support prompting for the problem statement if not provided
			// and from file flags, later.
			if len(args) == 0 {
				return cmdutil.FlagErrorf("a problem statement is required")
			}

			opts.ProblemStatement = args[0]
			// Support -R/--repo override
			if f != nil {
				opts.BaseRepo = f.BaseRepo
			}
			if runF != nil {
				return runF(opts)
			}
			return createRun(opts)
		},
	}
	if f != nil {
		cmdutil.EnableRepoOverride(cmd, f)
	}

	opts.CapiClient = func() (capi.CapiClient, error) {
		cfg, err := f.Config()
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

func createRun(opts *CreateOptions) error {
	if opts.ProblemStatement == "" {
		return cmdutil.FlagErrorf("a problem statement is required")
	}
	if opts.BaseRepo == nil {
		return errors.New("failed to resolve repository")
	}
	repo, err := opts.BaseRepo()
	if err != nil || repo == nil || repo.RepoOwner() == "" || repo.RepoName() == "" {
		// Not printing the error that came back from BaseRepo() here because we want
		// something clear, human friendly, and actionable.
		return fmt.Errorf("a repository is required; re-run in a repository or supply one with --repo owner/name")
	}

	client, err := opts.CapiClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	opts.IO.StartProgressIndicatorWithLabel(fmt.Sprintf("Creating agent task in %s/%s...", repo.RepoOwner(), repo.RepoName()))
	defer opts.IO.StopProgressIndicator()

	job, err := client.CreateJob(ctx, repo.RepoOwner(), repo.RepoName(), opts.ProblemStatement)
	if err != nil {
		return err
	}

	// Print this agent session URL and exit if we happen to get it.
	// Right now, this never happens.
	if job.PullRequest != nil && job.PullRequest.Number > 0 {
		fmt.Fprintf(opts.IO.Out, "%s\n", agentSessionWebURL(repo, job))
		return nil
	}

	// Otherwise, poll using exponential backoff until we either observe a PR or hit the overall timeout.
	// Ensure we have a backoff strategy.
	if opts.BackOff == nil {
		opts.BackOff = backoff.NewExponentialBackOff(
			backoff.WithMaxElapsedTime(4*time.Second),
			backoff.WithInitialInterval(300*time.Millisecond),
			backoff.WithMaxInterval(2*time.Second),
			backoff.WithMultiplier(1.5),
		)
	}

	jobWithPR, err := fetchJobWithBackoff(ctx, client, repo, job.ID, opts.IO.ErrOut, opts.BackOff)
	if err != nil {
		return err
	}

	if jobWithPR != nil {
		opts.IO.StopProgressIndicator()
		fmt.Fprintln(opts.IO.Out, agentSessionWebURL(repo, jobWithPR))
		return nil
	}

	// Fallback if PR not yet ready
	opts.IO.StopProgressIndicator()
	fmt.Fprintf(opts.IO.Out, "job %s queued. View progress: https://github.com/copilot/agents\n", job.ID)
	return nil
}

func agentSessionWebURL(repo ghrepo.Interface, j *capi.Job) string {
	if j == nil || j.PullRequest == nil {
		return ""
	}
	if j.SessionID == "" {
		return fmt.Sprintf("https://github.com/%s/%s/pull/%d", repo.RepoOwner(), repo.RepoName(), j.PullRequest.Number)
	}
	return fmt.Sprintf("https://github.com/%s/%s/pull/%d/agent-sessions/%s", repo.RepoOwner(), repo.RepoName(), j.PullRequest.Number, j.SessionID)
}

// fetchJobWithBackoff polls the job resource until a PR number is present or the overall
// timeout elapses. It returns the updated Job on success, (nil, nil) on timeout,
// and (nil, error) only for non-retryable failures.
func fetchJobWithBackoff(ctx context.Context, client capi.CapiClient, repo ghrepo.Interface, jobID string, errOut io.Writer, bo backoff.BackOff) (*capi.Job, error) {
	// sentinel error to signal retry without surfacing to caller
	var errPRNotReady = errors.New("job not ready")

	var result *capi.Job
	retryErr := backoff.Retry(func() error {
		j, getErr := client.GetJob(ctx, repo.RepoOwner(), repo.RepoName(), jobID)
		if getErr != nil {
			fmt.Fprintf(errOut, "warning: failed to get job status: %v\n", getErr)
			return errPRNotReady
		}
		if j.PullRequest != nil && j.PullRequest.Number > 0 {
			result = j
			return nil
		}
		return errPRNotReady
	}, backoff.WithContext(bo, ctx))

	if retryErr != nil {
		if errors.Is(retryErr, errPRNotReady) {
			// Timed out or failed to fetch
			return nil, nil
		}
		return nil, retryErr
	}
	return result, nil
}
