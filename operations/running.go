package operations

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/robot"
	"github.com/robocorp/rcc/shell"
)

var (
	rcHosts  = []string{"RC_API_SECRET_HOST", "RC_API_WORKITEM_HOST"}
	rcTokens = []string{"RC_API_SECRET_TOKEN", "RC_API_WORKITEM_TOKEN"}
)

type RunFlags struct {
	AccountName     string
	WorkspaceId     string
	ValidityTime    int
	EnvironmentFile string
	RobotYaml       string
}

func PipFreeze(searchPath pathlib.PathParts, directory, outputDir string, environment []string) bool {
	pip, ok := searchPath.Which("pip", conda.FileExtensions)
	if !ok {
		return false
	}
	fullPip, err := filepath.EvalSymlinks(pip)
	if err != nil {
		return false
	}
	common.Log("Installed pip packages:")
	_, err = shell.New(environment, directory, fullPip, "freeze", "--all").Tee(outputDir, false)
	if err != nil {
		return false
	}
	common.Log("--")
	return true
}

func LoadTaskWithEnvironment(packfile, theTask string, force bool) (bool, robot.Robot, robot.Task, string) {
	FixRobot(packfile)
	config, err := robot.LoadRobotYaml(packfile, true)
	if err != nil {
		pretty.Exit(1, "Error: %v", err)
	}

	ok, err := config.Validate()
	if !ok {
		pretty.Exit(2, "Error: %v", err)
	}

	todo := config.TaskByName(theTask)
	if todo == nil {
		pretty.Exit(3, "Error: Could not resolve task to run. Available tasks are: %v", strings.Join(config.AvailableTasks(), ", "))
	}

	if !config.UsesConda() {
		return true, config, todo, ""
	}

	label, err := conda.NewEnvironment(force, config.CondaConfigFile())
	if err != nil {
		pretty.Exit(4, "Error: %v", err)
	}
	return false, config, todo, label
}

func SelectExecutionModel(runFlags *RunFlags, simple bool, template []string, config robot.Robot, todo robot.Task, label string, interactive bool, extraEnv map[string]string) {
	common.Timeline("execution starts.")
	defer common.Timeline("execution done.")
	if simple {
		ExecuteSimpleTask(runFlags, template, config, todo, interactive, extraEnv)
	} else {
		ExecuteTask(runFlags, template, config, todo, label, interactive, extraEnv)
	}
}

func ExecuteSimpleTask(flags *RunFlags, template []string, config robot.Robot, todo robot.Task, interactive bool, extraEnv map[string]string) {
	common.Debug("Command line is: %v", template)
	task := make([]string, len(template))
	copy(task, template)
	searchPath := pathlib.TargetPath()
	searchPath = searchPath.Prepend(todo.Paths(config)...)
	found, ok := searchPath.Which(task[0], conda.FileExtensions)
	if !ok {
		pretty.Exit(6, "Error: Cannot find command: %v", task[0])
	}
	fullpath, err := filepath.EvalSymlinks(found)
	if err != nil {
		pretty.Exit(7, "Error: %v", err)
	}
	var data Token
	if len(flags.WorkspaceId) > 0 {
		claims := RunClaims(flags.ValidityTime*60, flags.WorkspaceId)
		data, err = AuthorizeClaims(flags.AccountName, claims)
	}
	if err != nil {
		pretty.Exit(8, "Error: %v", err)
	}
	task[0] = fullpath
	directory := todo.WorkingDirectory(config)
	environment := robot.PlainEnvironment([]string{searchPath.AsEnvironmental("PATH")}, true)
	if len(data) > 0 {
		endpoint := data["endpoint"]
		for _, key := range rcHosts {
			environment = append(environment, fmt.Sprintf("%s=%s", key, endpoint))
		}
		token := data["token"]
		for _, key := range rcTokens {
			environment = append(environment, fmt.Sprintf("%s=%s", key, token))
		}
		environment = append(environment, fmt.Sprintf("RC_WORKSPACE_ID=%s", flags.WorkspaceId))
	}
	if extraEnv != nil {
		for key, value := range extraEnv {
			environment = append(environment, fmt.Sprintf("%s=%s", key, value))
		}
	}
	outputDir := todo.ArtifactDirectory(config)
	common.Debug("DEBUG: about to run command - %v", task)
	_, err = shell.New(environment, directory, task...).Tee(outputDir, interactive)
	if err != nil {
		pretty.Exit(9, "Error: %v", err)
	}
	pretty.Ok()
}

