package conda

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
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
	if common.Stageonly {
		return nil
	}
	return ioutil.WriteFile(metafile(location), []byte(data), 0644)
}

func touchMetafile(location string) {
	pathlib.TouchWhen(metafile(location), time.Now())
}

func LastUsed(location string) (time.Time, error) {
	return pathlib.Modtime(metafile(location))
}

func IsPristine(folder string) bool {
	digest, err := DigestFor(folder, nil)
	if err != nil {
		return false
	}
	meta, err := metaLoad(folder)
	if err != nil {
		return false
	}
	return Hexdigest(digest) == meta
}

func reuseExistingLive(key string) (bool, error) {
	if common.Stageonly {
		return false, nil
	}
	candidate := LiveFrom(key)
	if IsPristine(candidate) {
		touchMetafile(candidate)
		return true, nil
	}
	err := removeClone(candidate)
	if err != nil {
		return false, err
	}
	return false, nil
}

func livePrepare(liveFolder string, command ...string) (*shell.Task, error) {
	searchPath := FindPath(liveFolder)
	commandName := command[0]
	task, ok := searchPath.Which(commandName, FileExtensions)
	if !ok {
		return nil, fmt.Errorf("Cannot find command: %v", commandName)
	}
	common.Debug("Using %v as command %v.", task, commandName)
	command[0] = task
	environment := EnvironmentFor(liveFolder)
	return shell.New(environment, ".", command...), nil
}

func LiveCapture(liveFolder string, command ...string) (string, int, error) {
	task, err := livePrepare(liveFolder, command...)
	if err != nil {
		return "", 9999, err
	}
	return task.CaptureOutput()
}

func LiveExecution(sink *os.File, liveFolder string, command ...string) error {
	defer sink.Sync()
	fmt.Fprintf(sink, "Command %q at %q:\n", command, liveFolder)
	task, err := livePrepare(liveFolder, command...)
	if err != nil {
		return err
	}
	_, err = task.Tracked(sink, false)
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
		cloud.BackgroundMetric(common.ControllerIdentity(), "rcc.env.creation.failure", common.Version)
		removeClone(targetFolder)
		location := filepath.Join(RobocorpHome(), "pkgs")
		common.Log("%sWARNING! Conda environment is unstable, see above error.%s", pretty.Red, pretty.Reset)
		common.Log("%sWARNING! To fix it, try to remove directory: %v%s", pretty.Red, location, pretty.Reset)
		return true
	}
	return false
}

func newLive(yaml, condaYaml, requirementsText, key string, force, freshInstall bool, postInstall []string) (bool, error) {
	targetFolder := LiveFrom(key)
	common.Debug("===  new live  ---  pre cleanup phase ===")
	common.Timeline("pre cleanup phase.")
	err := removeClone(targetFolder)
	if err != nil {
		return false, err
	}
	common.Debug("===  new live  ---  first try phase ===")
	common.Timeline("first try.")
	success, fatal := newLiveInternal(yaml, condaYaml, requirementsText, key, force, freshInstall, postInstall)
	if !success && !force && !fatal {
		common.Debug("===  new live  ---  second try phase ===")
		common.Timeline("second try.")
		common.ForceDebug()
		common.Log("Retry! First try failed ... now retrying with debug and force options!")
		err = removeClone(targetFolder)
		if err != nil {
			return false, err
		}
		success, _ = newLiveInternal(yaml, condaYaml, requirementsText, key, true, freshInstall, postInstall)
	}
	return success, nil
}

