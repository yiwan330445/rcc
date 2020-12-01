package conda

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/shlex"
	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/shell"
	"github.com/robocorp/rcc/xviper"
)

func Hexdigest(raw []byte) string {
	return fmt.Sprintf("%02x", raw)
}

func metafile(folder string) string {
	return ExpandPath(folder + ".meta")
}

func metaLoad(location string) (string, error) {
	raw, err := ioutil.ReadFile(metafile(location))
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func metaSave(location, data string) error {
	return ioutil.WriteFile(metafile(location), []byte(data), 0644)
}

func touchMetafile(location string) {
	pathlib.TouchWhen(metafile(location), time.Now())
}

func LastUsed(location string) (time.Time, error) {
	return pathlib.Modtime(metafile(location))
}

func IsPristine(folder string) bool {
	digest, err := DigestFor(folder)
	if err != nil {
		return false
	}
	meta, err := metaLoad(folder)
	if err != nil {
		return false
	}
	return Hexdigest(digest) == meta
}

func reuseExistingLive(key string) bool {
	candidate := LiveFrom(key)
	if IsPristine(candidate) {
		touchMetafile(candidate)
		return true
	}
	removeClone(candidate)
	return false
}

func LiveExecution(liveFolder string, command ...string) error {
	searchPath := FindPath(liveFolder)
	commandName := command[0]
	task, ok := searchPath.Which(commandName, FileExtensions)
	if !ok {
		return fmt.Errorf("Cannot find command: %v", commandName)
	}
	common.Debug("Using %v as command %v.", task, commandName)
	command[0] = task
	environment := EnvironmentFor(liveFolder)
	_, err := shell.New(environment, ".", command...).StderrOnly().Transparent()
	return err
}

type InstallObserver map[string]bool

func (it InstallObserver) Write(content []byte) (int, error) {
	text := strings.ToLower(string(content))
	if strings.Contains(text, "safetyerror:") {
		it["safetyerror"] = true
	}
	if strings.Contains(text, "pkgs") {
		it["pkgs"] = true
	}
	if strings.Contains(text, "appears to be corrupted") {
		it["corrupted"] = true
	}
	return len(content), nil
}

func (it InstallObserver) HasFailures(targetFolder string) bool {
	if it["safetyerror"] && it["corrupted"] && len(it) > 2 {
		cloud.BackgroundMetric("rcc", "rcc.env.creation.failure", common.Version)
		removeClone(targetFolder)
		location := filepath.Join(MinicondaLocation(), "pkgs")
		common.Log("%sWARNING! Conda environment is unstable, see above error.%s", pretty.Red, pretty.Reset)
		common.Log("%sWARNING! To fix it, try to remove directory: %v%s", pretty.Red, location, pretty.Reset)
		return true
	}
	return false
}

func newLive(condaYaml, requirementsText, key string, force, freshInstall bool, postInstall []string) bool {
	targetFolder := LiveFrom(key)
	common.Debug("===  new live  ---  pre cleanup phase ===")
	removeClone(targetFolder)
	common.Debug("===  new live  ---  first try phase ===")
	success, fatal := newLiveInternal(condaYaml, requirementsText, key, force, freshInstall, postInstall)
	if !success && !force && !fatal {
		common.Debug("===  new live  ---  second try phase ===")
		common.ForceDebug()
		common.Log("Retry! First try failed ... now retrying with debug and force options!")
		removeClone(targetFolder)
		success, _ = newLiveInternal(condaYaml, requirementsText, key, true, freshInstall, postInstall)
	}
	return success
}

func newLiveInternal(condaYaml, requirementsText, key string, force, freshInstall bool, postInstall []string) (bool, bool) {
	targetFolder := LiveFrom(key)
	when := time.Now()
	if force {
		when = when.Add(-20 * 24 * time.Hour)
	}
	if force || !freshInstall {
		common.Log("rcc touching conda cache. (Stamp: %v)", when)
		SilentTouch(CondaCache(), when)
	}
	common.Debug("Setting up new conda environment using %v to folder %v", condaYaml, targetFolder)
	command := []string{CondaExecutable(), "env", "create", "-q", "-f", condaYaml, "-p", targetFolder}
	if common.DebugFlag {
		command = []string{CondaExecutable(), "env", "create", "-f", condaYaml, "-p", targetFolder}
	}
	observer := make(InstallObserver)
	common.Debug("===  new live  ---  conda env create phase ===")
	code, err := shell.New(nil, ".", command...).StderrOnly().Observed(observer, false)
	if err != nil || code != 0 {
		common.Error("Conda error", err)
		return false, false
	}
	if observer.HasFailures(targetFolder) {
		return false, true
	}
	common.Debug("Updating new environment at %v with pip requirements from %v", targetFolder, requirementsText)
	pipCommand := []string{"pip", "install", "--no-color", "--disable-pip-version-check", "--prefer-binary", "--cache-dir", PipCache(), "--find-links", WheelCache(), "--requirement", requirementsText, "--quiet"}
	if common.DebugFlag {
		pipCommand = []string{"pip", "install", "--no-color", "--disable-pip-version-check", "--prefer-binary", "--cache-dir", PipCache(), "--find-links", WheelCache(), "--requirement", requirementsText}
	}
	common.Debug("===  new live  ---  pip install phase ===")
	err = LiveExecution(targetFolder, pipCommand...)
	if err != nil {
		common.Error("Pip error", err)
		return false, false
	}
	if postInstall != nil && len(postInstall) > 0 {
		common.Debug("===  new live  ---  post install phase ===")
		for _, script := range postInstall {
			scriptCommand, err := shlex.Split(script)
			if err != nil {
				common.Log("%sScript '%s' parsing failure: %v%s", pretty.Red, script, err, pretty.Reset)
				return false, false
			}
			common.Log("Running post install script '%s' ...", script)
			err = LiveExecution(targetFolder, scriptCommand...)
			if err != nil {
				common.Log("%sScript '%s' failure: %v%s", pretty.Red, script, err, pretty.Reset)
				return false, false
			}
		}
	}
	common.Debug("===  new live  ---  finalize phase ===")
	digest, err := DigestFor(targetFolder)
	if err != nil {
		common.Error("Digest", err)
		return false, false
	}
	return metaSave(targetFolder, Hexdigest(digest)) == nil, false
}

func temporaryConfig(condaYaml, requirementsText string, filenames ...string) (string, *Environment, error) {
	var left, right *Environment
	var err error

	for _, filename := range filenames {
		left = right
		right, err = ReadCondaYaml(filename)
		if err != nil {
			return "", nil, err
		}
		if left == nil {
			continue
		}
		right, err = left.Merge(right)
		if err != nil {
			return "", nil, err
		}
	}
	yaml, err := right.AsYaml()
	if err != nil {
		return "", nil, err
	}
	common.Trace("FINAL union conda environment descriptior:\n---\n%v---", yaml)
	hash := shortDigest(yaml)
	err = right.SaveAsRequirements(requirementsText)
	if err != nil {
		return "", nil, err
	}
	pure := right.AsPureConda()
	return hash, right, pure.SaveAs(condaYaml)
}

func shortDigest(content string) string {
	digester := sha256.New()
	digester.Write([]byte(content))
	result := Hexdigest(digester.Sum(nil))
	return result[:16]
}

func NewEnvironment(force bool, configurations ...string) (string, error) {
	lockfile := MinicondaLock()
	locker, err := pathlib.Locker(lockfile, 30000)
	if err != nil {
		common.Log("Could not get lock on miniconda. Quitting!")
		return "", err
	}
	defer locker.Release()
	defer os.Remove(lockfile)

	requests := xviper.GetInt("stats.env.request") + 1
	hits := xviper.GetInt("stats.env.hit")
	dirty := xviper.GetInt("stats.env.dirty")
	misses := xviper.GetInt("stats.env.miss")
	failures := xviper.GetInt("stats.env.failures")
	merges := xviper.GetInt("stats.env.merges")
	templates := len(TemplateList())
	freshInstall := templates == 0

	defer func() {
		common.Log("####  Progress: 4/4  [Done.] [Stats: %d environments, %d requests, %d merges, %d hits, %d dirty, %d misses, %d failures | %s]", templates, requests, merges, hits, dirty, misses, failures, common.Version)
	}()
	common.Log("####  Progress: 0/4  [try use existing live same environment?] %v", xviper.TrackingIdentity())

	xviper.Set("stats.env.request", requests)

	if len(configurations) > 1 {
		merges += 1
		xviper.Set("stats.env.merges", merges)
	}

	marker := time.Now().Unix()
	condaYaml := filepath.Join(os.TempDir(), fmt.Sprintf("conda_%x.yaml", marker))
	requirementsText := filepath.Join(os.TempDir(), fmt.Sprintf("require_%x.txt", marker))
	common.Debug("Using temporary conda.yaml file: %v and requirement.txt file: %v", condaYaml, requirementsText)
	key, finalEnv, err := temporaryConfig(condaYaml, requirementsText, configurations...)
	if err != nil {
		failures += 1
		xviper.Set("stats.env.failures", failures)
		return "", err
	}
	defer os.Remove(condaYaml)
	defer os.Remove(requirementsText)

	liveFolder := LiveFrom(key)
	if reuseExistingLive(key) {
		hits += 1
		xviper.Set("stats.env.hit", hits)
		return liveFolder, nil
	}
	common.Log("####  Progress: 1/4  [try clone existing same template to live, key: %v]", key)
	if CloneFromTo(TemplateFrom(key), liveFolder) {
		dirty += 1
		xviper.Set("stats.env.dirty", dirty)
		return liveFolder, nil
	}
	common.Log("####  Progress: 2/4  [try create new environment from scratch]")
	if newLive(condaYaml, requirementsText, key, force, freshInstall, finalEnv.PostInstall) {
		misses += 1
		xviper.Set("stats.env.miss", misses)
		if !common.Liveonly {
			common.Log("####  Progress: 3/4  [backup new environment as template]")
			CloneFromTo(liveFolder, TemplateFrom(key))
		} else {
			common.Log("####  Progress: 3/4  [skipped]")
		}
		return liveFolder, nil
	}

	failures += 1
	xviper.Set("stats.env.failures", failures)
	return "", errors.New("Could not create environment.")
}

func RemoveEnvironment(label string) {
	removeClone(LiveFrom(label))
	removeClone(TemplateFrom(label))
}

func removeClone(location string) {
	os.Remove(metafile(location))
	os.RemoveAll(location)
}

func CloneFromTo(source, target string) bool {
	removeClone(target)
	os.MkdirAll(target, 0755)

	if !IsPristine(source) {
		removeClone(source)
		return false
	}
	expected, err := metaLoad(source)
	if err != nil {
		return false
	}
	success := cloneFolder(source, target, 8)
	if !success {
		removeClone(target)
		return false
	}
	digest, err := DigestFor(target)
	if err != nil || Hexdigest(digest) != expected {
		removeClone(target)
		return false
	}
	metaSave(target, expected)
	touchMetafile(source)
	return true
}

func cloneFolder(source, target string, workers int) bool {
	queue := make(chan copyRequest)
	done := make(chan bool)

	for x := 0; x < workers; x++ {
		go copyWorker(queue, done)
	}

	success := copyFolder(source, target, queue)
	close(queue)

	for x := 0; x < workers; x++ {
		<-done
	}

	return success
}

func SilentTouch(directory string, when time.Time) bool {
	handle, err := os.Open(directory)
	if err != nil {
		return false
	}
	entries, err := handle.Readdir(-1)
	handle.Close()
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			pathlib.TouchWhen(filepath.Join(directory, entry.Name()), when)
		}
	}
	return true
}

func copyFolder(source, target string, queue chan copyRequest) bool {
	os.Mkdir(target, 0755)

	handle, err := os.Open(source)
	if err != nil {
		common.Error("OPEN", err)
		return false
	}
	entries, err := handle.Readdir(-1)
	handle.Close()
	if err != nil {
		common.Error("DIR", err)
		return false
	}

	success := true
	expect := 0
	for _, entry := range entries {
		if entry.Name() == "__pycache__" {
			continue
		}
		newSource := filepath.Join(source, entry.Name())
		newTarget := filepath.Join(target, entry.Name())
		if entry.IsDir() {
			copyFolder(newSource, newTarget, queue)
		} else {
			queue <- copyRequest{newSource, newTarget}
			expect += 1
		}
	}

	return success
}

type copyRequest struct {
	source, target string
}

func copyWorker(tasks chan copyRequest, done chan bool) {
	for {
		task, ok := <-tasks
		if !ok {
			break
		}
		link, err := os.Readlink(task.source)
		if err != nil {
			pathlib.CopyFile(task.source, task.target, false)
			continue
		}
		err = os.Symlink(link, task.target)
		if err != nil {
			common.Error("LINK", err)
			continue
		}
	}

	done <- true
}
