package run

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"dagger.io/dagger"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aweris/gale/internal/gh"
	"github.com/aweris/gale/internal/model"
	"github.com/aweris/gale/pkg/config"
	"github.com/aweris/gale/pkg/gale"
	"github.com/aweris/gale/pkg/repository"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	// Flags for the gale command
	var (
		export          bool
		disableCheckout bool
	)

	cmd := &cobra.Command{
		Use:          "run <workflow> <job> [flags]",
		Short:        "Run Github Actions by providing workflow and job name",
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true, // don't print usage when error occurs
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			client, err := dagger.Connect(cmd.Context(), dagger.WithLogOutput(os.Stdout))
			if err != nil {
				return err
			}
			defer client.Close()

			githubCtx, err := GetGithubContext()
			if err != nil {
				return err
			}

			workflows, err := repository.LoadWorkflows(cmd.Context(), client)
			if err != nil {
				return err
			}

			workflow, ok := workflows[args[0]]
			if !ok {
				return fmt.Errorf("workflow %s not found", args[0])
			}

			_, ok = workflow.Jobs[args[1]]
			if !ok {
				return fmt.Errorf("job %s not found in workflow %s", args[1], args[0])
			}

			gc := gale.New(cfg, client).
				WithGithubContext(githubCtx).
				WithJob(args[0], args[1])

			// TODO: temporary hack to disable checkout step. This is useful when we want to run the existing version of the repo.
			// We're mounting the current directory to the container. This is useful for testing for current directory.
			// This assumes that first step is checkout step. If not, we'll have to find the checkout step and mount the
			if disableCheckout {
				gc = gc.WithStep(
					&model.Step{ID: "0", Run: "echo 'Checkout Disabled to run existing version of the repo' "},
					true,
				)
			}

			result, resultErr := gc.Exec(cmd.Context())

			// even we have an error, we still want to export the runner directory. This is for debugging purposes.
			// no need to return the error here.
			if resultErr != nil {
				fmt.Printf("Error executing job: %v", err)
			}

			if export {
				if err = result.ExportRunnerDirectory(cmd.Context(), fmt.Sprintf(".gale/runs/%s", strconv.FormatInt(time.Now().Unix(), 10))); err != nil {
					return err
				}
			}

			// make sure we return the error if there is any
			// TODO: we need to improve this.
			return resultErr
		},
	}

	// Define flags for the Step command
	cmd.Flags().BoolVar(&export, "export", false, "Export the runner directory after the execution. Exported directory will be placed under .gale directory in the current directory.")

	// Hidden flags

	cmd.Flags().BoolVar(&disableCheckout, "disable-checkout", false, "Disable checkout step. This is useful when you want to run the existing version of the repository.")
	cmd.Flags().MarkHidden("disable-checkout") // This is a temporary flag until we have a expression parser. We need to disable checkout step for the existing repository to avoid authentication issues.

	cmd.Flags().String("ghx-version", "v0.0.2", "The version of the ghx binary to use")
	cmd.Flags().MarkHidden("ghx-version")
	viper.BindPFlag("ghx_version", cmd.Flags().Lookup("ghx-version")) // bind the flag to viper so that we can use it in the config file

	return cmd
}

// TODO: we should find a better way to get the github contexts. This is a temporary solution.

func GetGithubContext() (*model.GithubContext, error) {
	github := model.NewGithubContextFromEnv()

	if !github.CI {
		// user information
		user, err := gh.CurrentUser()
		if err != nil {
			return nil, err
		}

		github.Actor = user.Login
		github.ActorID = strconv.Itoa(user.ID)
		github.TriggeringActor = user.Login

		// repository , currently we're only supporting the current repository
		repo, err := gh.CurrentRepository()
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
		token, err := gh.GetToken()
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
		github.RetentionDays = "0"
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