func ExecuteTask(flags *RunFlags, template []string, config robot.Robot, todo robot.Task, label string, interactive bool, extraEnv map[string]string) {
	common.Debug("Command line is: %v", template)
	developmentEnvironment, err := robot.LoadEnvironmentSetup(flags.EnvironmentFile)
	if err != nil {
		pretty.Exit(5, "Error: %v", err)
	}
	task := make([]string, len(template))
	copy(task, template)
	searchPath := todo.SearchPath(config, label)
	found, ok := searchPath.Which(task[0], conda.FileExtensions)
	if !ok {
		pretty.Exit(6, "Error: Cannot find command: %v", task[0])
	}
	fullpath, err := filepath.EvalSymlinks(found)
	if err != nil {
		pretty.Exit(7, "Error: %v", err)
	}
	var data Token
	if len(flags.WorkspaceId) > 0 {
		claims := RunClaims(flags.ValidityTime*60, flags.WorkspaceId)
		data, err = AuthorizeClaims(flags.AccountName, claims)
	}
	if err != nil {
		pretty.Exit(8, "Error: %v", err)
	}
	task[0] = fullpath
	directory := todo.WorkingDirectory(config)
	environment := todo.ExecutionEnvironment(config, label, developmentEnvironment.AsEnvironment(), true)
	if len(data) > 0 {
		endpoint := data["endpoint"]
		for _, key := range rcHosts {
			environment = append(environment, fmt.Sprintf("%s=%s", key, endpoint))
		}
		token := data["token"]
		for _, key := range rcTokens {
			environment = append(environment, fmt.Sprintf("%s=%s", key, token))
		}
		environment = append(environment, fmt.Sprintf("RC_WORKSPACE_ID=%s", flags.WorkspaceId))
	}
	if extraEnv != nil {
		for key, value := range extraEnv {
			environment = append(environment, fmt.Sprintf("%s=%s", key, value))
		}
	}
	before := make(map[string]string)
	beforeHash, beforeErr := conda.DigestFor(label, before)
	outputDir := todo.ArtifactDirectory(config)
	if !common.Silent && !interactive {
		PipFreeze(searchPath, directory, outputDir, environment)
	}
	common.Debug("DEBUG: about to run command - %v", task)
	_, err = shell.New(environment, directory, task...).Tee(outputDir, interactive)
	after := make(map[string]string)
	afterHash, afterErr := conda.DigestFor(label, after)
	diagnoseLive(label, beforeHash, afterHash, beforeErr, afterErr, before, after)
	if err != nil {
		pretty.Exit(9, "Error: %v", err)
	}
	pretty.Ok()
}

func MakeRelativeMap(root string, entries map[string]string) map[string]string {
	result := make(map[string]string)
	for key, value := range entries {
		if !strings.HasPrefix(key, root) {
			result[key] = value
			continue
		}
		short, err := filepath.Rel(root, key)
		if err == nil {
			key = short
		}
		result[key] = value
	}
	return result
}

func DirhashDiff(history, future map[string]string, warning bool) {
	removed := []string{}
	added := []string{}
	changed := []string{}
	for key, value := range history {
		next, ok := future[key]
		if !ok {
			removed = append(removed, key)
			continue
		}
		if value != next {
			changed = append(changed, key)
		}
	}
	for key, _ := range future {
		_, ok := history[key]
		if !ok {
			added = append(added, key)
		}
	}
	if len(removed)+len(added)+len(changed) == 0 {
		return
	}
	common.Log("----  rcc env diff  ----")
	sort.Strings(removed)
	sort.Strings(added)
	sort.Strings(changed)
	separate := false
	for _, folder := range removed {
		common.Log("- diff: removed %q", folder)
		separate = true
	}
	if len(changed) > 0 {
		if separate {
			common.Log("-------")
			separate = false
		}
		for _, folder := range changed {
			common.Log("- diff: changed %q", folder)
			separate = true
		}
	}
	if len(added) > 0 {
		if separate {
			common.Log("-------")
			separate = false
		}
		for _, folder := range added {
			common.Log("- diff: added %q", folder)
			separate = true
		}
	}
	if warning {
		if separate {
			common.Log("-------")
			separate = false
		}
		common.Log("Notice: Robot run modified the environment which will slow down the next run.")
		common.Log("        Please inform the robot developer about this.")
	}
	common.Log("----  rcc env diff  ----")
}

func diagnoseLive(label string, beforeHash, afterHash []byte, beforeErr, afterErr error, beforeDetails, afterDetails map[string]string) {
	if beforeErr != nil || afterErr != nil {
		common.Debug("live %q diagnosis failed, before: %v, after: %v", label, beforeErr, afterErr)
		return
	}
	beforeSummary := fmt.Sprintf("%02x", beforeHash)
	afterSummary := fmt.Sprintf("%02x", afterHash)
	if beforeSummary == afterSummary {
		common.Debug("live %q diagnosis: did not change during run [%s]", label, afterSummary)
		return
	}
	common.Debug("live %q diagnosis: corrupted [%s] => [%s]", label, beforeSummary, afterSummary)
	beforeDetails = MakeRelativeMap(label, beforeDetails)
	afterDetails = MakeRelativeMap(label, afterDetails)
	DirhashDiff(beforeDetails, afterDetails, true)
}
