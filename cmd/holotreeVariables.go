package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/operations"
	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/robot"
	"github.com/spf13/cobra"
)

var (
	holotreeBlueprint []byte
	holotreeForce     bool
	holotreeJson      bool
)

func holotreeExpandEnvironment(userFiles []string, packfile, environment, workspace string, validity int, force bool) []string {
	var extra []string
	var data operations.Token

	config, holotreeBlueprint, err := htfs.ComposeFinalBlueprint(userFiles, packfile)
	pretty.Guard(err == nil, 5, "%s", err)

	condafile := filepath.Join(conda.RobocorpTemp(), htfs.BlueprintHash(holotreeBlueprint))
	err = os.WriteFile(condafile, holotreeBlueprint, 0o640)
	pretty.Guard(err == nil, 6, "%s", err)

	path, err := htfs.NewEnvironment(force, condafile)
	pretty.Guard(err == nil, 6, "%s", err)

	if Has(environment) {
		developmentEnvironment, err := robot.LoadEnvironmentSetup(environment)
		if err == nil {
			extra = developmentEnvironment.AsEnvironment()
		}
	}

	env := conda.EnvironmentExtensionFor(path)
	if config != nil {
		env = config.ExecutionEnvironment(path, extra, false)
	}

	if Has(workspace) {
		claims := operations.RunRobotClaims(validity*60, workspace)
		data, err = operations.AuthorizeClaims(AccountName(), claims)
		pretty.Guard(err == nil, 9, "Failed to get cloud data, reason: %v", err)
	}

	if len(data) > 0 {
		endpoint := data["endpoint"]
		for _, key := range rcHosts {
			env = append(env, fmt.Sprintf("%s=%s", key, endpoint))
		}
		token := data["token"]
		for _, key := range rcTokens {
			env = append(env, fmt.Sprintf("%s=%s", key, token))
		}
		env = append(env, fmt.Sprintf("RC_WORKSPACE_ID=%s", workspaceId))
	}

	return env
}

var holotreeVariablesCmd = &cobra.Command{
	Use:     "variables conda.yaml+",
	Aliases: []string{"vars"},
	Short:   "Do holotree operations.",
	Long:    "Do holotree operations.",
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Holotree variables command lasted").Report()
		}

		ok := conda.MustMicromamba()
		pretty.Guard(ok, 1, "Could not get micromamba installed.")

		env := holotreeExpandEnvironment(args, robotFile, environmentFile, workspaceId, validityTime, holotreeForce)
		if holotreeJson {
			asJson(env)
		} else {
			asExportedText(env)
		}
	},
}

func init() {
	holotreeCmd.AddCommand(holotreeVariablesCmd)
	holotreeVariablesCmd.Flags().StringVarP(&environmentFile, "environment", "e", "", "Full path to 'env.json' development environment data file. <optional>")
	holotreeVariablesCmd.Flags().StringVarP(&robotFile, "robot", "r", "robot.yaml", "Full path to 'robot.yaml' configuration file. <optional>")
	holotreeVariablesCmd.Flags().StringVarP(&workspaceId, "workspace", "w", "", "Optional workspace id to get authorization tokens for. <optional>")
	holotreeVariablesCmd.Flags().IntVarP(&validityTime, "minutes", "m", 0, "How many minutes the authorization should be valid for. <optional>")
	holotreeVariablesCmd.Flags().StringVarP(&accountName, "account", "a", "", "Account used for workspace. <optional>")

	holotreeVariablesCmd.Flags().StringVarP(&common.HolotreeSpace, "space", "s", "", "Client specific name to identify this environment.")
	holotreeVariablesCmd.MarkFlagRequired("space")
	holotreeVariablesCmd.Flags().BoolVarP(&holotreeForce, "force", "f", false, "Force environment creation with refresh.")
	holotreeVariablesCmd.Flags().BoolVarP(&holotreeJson, "json", "j", false, "Show environment as JSON.")
}
