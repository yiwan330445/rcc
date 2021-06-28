package robot

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/xviper"

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
	CondaHash() string
	RootDirectory() string
	HasHolozip() bool
	Holozip() string
	Validate() (bool, error)
	Diagnostics(*common.DiagnosticStatus, bool)
	DependenciesFile() (string, bool)
	IdealCondaYaml() (string, bool)

	WorkingDirectory() string
	ArtifactDirectory() string
	FreezeFilename() string
	Paths() pathlib.PathParts
	PythonPaths() pathlib.PathParts
	SearchPath(location string) pathlib.PathParts
	ExecutionEnvironment(location string, inject []string, full bool) []string
}

type Task interface {
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

func (it *robot) diagnoseTasks(diagnose common.Diagnoser) {
	if it.Tasks == nil {
		diagnose.Fail("", "Missing 'tasks:' from robot.yaml.")
		return
	}
	ok := true
	if len(it.Tasks) == 0 {
		diagnose.Fail("", "There must be at least one task defined in 'tasks:' section in robot.yaml.")
		ok = false
	} else {
		diagnose.Ok("Tasks are defined in robot.yaml")
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
			diagnose.Fail("", "In robot.yaml, task '%s' needs exactly one of robotTaskName/shell/command definition!", name)
			ok = false
		}
	}
	if ok {
		diagnose.Ok("Each task has exactly one definition.")
	}
}

func (it *robot) diagnoseVariousPaths(diagnose common.Diagnoser) {
	ok := true
	for _, path := range it.Path {
		if filepath.IsAbs(path) {
			diagnose.Fail("", "PATH entry %q seems to be absolute, which makes robot machine dependent.", path)
			ok = false
		}
	}
	if ok {
		diagnose.Ok("PATH settings in robot.yaml are ok.")
	}
	ok = true
	for _, path := range it.Pythonpath {
		if filepath.IsAbs(path) {
			diagnose.Fail("", "PYTHONPATH entry %q seems to be absolute, which makes robot machine dependent.", path)
			ok = false
		}
	}
	if ok {
		diagnose.Ok("PYTHONPATH settings in robot.yaml are ok.")
	}
	ok = true
	if it.Ignored == nil || len(it.Ignored) == 0 {
		diagnose.Warning("", "No ignoreFiles defined, so everything ends up inside robot.zip file.")
		ok = false
	} else {
		for _, path := range it.Ignored {
			if filepath.IsAbs(path) {
				diagnose.Fail("", "ignoreFiles entry %q seems to be absolute, which makes robot machine dependent.", path)
				ok = false
			}
		}
	}
	if ok {
		diagnose.Ok("ignoreFiles settings in robot.yaml are ok.")
	}
}

