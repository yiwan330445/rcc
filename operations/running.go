package operations

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/shlex"
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
	NoPipFreeze     bool
}

func FreezeEnvironmentListing(label string, config robot.Robot) {
	goldenfile := conda.GoldenMasterFilename(label)
	listing := conda.LoadWantedDependencies(goldenfile)
	if len(listing) == 0 {
		common.Log("No dependencies found at %q", goldenfile)
		return
	}
	env, err := conda.ReadCondaYaml(config.CondaConfigFile())
	if err != nil {
		common.Log("Could not read %q, reason: %v", config.CondaConfigFile(), err)
		return
	}
	frozen := env.FreezeDependencies(listing)
	err = frozen.SaveAs(config.FreezeFilename())
	if err != nil {
		common.Log("Could not save %q, reason: %v", config.FreezeFilename(), err)
	}
}

func ExecutionEnvironmentListing(wantedfile, label string, searchPath pathlib.PathParts, directory, outputDir string, environment []string) bool {
	common.Timeline("execution environment listing")
	defer common.Log("--")
	goldenfile := conda.GoldenMasterFilename(label)
	err := conda.SideBySideViewOfDependencies(goldenfile, wantedfile)
	if err != nil {
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
	}
	if err != nil {
		return false
	}
	return true
}

func LoadAnyTaskEnvironment(packfile string, force bool) (bool, robot.Robot, robot.Task, string) {
	FixRobot(packfile)
	config, err := robot.LoadRobotYaml(packfile, false)
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
		pretty.Exit(3, "Error: Could not resolve what task to run. Select one using --task option.\nAvailable task names are: %v.", strings.Join(config.AvailableTasks(), ", "))
	}

	if config.HasHolozip() && !common.UsesHolotree() {
		pretty.Exit(4, "Error: this robot requires holotree, but no --space was given!")
	}

	if !config.UsesConda() {
		return true, config, todo, ""
	}

	label, _, err := htfs.NewEnvironment(config.CondaConfigFile(), config.Holozip(), true, force)
	if err != nil {
		pretty.Exit(4, "Error: %v", err)
	}
	return false, config, todo, label
}

func SelectExecutionModel(runFlags *RunFlags, simple bool, template []string, config robot.Robot, todo robot.Task, label string, interactive bool, extraEnv map[string]string) {
	common.Timeline("robot execution starts (simple=%v).", simple)
	defer common.Timeline("robot execution done.")
	pathlib.EnsureDirectoryExists(config.ArtifactDirectory())
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
	outputDir, err := pathlib.EnsureDirectory(config.ArtifactDirectory())
	if err != nil {
		pretty.Exit(9, "Error: %v", err)
	}
	common.Debug("about to run command - %v", task)
	if common.NoOutputCapture {
		_, err = shell.New(environment, directory, task...).Execute(interactive)
	} else {
		_, err = shell.New(environment, directory, task...).Tee(outputDir, interactive)
	}
	if err != nil {
		pretty.Exit(10, "Error: %v", err)
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
	environment := config.RobotExecutionEnvironment(label, developmentEnvironment.AsEnvironment(), true)
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
	outputDir, err := pathlib.EnsureDirectory(config.ArtifactDirectory())
	if err != nil {
		pretty.Exit(9, "Error: %v", err)
	}
	if !flags.NoPipFreeze && !flags.Assistant && !common.Silent && !interactive {
		wantedfile, _ := config.DependenciesFile()
		ExecutionEnvironmentListing(wantedfile, label, searchPath, directory, outputDir, environment)
	}
	FreezeEnvironmentListing(label, config)

	preRunScripts := config.PreRunScripts()
	if preRunScripts != nil && len(preRunScripts) > 0 {
		common.Debug("===  pre run script phase ===")
		for _, script := range preRunScripts {
			scriptCommand, err := shlex.Split(script)
			if err != nil {
				pretty.Exit(11, "%sScript '%s' parsing failure: %v%s", pretty.Red, script, err, pretty.Reset)
			}
			common.Debug("Running pre run script '%s' ...", script)
			_, err = shell.New(environment, directory, scriptCommand...).Execute(interactive)
			if err != nil {
				pretty.Exit(12, "%sScript '%s' failure: %v%s", pretty.Red, script, err, pretty.Reset)
			}
		}
	}

	common.Debug("about to run command - %v", task)
	if common.NoOutputCapture {
		_, err = shell.New(environment, directory, task...).Execute(interactive)
	} else {
		_, err = shell.New(environment, directory, task...).Tee(outputDir, interactive)
	}
	after := make(map[string]string)
	afterHash, afterErr := conda.DigestFor(label, after)
	conda.DiagnoseDirty(label, label, beforeHash, afterHash, beforeErr, afterErr, before, after, true)
	if err != nil {
		pretty.Exit(10, "Error: %v", err)
	}
	pretty.Ok()
}
