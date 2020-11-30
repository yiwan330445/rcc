package robot

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"

	"github.com/google/shlex"
	"gopkg.in/yaml.v2"
)

type Robot interface {
	IgnoreFiles() []string
	AvailableTasks() []string
	DefaultTask() Task
	TaskByName(string) Task
	UsesConda() bool
	CondaConfigFile() string
	RootDirectory() string
	Validate() (bool, error)

	// compatibility "string" argument (task name)
	WorkingDirectory(string) string
	ArtifactDirectory(string) string
	Paths(string) pathlib.PathParts
	PythonPaths(string) pathlib.PathParts
	SearchPath(taskname, location string) pathlib.PathParts
	ExecutionEnvironment(taskname, location string, inject []string, full bool) []string
}

type Task interface {
	WorkingDirectory(Robot) string
	ArtifactDirectory(Robot) string
	Paths(Robot) pathlib.PathParts
	PythonPaths(Robot) pathlib.PathParts
	SearchPath(Robot, string) pathlib.PathParts
	ExecutionEnvironment(robot Robot, location string, inject []string, full bool) []string
	Commandline() []string
}

type robot struct {
	Tasks      map[string]*task `yaml:"tasks"`
	Conda      string           `yaml:"condaConfigFile"`
	Ignored    []string         `yaml:"ignoreFiles"`
	Artifacts  string           `yaml:"artifactsDir"`
	Path       []string         `yaml:"PATH"`
	Pythonpath []string         `yaml:"PYTHONPATH"`
	Root       string
}

type task struct {
	Task    string   `yaml:"robotTaskName,omitempty"`
	Shell   string   `yaml:"shell,omitempty"`
	Command []string `yaml:"command,omitempty"`
}

func (it *robot) Validate() (bool, error) {
	if it.Tasks == nil {
		return false, errors.New("In robot.yaml, 'tasks:' is required!")
	}
	if len(it.Tasks) == 0 {
		return false, errors.New("In robot.yaml, 'tasks:' must have at least one task defined!")
	}
	if it.Artifacts == "" {
		return false, errors.New("In robot.yaml, 'artifactsDir:' is required!")
	}
	for name, task := range it.Tasks {
		count := 0
		if len(task.Task) > 0 {
			count += 1
		}
		if len(task.Shell) > 0 {
			count += 1
		}
		if task.Command != nil && len(task.Command) > 0 {
			count += 1
		}
		if count != 1 {
			return false, fmt.Errorf("In robot.yaml, task '%s' needs exactly one of robotTaskName/shell/command definition!", name)
		}
	}
	return true, nil
}

func (it *robot) RootDirectory() string {
	return it.Root
}

func (it *robot) IgnoreFiles() []string {
	if it.Ignored == nil {
		return []string{}
	}
	result := make([]string, 0, len(it.Ignored))
	for _, entry := range it.Ignored {
		result = append(result, filepath.Join(it.Root, entry))
	}
	return result
}