func (it *robot) Diagnostics(target *common.DiagnosticStatus, production bool) {
	diagnose := target.Diagnose("Robot")
	it.diagnoseTasks(diagnose)
	it.diagnoseVariousPaths(diagnose)
	if it.Artifacts == "" {
		diagnose.Fail("", "In robot.yaml, 'artifactsDir:' is required!")
	} else {
		if filepath.IsAbs(it.Artifacts) {
			diagnose.Fail("", "artifactDir %q seems to be absolute, which makes robot machine dependent.", it.Artifacts)
		} else {
			diagnose.Ok("Artifacts directory defined in robot.yaml")
		}
	}
	if it.Conda == "" {
		diagnose.Ok("In robot.yaml, 'condaConfigFile:' is missing. So this is shell robot.")
	} else {
		if filepath.IsAbs(it.Conda) {
			diagnose.Fail("", "condaConfigFile %q seems to be absolute, which makes robot machine dependent.", it.Artifacts)
		} else {
			diagnose.Ok("In robot.yaml, 'condaConfigFile:' is present. So this is python robot.")
			condaEnv, err := conda.ReadCondaYaml(it.CondaConfigFile())
			if err != nil {
				diagnose.Fail("", "From robot.yaml, loading conda.yaml failed with: %v", err)
			} else {
				condaEnv.Diagnostics(target, production)
			}
		}
	}
	target.Details["robot-use-conda"] = fmt.Sprintf("%v", it.UsesConda())
	target.Details["robot-conda-file"] = it.CondaConfigFile()
	target.Details["robot-conda-hash"] = it.CondaHash()
	target.Details["hololib.zip"] = it.Holozip()
	plan, ok := conda.InstallationPlan(it.CondaHash())
	if ok {
		target.Details["robot-conda-plan"] = plan
	}
	target.Details["robot-root-directory"] = it.RootDirectory()
	target.Details["robot-working-directory"] = it.WorkingDirectory()
	target.Details["robot-artifact-directory"] = it.ArtifactDirectory()
	target.Details["robot-paths"] = strings.Join(it.Paths(), ", ")
	target.Details["robot-python-paths"] = strings.Join(it.PythonPaths(), ", ")
	dependencies, ok := it.DependenciesFile()
	if !ok {
		dependencies = "missing"
	} else {
		if it.VerifyCondaDependencies() {
			diagnose.Ok("Dependencies in conda.yaml and dependencies.yaml match.")
		}
	}
	target.Details["robot-dependencies-yaml"] = dependencies
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

func (it *robot) DependenciesFile() (string, bool) {
	filename := filepath.Join(it.Root, "dependencies.yaml")
	return filename, pathlib.IsFile(filename)
}

func (it *robot) IdealCondaYaml() (string, bool) {
	wanted, ok := it.DependenciesFile()
	if !ok {
		return "", false
	}
	dependencies := conda.LoadWantedDependencies(wanted)
	if len(dependencies) == 0 {
		return "", false
	}
	condaEnv, err := conda.ReadCondaYaml(it.CondaConfigFile())
	if err != nil {
		return "", false
	}
	ideal, ok := condaEnv.FromDependencies(dependencies)
	if !ok {
		return "", false
	}
	body, err := ideal.AsYaml()
	if err != nil {
		return "", false
	}
	return body, true
}

func (it *robot) VerifyCondaDependencies() bool {
	wanted, ok := it.DependenciesFile()
	if !ok {
		return true
	}
	dependencies := conda.LoadWantedDependencies(wanted)
	if len(dependencies) == 0 {
		return true
	}
	condaEnv, err := conda.ReadCondaYaml(it.CondaConfigFile())
	if err != nil {
		return true
	}
	ideal, ok := condaEnv.FromDependencies(dependencies)
	if !ok {
		body, err := ideal.AsYaml()
		if err == nil {
			fmt.Println("IDEAL:", body)
		}
	}
	return ok
}

func (it *robot) RootDirectory() string {
	return it.Root
}

func (it *robot) HasHolozip() bool {
	return len(it.Holozip()) > 0
}

func (it *robot) Holozip() string {
	zippath := filepath.Join(it.Root, "hololib.zip")
	if pathlib.IsFile(zippath) {
		return zippath
	}
	return ""
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

func (it *robot) CondaHash() string {
	result, err := conda.CalculateComboHash(filepath.Join(it.Root, it.Conda))
	if err != nil {
		return ""
	}
	return result
}

func (it *robot) WorkingDirectory() string {
	return it.Root
}

func (it *robot) FreezeFilename() string {
	return filepath.Join(it.ArtifactDirectory(), fmt.Sprintf("environment_%s_%s_freeze.yaml", runtime.GOOS, runtime.GOARCH))
}

func (it *robot) ArtifactDirectory() string {
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

func (it *robot) Paths() pathlib.PathParts {
	if it == nil {
		return pathlib.PathFrom()
	}
	return pathBuilder(it.Root, it.Path)
}

func (it *robot) PythonPaths() pathlib.PathParts {
	if it == nil {
		return pathlib.PathFrom()
	}
	return pathBuilder(it.Root, it.Pythonpath)
}

func (it *robot) SearchPath(location string) pathlib.PathParts {
	return conda.FindPath(location).Prepend(it.Paths()...)
}

func (it *robot) ExecutionEnvironment(location string, inject []string, full bool) []string {
	environment := make([]string, 0, 100)
	if full {
		environment = append(environment, os.Environ()...)
	}
	environment = append(environment, inject...)
	searchPath := it.SearchPath(location)
	python, ok := searchPath.Which("python3", conda.FileExtensions)
	if !ok {
		python, ok = searchPath.Which("python", conda.FileExtensions)
	}
	if ok {
		environment = append(environment, "PYTHON_EXE="+python)
	}
	environment = append(environment,
		"CONDA_DEFAULT_ENV=rcc",
		"CONDA_PREFIX="+location,
		"CONDA_PROMPT_MODIFIER=(rcc) ",
		"CONDA_SHLVL=1",
		"PYTHONHOME=",
		"PYTHONSTARTUP=",
		"PYTHONEXECUTABLE=",
		"PYTHONNOUSERSITE=1",
		"PYTHONDONTWRITEBYTECODE=x",
		"PYTHONPYCACHEPREFIX="+conda.RobocorpTemp(),
		"ROBOCORP_HOME="+common.RobocorpHome(),
		"RCC_ENVIRONMENT_HASH="+common.EnvironmentHash,
		"RCC_INSTALLATION_ID="+xviper.TrackingIdentity(),
		"RCC_TRACKING_ALLOWED="+fmt.Sprintf("%v", xviper.CanTrack()),
		"TEMP="+conda.RobocorpTemp(),
		"TMP="+conda.RobocorpTemp(),
		searchPath.AsEnvironmental("PATH"),
		it.PythonPaths().AsEnvironmental("PYTHONPATH"),
		fmt.Sprintf("ROBOT_ROOT=%s", it.WorkingDirectory()),
		fmt.Sprintf("ROBOT_ARTIFACTS=%s", it.ArtifactDirectory()),
	)
	environment = append(environment, conda.LoadActivationEnvironment(location)...)
	return environment
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

func PlainEnvironment(inject []string, full bool) []string {
	environment := make([]string, 0, 100)
	if full {
		environment = append(environment, os.Environ()...)
	}
	environment = append(environment, inject...)
	return environment
}

func LoadRobotYaml(filename string, visible bool) (Robot, error) {
	fullpath, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", filename, err)
	}
	content, err := ioutil.ReadFile(fullpath)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", fullpath, err)
	}
	if visible {
		common.Log("%q as robot.yaml is:\n%s", fullpath, string(content))
	}
	robot, err := robotFrom(content)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", fullpath, err)
	}
	robot.Root = filepath.Dir(fullpath)
	return robot, nil
}

func DetectConfigurationName(directory string) string {
	robot, err := pathlib.FindNamedPath(directory, "robot.yaml")
	if err == nil && len(robot) > 0 {
		return robot
	}
	return filepath.Join(directory, "robot.yaml")
}
