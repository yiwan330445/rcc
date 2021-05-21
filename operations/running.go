package operations

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/htfs"
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
	Assistant       bool
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
	if common.NoOutputCapture {
		_, err = shell.New(environment, directory, fullPip, "freeze", "--all").Execute(false)
	} else {
		_, err = shell.New(environment, directory, fullPip, "freeze", "--all").Tee(outputDir, false)
	}
	if err != nil {
		return false
	}
	common.Log("--")
	return true
}

func LoadAnyTaskEnvironment(packfile string, force bool) (bool, robot.Robot, robot.Task, string) {
	FixRobot(packfile)
	config, err := robot.LoadRobotYaml(packfile, true)
	if err != nil {
		pretty.Exit(1, "Error: %v", err)
	}
	anytasks := config.AvailableTasks()
	if len(anytasks) == 0 {
		pretty.Exit(1, "Could not find tasks from %q.", packfile)
	}
	return LoadTaskWithEnvironment(packfile, anytasks[0], force)
}

func LoadTaskWithEnvironment(packfile, theTask string, force bool) (bool, robot.Robot, robot.Task, string) {
	common.Timeline("task environment load started")
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

	var label string
	if len(common.HolotreeSpace) > 0 {
		label, err = htfs.NewEnvironment(force, config.CondaConfigFile())
	} else {
		label, err = conda.NewEnvironment(force, config.CondaConfigFile())
	}
	if err != nil {
		pretty.Exit(4, "Error: %v", err)
	}
	return false, config, todo, label
}

func SelectExecutionModel(runFlags *RunFlags, simple bool, template []string, config robot.Robot, todo robot.Task, label string, interactive bool, extraEnv map[string]string) {
	common.Timeline("robot execution starts (simple=%v).", simple)
	defer common.Timeline("robot execution done.")
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
	searchPath = searchPath.Prepend(config.Paths()...)
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
		claims := RunRobotClaims(flags.ValidityTime*60, flags.WorkspaceId)
		data, err = AuthorizeClaims(flags.AccountName, claims)
	}
	if err != nil {
		pretty.Exit(8, "Error: %v", err)
	}
	task[0] = fullpath
	directory := config.WorkingDirectory()
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
	outputDir := config.ArtifactDirectory()
	common.Debug("DEBUG: about to run command - %v", task)
	if common.NoOutputCapture {
		_, err = shell.New(environment, directory, task...).Execute(interactive)
	} else {
		_, err = shell.New(environment, directory, task...).Tee(outputDir, interactive)
	}
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
	searchPath := config.SearchPath(label)
	found, ok := searchPath.Which(task[0], conda.FileExtensions)
	if !ok {
		pretty.Exit(6, "Error: Cannot find command: %v", task[0])
	}
	fullpath, err := filepath.EvalSymlinks(found)
	if err != nil {
		pretty.Exit(7, "Error: %v", err)
	}
	var data Token
	if !flags.Assistant && len(flags.WorkspaceId) > 0 {
		claims := RunRobotClaims(flags.ValidityTime*60, flags.WorkspaceId)
		data, err = AuthorizeClaims(flags.AccountName, claims)
	}
	if err != nil {
		pretty.Exit(8, "Error: %v", err)
	}
	task[0] = fullpath
	directory := config.WorkingDirectory()
	environment := config.ExecutionEnvironment(label, developmentEnvironment.AsEnvironment(), true)
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
	outputDir := config.ArtifactDirectory()
	if !flags.Assistant && !common.Silent && !interactive {
		PipFreeze(searchPath, directory, outputDir, environment)
	}
	common.Debug("DEBUG: about to run command - %v", task)
	if common.NoOutputCapture {
		_, err = shell.New(environment, directory, task...).Execute(interactive)
	} else {
		_, err = shell.New(environment, directory, task...).Tee(outputDir, interactive)
	}
	after := make(map[string]string)
	afterHash, afterErr := conda.DigestFor(label, after)
	conda.DiagnoseDirty(label, label, beforeHash, afterHash, beforeErr, afterErr, before, after, true)
	if err != nil {
		pretty.Exit(9, "Error: %v", err)
	}
	pretty.Ok()
}
