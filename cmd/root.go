package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"dagger.io/dagger"

	"github.com/spf13/cobra"

	"github.com/aweris/gale/pkg/gale"
	"github.com/aweris/gale/pkg/gh"
	"github.com/aweris/gale/pkg/model"
	"github.com/aweris/gale/pkg/repository"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	// Flags for the gale command
	var (
		workflowName    string
		jobName         string
		export          bool
		disableCheckout bool
	)

	cmd := &cobra.Command{
		Use: "gale",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if workflowName == "" || jobName == "" {
				return fmt.Errorf("workflow and job name must be provided")
			}

			client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
			if err != nil {
				return err
			}
			defer client.Close()

			repo, err := repository.Current(ctx, client)
			if err != nil {
				return err
			}

			githubCtx, err := GetGithubContext(ctx, client)
			if err != nil {
				return err
			}

			runnerCtx := GetRunnerContext()

			gc := gale.New(client).
				WithRepository(repo).
				WithGithubContext(githubCtx, runnerCtx).
				WithModifier(func(container *dagger.Container) (*dagger.Container, error) {
					return container.WithEnvVariable("DEBUG", "1"), nil
				}).
				WithJob(workflowName, jobName)

			// TODO: temporary hack to disable checkout step. This is useful when we want to run the existing version of the repo.
			// We're mounting the current directory to the container. This is useful for testing for current directory.
			// This assumes that first step is checkout step. If not, we'll have to find the checkout step and mount the
			if disableCheckout {
				gc = gc.WithStep(&model.Step{ID: "0", Run: "echo 'Checkout Disabled to run existing version of the repo' "}, true)
			}

			result, err := gc.Exec(ctx)

			// even we have an error, we still want to export the runner directory. This is for debugging purposes.
			// no need to return the error here.
			if err != nil {
				fmt.Printf("Error executing job: %v", err)
			}

			if export {
				if err = result.ExportRunnerDirectory(ctx, fmt.Sprintf(".gale/%s", time.Now().Format(time.RFC3339Nano))); err != nil {
					return err
				}
			}

			return err
		},
	}

	// Define flags for the Step command
	cmd.Flags().StringVar(&workflowName, "workflow", "", "Name of the workflow. If workflow doesn't have name, than it must be relative path to the workflow file")
	cmd.Flags().StringVar(&jobName, "job", "", "Name of the job")
	cmd.Flags().BoolVar(&export, "export", false, "Export the runner directory after the execution. Exported directory will be placed under .gale directory in the current directory.")
	cmd.Flags().BoolVar(&disableCheckout, "disable-checkout", false, "Disable checkout step. This is useful when you want to run the existing version of the repository.")
	cmd.Flags().MarkHidden("disable-checkout") // This is a temporary flag until we have a expression parser. We need to disable checkout step for the existing repository to avoid authentication issues.

	return cmd
}

// Execute runs the command.
func Execute() {
	rootCmd := NewCommand()

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}
}

// TODO: we should find a better way to get the github contexts. This is a temporary solution.

func GetGithubContext(ctx context.Context, client *dagger.Client) (*model.GithubContext, error) {
	github, err := model.NewGithubContextFromEnv()
	if err != nil {
		return nil, err
	}

	// Not checking github.CI here because it's always true to mimic the github actions environment.
	if os.Getenv("CI") != "true" {
		// user information
		user, err := gh.CurrentUser(ctx)
		if err != nil {
			return nil, err
		}

		github.Actor = user.Login
		github.ActorID = strconv.Itoa(user.ID)
		github.TriggeringActor = user.Login

		// repository , currently we're only supporting the current repository
		repo, err := repository.Current(ctx, client)
		if err != nil {
			return nil, err
		}

		github.Repository = repo.NameWithOwner
		github.RepositoryID = repo.ID
		github.RepositoryOwner = repo.Owner.Login
		github.RepositoryOwnerID = repo.Owner.ID
		github.RepositoryURL = repo.URL
		github.Workspace = fmt.Sprintf("/home/runner/work/%s/%s", repo.Name, repo.Name)

		// token

		token, err := gh.GetToken(ctx)
		if err != nil {
			return nil, err
		}

		github.Token = token

		// default values
		github.ApiURL = "https://api.github.com"                    // TODO: make this configurable for github enterprise
		github.Event = make(map[string]interface{})                 // TODO: generate event data
		github.EventName = "push"                                   // TODO: make this configurable, this is for testing purposes
		github.EventPath = "/home/runner/_temp/workflow/event.json" // TODO: make this configurable or get from runner
		github.GraphqlURL = "https://api.github.com/graphql"        // TODO: make this configurable for github enterprise
		github.RetentionDays = 0
		github.RunID = "1"
		github.RunNumber = "1"
		github.RunAttempt = "1"
		github.SecretSource = "None"            // TODO: double check if it's possible to get this value from github cli
		github.ServerURL = "https://github.com" // TODO: make this configurable for github enterprise
		github.Workflow = ""                    // TODO: fill this value
		github.WorkflowRef = ""                 // TODO: fill this value
		github.WorkflowSHA = ""                 // TODO: fill this value
	}

	return github, nil
}

func GetRunnerContext() *model.RunnerContext {
	return &model.RunnerContext{
		Name:      "", // TODO: Not sure if we need this at all. Remove after double-checking.
		OS:        "linux",
		Arch:      "x64", // TODO: This should be determined by the host
		Temp:      "/home/runner/_temp",
		ToolCache: "/home/runner/_tool",
		Debug:     "", // TODO: "1" for debug mode, "" for normal mode get it from config
	}
}