func newLiveInternal(yaml, condaYaml, requirementsText, key string, force, freshInstall bool, postInstall []string) (bool, bool) {
	targetFolder := LiveFrom(key)
	planfile := fmt.Sprintf("%s.plan", targetFolder)
	planWriter, err := os.OpenFile(planfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return false, false
	}
	defer func() {
		planWriter.Close()
		os.Remove(planfile)
	}()
	fmt.Fprintf(planWriter, "---  installation plan %q %s [force: %v, fresh: %v]  ---\n\n", key, time.Now().Format(time.RFC3339), force, freshInstall)
	stopwatch := common.Stopwatch("installation plan")
	fmt.Fprintf(planWriter, "---  plan blueprint @%ss  ---\n\n", stopwatch)
	fmt.Fprintf(planWriter, "%s\n", yaml)

	common.Debug("Setting up new conda environment using %v to folder %v", condaYaml, targetFolder)
	ttl := "57600"
	if force {
		ttl = "0"
	}
	command := []string{BinMicromamba(), "create", "--extra-safety-checks", "fail", "--retry-with-clean-cache", "--strict-channel-priority", "--repodata-ttl", ttl, "-q", "-y", "-f", condaYaml, "-p", targetFolder}
	if true || common.DebugFlag {
		command = []string{BinMicromamba(), "create", "--extra-safety-checks", "fail", "--retry-with-clean-cache", "--strict-channel-priority", "--repodata-ttl", ttl, "-y", "-f", condaYaml, "-p", targetFolder}
	}
	observer := make(InstallObserver)
	common.Debug("===  new live  ---  micromamba create phase ===")
	common.Timeline("Micromamba start.")
	fmt.Fprintf(planWriter, "\n---  micromamba plan @%ss  ---\n\n", stopwatch)
	tee := io.MultiWriter(observer, planWriter)
	code, err := shell.New(CondaEnvironment(), ".", command...).Tracked(tee, false)
	if err != nil || code != 0 {
		common.Timeline("micromamba fail.")
		common.Fatal("Micromamba", err)
		return false, false
	}
	common.Timeline("micromamba done.")
	if observer.HasFailures(targetFolder) {
		return false, true
	}
	fmt.Fprintf(planWriter, "\n---  pip plan @%ss  ---\n\n", stopwatch)
	pipCache, wheelCache := PipCache(), WheelCache()
	size, ok := pathlib.Size(requirementsText)
	if !ok || size == 0 {
		common.Log("####  Progress: 4/6  [pip install phase skipped -- no pip dependencies]")
		common.Timeline("4/6 no pip.")
	} else {
		common.Log("####  Progress: 4/6  [pip install phase]")
		common.Timeline("4/6 pip install start.")
		common.Debug("Updating new environment at %v with pip requirements from %v (size: %v)", targetFolder, requirementsText, size)
		pipCommand := []string{"pip", "install", "--no-color", "--disable-pip-version-check", "--prefer-binary", "--cache-dir", pipCache, "--find-links", wheelCache, "--requirement", requirementsText, "--quiet"}
		if true || common.DebugFlag {
			pipCommand = []string{"pip", "install", "--no-color", "--disable-pip-version-check", "--prefer-binary", "--cache-dir", pipCache, "--find-links", wheelCache, "--requirement", requirementsText}
		}
		common.Debug("===  new live  ---  pip install phase ===")
		err = LiveExecution(planWriter, targetFolder, pipCommand...)
		if err != nil {
			common.Timeline("pip fail.")
			common.Fatal("Pip", err)
			return false, false
		}
		common.Timeline("pip done.")
	}
	fmt.Fprintf(planWriter, "\n---  post install plan @%ss  ---\n\n", stopwatch)
	if postInstall != nil && len(postInstall) > 0 {
		common.Timeline("post install.")
		common.Debug("===  new live  ---  post install phase ===")
		for _, script := range postInstall {
			scriptCommand, err := shlex.Split(script)
			if err != nil {
				common.Fatal("post-install", err)
				common.Log("%sScript '%s' parsing failure: %v%s", pretty.Red, script, err, pretty.Reset)
				return false, false
			}
			common.Debug("Running post install script '%s' ...", script)
			err = LiveExecution(planWriter, targetFolder, scriptCommand...)
			if err != nil {
				common.Fatal("post-install", err)
				common.Log("%sScript '%s' failure: %v%s", pretty.Red, script, err, pretty.Reset)
				return false, false
			}
		}
	}
	common.Debug("===  new live  ---  activate phase ===")
	fmt.Fprintf(planWriter, "\n---  activation plan @%ss  ---\n\n", stopwatch)
	err = Activate(planWriter, targetFolder)
	if err != nil {
		common.Log("%sActivation failure: %v%s", pretty.Yellow, err, pretty.Reset)
	}
	for _, line := range LoadActivationEnvironment(targetFolder) {
		fmt.Fprintf(planWriter, "%s\n", line)
	}
	fmt.Fprintf(planWriter, "\n---  installation plan complete @%ss  ---\n\n", stopwatch)
	planWriter.Sync()
	planWriter.Close()
	finalplan := filepath.Join(targetFolder, "rcc_plan.log")
	os.Rename(planfile, finalplan)
	common.Log("%sInstallation plan is: %v%s", pretty.Yellow, finalplan, pretty.Reset)
	common.Debug("===  new live  ---  finalize phase ===")

	markerFile := filepath.Join(targetFolder, "identity.yaml")
	err = ioutil.WriteFile(markerFile, []byte(yaml), 0o644)
	if err != nil {
		return false, false
	}

	digest, err := DigestFor(targetFolder, nil)
	if err != nil {
		common.Fatal("Digest", err)
		return false, false
	}
	return metaSave(targetFolder, Hexdigest(digest)) == nil, false
}