func (it *robot) AvailableTasks() []string {
	result := make([]string, 0, len(it.Tasks))
	for name, _ := range it.Tasks {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func (it *robot) DefaultTask() Task {
	if len(it.Tasks) != 1 {
		return nil
	}
	var result *task
	for _, value := range it.Tasks {
		result = value
		break
	}
	return result
}

func (it *robot) TaskByName(name string) Task {
	if len(name) == 0 {
		return it.DefaultTask()
	}
	key := strings.TrimSpace(name)
	found, ok := it.Tasks[key]
	if ok {
		return found
	}
	caseless := strings.ToLower(key)
	for name, value := range it.Tasks {
		if caseless == strings.ToLower(strings.TrimSpace(name)) {
			return value
		}
	}
	return nil
}

func (it *robot) UsesConda() bool {
	return len(it.Conda) > 0
}

func (it *robot) CondaConfigFile() string {
	return filepath.Join(it.Root, it.Conda)
}

func (it *robot) WorkingDirectory(string) string {
	return it.Root
}

func (it *robot) ArtifactDirectory(string) string {
	return filepath.Join(it.Root, it.Artifacts)
}

func pathBuilder(root string, tails []string) pathlib.PathParts {
	result := make([]string, 0, len(tails))
	for _, part := range tails {
		if filepath.IsAbs(part) && pathlib.IsDir(part) {
			result = append(result, part)
			continue
		}
		fullpath := filepath.Join(root, part)
		realpath, err := filepath.Abs(fullpath)
		if err == nil {
			result = append(result, realpath)
		}
	}
	return pathlib.PathFrom(result...)
}

func (it *robot) Paths(string) pathlib.PathParts {
	if it == nil {
		return pathlib.PathFrom()
	}
	return pathBuilder(it.Root, it.Path)
}

func (it *robot) PythonPaths(string) pathlib.PathParts {
	if it == nil {
		return pathlib.PathFrom()
	}
	return pathBuilder(it.Root, it.Pythonpath)
}

func (it *robot) SearchPath(taskname, location string) pathlib.PathParts {
	return conda.FindPath(location).Prepend(it.Paths("")...)
}

func (it *robot) ExecutionEnvironment(taskname, location string, inject []string, full bool) []string {
	environment := make([]string, 0, 100)
	if full {
		environment = append(environment, os.Environ()...)
	}
	environment = append(environment, inject...)
	searchPath := it.SearchPath(taskname, location)
	python, ok := searchPath.Which("python3", conda.FileExtensions)
	if !ok {
		python, ok = searchPath.Which("python", conda.FileExtensions)
	}
	if ok {
		environment = append(environment, "PYTHON_EXE="+python)
	}
	return append(environment,
		"CONDA_DEFAULT_ENV=rcc",
		"CONDA_EXE="+conda.BinConda(),
		"CONDA_PREFIX="+location,
		"CONDA_PROMPT_MODIFIER=(rcc)",
		"CONDA_PYTHON_EXE="+conda.BinPython(),
		"CONDA_SHLVL=1",
		"PYTHONHOME=",
		"PYTHONSTARTUP=",
		"PYTHONEXECUTABLE=",
		"PYTHONNOUSERSITE=1",
		"ROBOCORP_HOME="+conda.RobocorpHome(),
		searchPath.AsEnvironmental("PATH"),
		it.PythonPaths("").AsEnvironmental("PYTHONPATH"),
		fmt.Sprintf("ROBOT_ROOT=%s", it.WorkingDirectory("")),
		fmt.Sprintf("ROBOT_ARTIFACTS=%s", it.ArtifactDirectory("")),
	)
}

func (it *task) WorkingDirectory(robot Robot) string {
	return robot.WorkingDirectory("")
}

func (it *task) ArtifactDirectory(robot Robot) string {
	return robot.ArtifactDirectory("")
}

func (it *task) SearchPath(robot Robot, location string) pathlib.PathParts {
	return robot.SearchPath("", location)
}

func (it *task) Paths(robot Robot) pathlib.PathParts {
	return robot.Paths("")
}

func (it *task) PythonPaths(robot Robot) pathlib.PathParts {
	return robot.PythonPaths("")
}

func (it *task) ExecutionEnvironment(robot Robot, location string, inject []string, full bool) []string {
	return robot.ExecutionEnvironment("", location, inject, full)
}

func (it *task) shellCommand() []string {
	result, err := shlex.Split(it.Shell)
	if err != nil {
		common.Log("Shell parsing failure: %v with command %v", err, it.Shell)
		return []string{}
	}
	return result
}

func (it *task) taskCommand() []string {
	return []string{
		"python",
		"-m", "robot",
		"--report", "NONE",
		"--outputdir", "output",
		"--logtitle", "Task log",
		"--task", it.Task,
		".",
	}
}

func (it *task) Commandline() []string {
	if len(it.Task) > 0 {
		return it.taskCommand()
	}
	if len(it.Shell) > 0 {
		return it.shellCommand()
	}
	return it.Command
}

func robotFrom(content []byte) (*robot, error) {
	config := robot{}
	err := yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func LoadRobotYaml(filename string) (Robot, error) {
	fullpath, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", filename, err)
	}
	content, err := ioutil.ReadFile(fullpath)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", fullpath, err)
	}
	robot, err := robotFrom(content)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", fullpath, err)
	}
	robot.Root = filepath.Dir(fullpath)
	return robot, nil
}

func LoadYamlConfiguration(filename string) (Robot, error) {
	if strings.HasSuffix(filename, "package.yaml") {
		common.Log("%sWARNING! Support for 'package.yaml' is deprecated. Upgrade to 'robot.yaml'!%s", pretty.Red, pretty.Reset)
		return LoadActivityPackage(filename)
	}
	return LoadRobotYaml(filename)
}

func DetectConfigurationName(directory string) string {
	robot, err := pathlib.FindNamedPath(directory, "robot.yaml")
	if err == nil && len(robot) > 0 {
		return robot
	}
	robot, err = pathlib.FindNamedPath(directory, "package.yaml")
	if err == nil && len(robot) > 0 {
		return robot
	}
	return filepath.Join(directory, "package.yaml")
}
