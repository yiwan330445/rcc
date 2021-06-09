package htfs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/fail"
	"github.com/robocorp/rcc/journal"
	"github.com/robocorp/rcc/pathlib"
)

type virtual struct {
	identity uint64
	root     *Root
	registry map[string]string
	key      string
}

func Virtual() MutableLibrary {
	return &virtual{
		identity: sipit([]byte(common.RobocorpHome())),
	}
}

func (it *virtual) Identity() string {
	return fmt.Sprintf("v%016xh", it.identity)
}

func (it *virtual) Stage() string {
	stage := filepath.Join(common.HolotreeLocation(), it.Identity())
	err := os.MkdirAll(stage, 0o755)
	if err != nil {
		panic(err)
	}
	return stage
}

func (it *virtual) Export([]string, string) error {
	return fmt.Errorf("Not supported yet on virtual holotree.")
}

func (it *virtual) Record(blueprint []byte) (err error) {
	defer fail.Around(&err)
	defer common.Stopwatch("Holotree recording took:").Debug()
	key := BlueprintHash(blueprint)
	common.Timeline("holotree record start %s (virtual)", key)
	fs, err := NewRoot(it.Stage())
	fail.On(err != nil, "Failed to create stage root: %v", err)
	err = fs.Lift()
	fail.On(err != nil, "Failed to lift structure out of stage: %v", err)
	common.Timeline("holotree (re)locator start (virtual)")
	err = fs.AllFiles(Locator(it.Identity()))
	fail.On(err != nil, "Failed to apply relocate to stage: %v", err)
	common.Timeline("holotree (re)locator done (virtual)")
	it.registry = make(map[string]string)
	fs.Treetop(DigestMapper(it.registry))
	fs.Blueprint = key
	it.root = fs
	it.key = key
	return nil
}

func (it *virtual) Restore(blueprint, client, tag []byte) (string, error) {
	defer common.Stopwatch("Holotree restore took:").Debug()
	key := BlueprintHash(blueprint)
	common.Timeline("holotree restore start %s (virtual)", key)
	prefix := textual(sipit(client), 9)
	suffix := textual(sipit(tag), 8)
	name := prefix + "_" + suffix
	metafile := filepath.Join(common.HolotreeLocation(), fmt.Sprintf("%s.meta", name))
	targetdir := filepath.Join(common.HolotreeLocation(), name)
	lockfile := filepath.Join(common.HolotreeLocation(), fmt.Sprintf("%s.lck", name))
	locker, err := pathlib.Locker(lockfile, 30000)
	fail.On(err != nil, "Could not get lock for %s. Quiting.", targetdir)
	defer locker.Release()
	journal.Post("space-used", metafile, "virutal holotree with blueprint %s", key)
	currentstate := make(map[string]string)
	shadow, err := NewRoot(targetdir)
	if err == nil {
		err = shadow.LoadFrom(metafile)
	}
	if err == nil {
		common.Timeline("holotree digest start (virtual)")
		shadow.Treetop(DigestRecorder(currentstate))
		common.Timeline("holotree digest done (virtual)")
	}
	fs := it.root
	err = fs.Relocate(targetdir)
	if err != nil {
		return "", err
	}
	common.Timeline("holotree make branches start (virtual)")
	err = fs.Treetop(MakeBranches)
	common.Timeline("holotree make branches done (virtual)")
	if err != nil {
		return "", err
	}
	score := &stats{}
	common.Timeline("holotree restore start (virtual)")
	err = fs.AllDirs(RestoreDirectory(it, fs, currentstate, score))
	if err != nil {
		return "", err
	}
	common.Timeline("holotree restore done (virtual)")
	defer common.Timeline("- dirty %d/%d", score.dirty, score.total)
	common.Debug("Holotree dirty workload: %d/%d\n", score.dirty, score.total)
	fs.Controller = string(client)
	fs.Space = string(tag)
	err = fs.SaveAs(metafile)
	if err != nil {
		return "", err
	}
	return targetdir, nil
}

func (it *virtual) Open(digest string) (readable io.Reader, closer Closer, err error) {
	return delegateOpen(it, digest)
}

func (it *virtual) ExactLocation(key string) string {
	return it.registry[key]
}

func (it *virtual) Location(key string) string {
	panic("Location is not supported on virtual holotree.")
}

func (it *virtual) HasBlueprint(blueprint []byte) bool {
	return it.key == BlueprintHash(blueprint)
}