func temporaryConfig(condaYaml, requirementsText string, save bool, filenames ...string) (string, string, *Environment, error) {
	var left, right *Environment
	var err error

	for _, filename := range filenames {
		left = right
		right, err = ReadCondaYaml(filename)
		if err != nil {
			return "", "", nil, err
		}
		if left == nil {
			continue
		}
		right, err = left.Merge(right)
		if err != nil {
			return "", "", nil, err
		}
	}
	yaml, err := right.AsYaml()
	if err != nil {
		return "", "", nil, err
	}
	common.Log("FINAL union conda environment descriptior:\n---\n%v---", yaml)
	hash := shortDigest(yaml)
	if !save {
		return hash, yaml, right, nil
	}
	err = right.SaveAsRequirements(requirementsText)
	if err != nil {
		return "", "", nil, err
	}
	pure := right.AsPureConda()
	err = pure.SaveAs(condaYaml)
	return hash, yaml, right, err
}

func shortDigest(content string) string {
	digester := sha256.New()
	digester.Write([]byte(content))
	result := Hexdigest(digester.Sum(nil))
	return result[:16]
}

func CalculateComboHash(configurations ...string) (string, error) {
	key, _, _, err := temporaryConfig("/dev/null", "/dev/null", false, configurations...)
	if err != nil {
		return "", err
	}
	return key, nil
}

func NewEnvironment(force bool, configurations ...string) (string, error) {
	common.Timeline("New environment.")
	cloud.BackgroundMetric(common.ControllerIdentity(), "rcc.env.create.start", common.Version)

	lockfile := RobocorpLock()
	locker, err := pathlib.Locker(lockfile, 30000)
	if err != nil {
		common.Log("Could not get lock on live environment. Quitting!")
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
		templates = len(TemplateList())
		common.Log("####  Progress: 6/6  [Done.] [Cache statistics: %d environments, %d requests, %d merges, %d hits, %d dirty, %d misses, %d failures | %s]", templates, requests, merges, hits, dirty, misses, failures, common.Version)
		common.Timeline("6/6 Done.")
	}()
	common.Log("####  Progress: 0/6  [try use existing live same environment?] %v", xviper.TrackingIdentity())
	common.Timeline("0/6 Existing.")

	xviper.Set("stats.env.request", requests)

	if len(configurations) > 1 {
		merges += 1
		xviper.Set("stats.env.merges", merges)
	}

	condaYaml := filepath.Join(os.TempDir(), fmt.Sprintf("conda_%x.yaml", common.When))
	requirementsText := filepath.Join(os.TempDir(), fmt.Sprintf("require_%x.txt", common.When))
	common.Debug("Using temporary conda.yaml file: %v and requirement.txt file: %v", condaYaml, requirementsText)
	key, yaml, finalEnv, err := temporaryConfig(condaYaml, requirementsText, true, configurations...)
	if err != nil {
		failures += 1
		xviper.Set("stats.env.failures", failures)
		return "", err
	}
	defer os.Remove(condaYaml)
	defer os.Remove(requirementsText)
	common.Log("####  Progress: 1/6  [environment key is: %s]", key)
	common.Timeline("1/6 key %s.", key)

	common.EnvironmentHash = key

	quickFolder, ok, err := LeaseInterceptor(key)
	if ok {
		return quickFolder, nil
	}
	if err != nil {
		return "", err
	}

	liveFolder := LiveFrom(key)
	after := make(map[string]string)
	afterHash, afterErr := DigestFor(liveFolder, after)
	reusable, err := reuseExistingLive(key)
	if err != nil {
		return "", err
	}
	if reusable {
		hits += 1
		xviper.Set("stats.env.hit", hits)
		return liveFolder, nil
	}
	if common.Stageonly {
		common.Log("####  Progress: 2/6  [skipped -- stage only]")
		common.Timeline("2/6 stage only.")
	} else {
		templateFolder := TemplateFrom(key)
		if IsPristine(templateFolder) {
			before := make(map[string]string)
			beforeHash, beforeErr := DigestFor(templateFolder, before)
			DiagnoseDirty(templateFolder, liveFolder, beforeHash, afterHash, beforeErr, afterErr, before, after, false)
		}
		common.Log("####  Progress: 2/6  [try clone existing same template to live, key: %v]", key)
		common.Timeline("2/6 base to live.")
		success, err := CloneFromTo(templateFolder, liveFolder, pathlib.CopyFile)
		if err != nil {
			return "", err
		}
		if success {
			dirty += 1
			xviper.Set("stats.env.dirty", dirty)
			return liveFolder, nil
		}
	}
	common.Log("####  Progress: 3/6  [try create new environment from scratch]")
	common.Timeline("3/6 env from scratch.")
	success, err := newLive(yaml, condaYaml, requirementsText, key, force, freshInstall, finalEnv.PostInstall)
	if err != nil {
		return "", err
	}
	if success {
		misses += 1
		xviper.Set("stats.env.miss", misses)
		if common.Liveonly {
			common.Log("####  Progress: 5/6  [skipped -- live only]")
			common.Timeline("5/6 live only.")
		} else {
			common.Log("####  Progress: 5/6  [backup new environment as template]")
			common.Timeline("5/6 backup to base.")
			_, err = CloneFromTo(liveFolder, TemplateFrom(key), pathlib.CopyFile)
			if err != nil {
				return "", err
			}
		}
		return liveFolder, nil
	}

	failures += 1
	xviper.Set("stats.env.failures", failures)
	return "", errors.New("Could not create environment.")
}

