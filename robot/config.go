package robot

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/pathlib"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Activities map[string]*Activity `yaml:"activities"`
	Conda      string               `yaml:"condaConfig"`
	Ignored    []string             `yaml:"ignoreFiles"`
	Root       string
}

type Activity struct {
	Output      string       `yaml:"output"`
	Root        string       `yaml:"activityRoot"`
	Environment *Environment `yaml:"environment"`
	Action      *Action      `yaml:"action"`
}

type Environment struct {
	Path       []string `yaml:"path"`
	PythonPath []string `yaml:"pythonPath"`
}

type Action struct {
	Command []string
}

func (it *Config) RootDirectory() string {
	return it.Root
}

func (it *Config) IgnoreFiles() []string {
	if it.Ignored == nil {
		return []string{}
	}
	result := make([]string, 0, len(it.Ignored))
	for _, entry := range it.Ignored {
		result = append(result, filepath.Join(it.Root, entry))
	}
	return result
}

func (it *Config) AvailableTasks() []string {
	result := make([]string, 0, len(it.Activities))
	for name, _ := range it.Activities {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func (it *Config) DefaultTask() Task {
	if len(it.Activities) != 1 {
		return nil
	}
	var result *Activity
	for _, value := range it.Activities {
		result = value
		break
	}
	return result
}

func (it *Config) TaskByName(key string) Task {
	if len(key) == 0 {
		return it.DefaultTask()
	}
	found, ok := it.Activities[key]
	if ok {
		return found
	}
	caseless := strings.ToLower(key)
	for name, value := range it.Activities {
		if caseless == strings.ToLower(name) {
			return value
		}
	}
	return nil
}

func (it *Config) UsesConda() bool {
	return len(it.Conda) > 0
}

func (it *Config) CondaConfigFile() string {
	return filepath.Join(it.Root, it.Conda)
}

func (it *Config) WorkingDirectory(taskname string) string {
	activity := it.TaskByName(taskname)
	if activity == nil {
		return ""
	}
	return activity.WorkingDirectory(it)
}

func (it *Config) ArtifactDirectory(taskname string) string {
	activity := it.TaskByName(taskname)
	if activity == nil {
		return ""
	}
	return activity.ArtifactDirectory(it)
}

func (it *Config) Paths(taskname string) pathlib.PathParts {
	activity := it.TaskByName(taskname)
	if activity == nil {
		return pathlib.PathFrom()
	}
	return activity.Paths(it)
}

func (it *Config) PythonPaths(taskname string) pathlib.PathParts {
	activity := it.TaskByName(taskname)
	if activity == nil {
		return pathlib.PathFrom()
	}
	return activity.PythonPaths(it)
}

func (it *Config) SearchPath(taskname, location string) pathlib.PathParts {
	activity := it.TaskByName(taskname)
	if activity == nil {
		return pathlib.PathFrom()
	}
	return activity.SearchPath(it, location)
}

func (it *Config) ExecutionEnvironment(taskname, location string, inject []string, full bool) []string {
	activity := it.TaskByName(taskname)
	if activity == nil {
		return []string{}
	}
	return activity.ExecutionEnvironment(it, location, inject, full)
}

func (it *Config) Validate() (bool, error) {
	if it.Activities == nil {
		return false, errors.New("In package.yaml, 'activities:' is required!")
	}
	if len(it.Activities) == 0 {
		return false, errors.New("In package.yaml, 'activities:' must have at least one activity defined!")
	}
	if it.Conda == "" {
		return false, errors.New("In package.yaml, 'condaConfig:' is required!")
	}
	for name, activity := range it.Activities {
		if activity.Output == "" {
			return false, fmt.Errorf("In package.yaml, 'output:' is required for activity %s!", name)
		}
		if activity.Root == "" {
			return false, fmt.Errorf("In package.yaml, 'activityRoot:' is required for activity %s!", name)
		}
		if activity.Action == nil {
			return false, fmt.Errorf("In package.yaml, 'action:' is required for activity %s!", name)
		}
		if activity.Action.Command == nil {
			return false, fmt.Errorf("In package.yaml, 'action/command:' is required for activity %s!", name)
		}
		if len(activity.Action.Command) == 0 {
			return false, fmt.Errorf("In package.yaml, 'action/command:' cannot be empty for activity %s!", name)
		}
	}
	return true, nil
}

func (it *Activity) Commandline() []string {
	return it.Action.Command
}

func (it *Activity) WorkingDirectory(base Robot) string {
	return filepath.Join(base.RootDirectory(), it.Root)
}

func (it *Activity) ArtifactDirectory(base Robot) string {
	return filepath.Join(it.WorkingDirectory(base), it.Output)
}

func (it *Activity) Paths(base Robot) pathlib.PathParts {
	if it.Environment == nil {
		return pathlib.PathFrom()
	}
	return pathBuilder(base.RootDirectory(), it.Environment.Path)
}

func (it *Activity) PythonPaths(base Robot) pathlib.PathParts {
	if it.Environment == nil {
		return pathlib.PathFrom()
	}
	return pathBuilder(base.RootDirectory(), it.Environment.PythonPath)
}

func (it *Activity) SearchPath(base Robot, location string) pathlib.PathParts {
	return conda.FindPath(location).Prepend(it.Paths(base)...)
}

func PlainEnvironment(inject []string, full bool) []string {
	environment := make([]string, 0, 100)
	if full {
		environment = append(environment, os.Environ()...)
	}
	environment = append(environment, inject...)
	return environment
}

func (it *Activity) ExecutionEnvironment(base Robot, location string, inject []string, full bool) []string {
	pythonPath := it.PythonPaths(base)
	environment := PlainEnvironment(inject, full)
	searchPath := it.SearchPath(base, location)
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
		pythonPath.AsEnvironmental("PYTHONPATH"),
		fmt.Sprintf("ROBOT_ROOT=%s", it.WorkingDirectory(base)),
		fmt.Sprintf("ROBOT_ARTIFACTS=%s", it.ArtifactDirectory(base)),
	)
}

func ActivityPackageFrom(content []byte) (*Config, error) {
	config := Config{}
	err := yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func LoadActivityPackage(filename string) (Robot, error) {
	fullpath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadFile(fullpath)
	if err != nil {
		return nil, err
	}
	config, err := ActivityPackageFrom(content)
	if err != nil {
		return nil, err
	}
	config.Root = filepath.Dir(fullpath)
	return config, nil
}
