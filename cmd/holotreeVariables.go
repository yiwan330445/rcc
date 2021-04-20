package cmd

import (
	"fmt"

	"github.com/robocorp/rcc/anywork"
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
	holotreeSpace     string
	holotreeForce     bool
	holotreeJson      bool
)

func robotBlueprints(userBlueprints []string, packfile string) (robot.Robot, []string) {
	var err error
	var config robot.Robot

	blueprints := make([]string, 0, len(userBlueprints)+2)

	if Has(packfile) {
		config, err = robot.LoadRobotYaml(packfile, false)
		if err == nil {
			blueprints = append(blueprints, config.CondaConfigFile())
		}
	}

	return config, append(blueprints, userBlueprints...)
}

func holotreeExpandEnvironment(userFiles []string, packfile, environment, workspace string, validity int, space string, force bool) []string {
	var left, right *conda.Environment
	var err error
	var extra []string
	var data operations.Token

	config, filenames := robotBlueprints(userFiles, packfile)

	if Has(environment) {
		developmentEnvironment, err := robot.LoadEnvironmentSetup(environment)
		if err == nil {
			extra = developmentEnvironment.AsEnvironment()
		}
	}

	for _, filename := range filenames {
		left = right
		right, err = conda.ReadCondaYaml(filename)
		pretty.Guard(err == nil, 2, "Failure: %v", err)
		if left == nil {
			continue
		}
		right, err = left.Merge(right)
		pretty.Guard(err == nil, 3, "Failure: %v", err)
	}
	pretty.Guard(right != nil, 4, "Missing environment specification(s).")
	content, err := right.AsYaml()
	pretty.Guard(err == nil, 5, "YAML error: %v", err)
	holotreeBlueprint = []byte(content)

	anywork.Scale(200)

	tree, err := htfs.RecordEnvironment(holotreeBlueprint, force)
	pretty.Guard(err == nil, 6, "%w", err)

	path, err := tree.Restore(holotreeBlueprint, []byte(common.ControllerIdentity()), []byte(space))
	pretty.Guard(err == nil, 7, "Failed to restore blueprint %q, reason: %v", string(holotreeBlueprint), err)

	env := conda.EnvironmentExtensionFor(path)
	if config != nil {
		env = config.ExecutionEnvironment(path, extra, false)
	}

	if Has(workspace) {
		claims := operations.RunClaims(validity*60, workspace)
		data, err = operations.AuthorizeClaims(AccountName(), claims)
		pretty.Guard(err == nil, 8, "Failed to get cloud data, reason: %v", err)
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

		env := holotreeExpandEnvironment(args, robotFile, environmentFile, workspaceId, validityTime, holotreeSpace, holotreeForce)
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

	holotreeVariablesCmd.Flags().StringVarP(&holotreeSpace, "space", "s", "", "Client specific name to identify this environment.")
	holotreeVariablesCmd.MarkFlagRequired("space")
	holotreeVariablesCmd.Flags().BoolVarP(&holotreeForce, "force", "f", false, "Force environment creation with refresh.")
	holotreeVariablesCmd.Flags().BoolVarP(&holotreeJson, "json", "j", false, "Show environment as JSON.")
}