func RemoveEnvironment(label string) error {
	if IsLeasedEnvironment(label) {
		return fmt.Errorf("WARNING: %q is leased by %q and wont be deleted!", label, WhoLeased(label))
	}
	err := removeClone(LiveFrom(label))
	if err != nil {
		return err
	}
	return removeClone(TemplateFrom(label))
}

func removeClone(location string) error {
	if !pathlib.IsDir(location) {
		return nil
	}
	randomLocation := fmt.Sprintf("%s.%08X", location, rand.Uint32())
	common.Debug("Rename/remove %q using %q as random name.", location, randomLocation)
	err := os.Rename(location, randomLocation)
	if err != nil {
		common.Log("Rename %q -> %q failed as: %v!", location, randomLocation, err)
		return err
	}
	err = os.RemoveAll(randomLocation)
	if err != nil {
		common.Log("Removal of %q failed as: %v!", randomLocation, err)
		return err
	}
	meta := metafile(location)
	if pathlib.IsFile(meta) {
		return os.Remove(meta)
	}
	return nil
}

func CloneFromTo(source, target string, copier pathlib.Copier) (bool, error) {
	err := removeClone(target)
	if err != nil {
		return false, err
	}
	os.MkdirAll(target, 0755)

	if !IsPristine(source) {
		if common.Liveonly {
			common.Debug("Clone source %q is dirty, but wont remove since --liveonly flag.", source)
		} else {
			err = removeClone(source)
			if err != nil {
				return false, fmt.Errorf("Source %q is not pristine! And could not remove: %v", source, err)
			}
		}
		return false, nil
	}
	expected, err := metaLoad(source)
	if err != nil {
		return false, nil
	}
	success := cloneFolder(source, target, 8, copier)
	if !success {
		err = removeClone(target)
		if err != nil {
			return false, fmt.Errorf("Cloning %q to %q failed! And cleanup failed: %v", source, target, err)
		}
		return false, nil
	}
	digest, err := DigestFor(target, nil)
	if err != nil || Hexdigest(digest) != expected {
		err = removeClone(target)
		if err != nil {
			return false, fmt.Errorf("Target %q does not match source %q! And cleanup failed: %v!", target, source, err)
		}
		return false, nil
	}
	metaSave(target, expected)
	touchMetafile(source)
	return true, nil
}

func cloneFolder(source, target string, workers int, copier pathlib.Copier) bool {
	queue := make(chan copyRequest)
	done := make(chan bool)

	for x := 0; x < workers; x++ {
		go copyWorker(queue, done, copier)
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

func copyWorker(tasks chan copyRequest, done chan bool, copier pathlib.Copier) {
	for {
		task, ok := <-tasks
		if !ok {
			break
		}
		link, err := os.Readlink(task.source)
		if err != nil {
			copier(task.source, task.target, false)
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
