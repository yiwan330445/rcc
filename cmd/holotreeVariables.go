package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func asSimpleMap(line string) map[string]string {
	parts := strings.SplitN(strings.TrimSpace(line), "=", 2)
	if len(parts) != 2 {
		return nil
	}
	if len(parts[0]) == 0 {
		return nil
	}
	result := make(map[string]string)
	result["key"] = parts[0]
	result["value"] = parts[1]
	return result
}

func asJson(items []string) error {
	result := make([]map[string]string, 0, len(items))
	for _, line := range items {
		entry := asSimpleMap(line)
		if entry != nil {
			result = append(result, entry)
		}
	}
	content, err := operations.NiceJsonOutput(result)
	if err != nil {
		return err
	}
	common.Stdout("%s\n", content)
	return nil
}

func asExportedText(items []string) {
	prefix := "export"
	if conda.IsWindows() {
		prefix = "SET"
	}
	for _, line := range items {
		common.Stdout("%s %s\n", prefix, line)
	}
}

func holotreeExpandEnvironment(userFiles []string, packfile, environment, workspace string, validity int, force bool) []string {
	var extra []string
	var data operations.Token

	config, holotreeBlueprint, err := htfs.ComposeFinalBlueprint(userFiles, packfile)
	pretty.Guard(err == nil, 5, "%s", err)

	condafile := filepath.Join(conda.RobocorpTemp(), htfs.BlueprintHash(holotreeBlueprint))
	err = os.WriteFile(condafile, holotreeBlueprint, 0o644)
	pretty.Guard(err == nil, 6, "%s", err)

	holozip := ""
	if config != nil {
		holozip = config.Holozip()
	}
	path, err := htfs.NewEnvironment(force, condafile, holozip)
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

	holotreeVariablesCmd.Flags().StringVarP(&common.HolotreeSpace, "space", "s", "user", "Client specific name to identify this environment.")
	holotreeVariablesCmd.Flags().BoolVarP(&holotreeForce, "force", "f", false, "Force environment creation with refresh.")
	holotreeVariablesCmd.Flags().BoolVarP(&holotreeJson, "json", "j", false, "Show environment as JSON.")
}
