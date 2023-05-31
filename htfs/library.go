package htfs

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/fail"
	"github.com/robocorp/rcc/journal"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/set"
)

const (
	epoc = 1610000000
)

var (
	motherTime = time.Unix(epoc, 0)
)

type stats struct {
	sync.Mutex
	total uint64
	dirty uint64
}

func (it *stats) Dirtyness() float64 {
	dirtyness := (1000 * it.dirty) / it.total
	return float64(dirtyness) / 10.0
}

func (it *stats) Dirty(dirty bool) {
	it.Lock()
	defer it.Unlock()

	it.total++
	if dirty {
		it.dirty++
	}
}

type Closer func() error

type Library interface {
	ValidateBlueprint([]byte) error
	HasBlueprint([]byte) bool
	Open(string) (io.Reader, Closer, error)
	TargetDir([]byte, []byte, []byte) (string, error)
	Restore([]byte, []byte, []byte) (string, error)
}

type MutableLibrary interface {
	Library

	Identity() string
	ExactLocation(string) string
	Export([]string, []string, string) error
	Remove([]string) error
	Location(string) string
	Record([]byte) error
	Stage() string
	CatalogPath(string) string
	WriteIdentity([]byte) error
}

type hololib struct {
	identity   uint64
	basedir    string
	queryCache map[string]bool
}

func (it *hololib) Open(digest string) (readable io.Reader, closer Closer, err error) {
	return delegateOpen(it, digest, true)
}

func (it *hololib) Location(digest string) string {
	return filepath.Join(common.HololibLibraryLocation(), digest[:2], digest[2:4], digest[4:6])
}

func ExactDefaultLocation(digest string) string {
	return filepath.Join(common.HololibLibraryLocation(), digest[:2], digest[2:4], digest[4:6], digest)
}

func RelativeDefaultLocation(digest string) string {
	location := ExactDefaultLocation(digest)
	relative, _ := filepath.Rel(common.HololibLocation(), location)
	return relative
}

func (it *hololib) ExactLocation(digest string) string {
	return ExactDefaultLocation(digest)
}

func (it *hololib) Identity() string {
	suffix := fmt.Sprintf("%016x", it.identity)
	return fmt.Sprintf("h%s_%st", common.UserHomeIdentity(), suffix[:14])
}

func (it *hololib) WriteIdentity(yaml []byte) error {
	markerFile := filepath.Join(it.Stage(), "identity.yaml")
	return pathlib.WriteFile(markerFile, yaml, 0o644)
}

func (it *hololib) Stage() string {
	stage := filepath.Join(common.HolotreeLocation(), it.Identity())
	err := os.MkdirAll(stage, 0o755)
	if err != nil {
		panic(err)
	}
	return stage
}

type zipseen struct {
	*zip.Writer
	seen map[string]bool
}

func (it zipseen) Ignore(relativepath string) {
	it.seen[relativepath] = true
}

func (it zipseen) Add(fullpath, relativepath string) (err error) {
	defer fail.Around(&err)

	if it.seen[relativepath] {
		return nil
	}
	it.seen[relativepath] = true

	source, err := os.Open(fullpath)
	fail.On(err != nil, "Could not open: %q -> %v", fullpath, err)
	defer source.Close()
	target, err := it.Create(relativepath)
	fail.On(err != nil, "Could not create: %q -> %v", relativepath, err)
	_, err = io.Copy(target, source)
	fail.On(err != nil, "Copy failure: %q -> %q -> %v", fullpath, relativepath, err)
	return nil
}

func (it *hololib) Remove(catalogs []string) (err error) {
	defer fail.Around(&err)

	common.TimelineBegin("holotree remove start")
	defer common.TimelineEnd()

	for _, name := range catalogs {
		catalog := filepath.Join(common.HololibCatalogLocation(), name)
		if !pathlib.IsFile(catalog) {
			pretty.Warning("Catalog %s (%s) is not a file! Ignored!", name, catalog)
			continue
		}
		err := os.Remove(catalog)
		fail.On(err != nil, "Could not remove catalog %s [filename: %q]", name, catalog)
	}
	return nil
}

func (it *hololib) Export(catalogs, known []string, archive string) (err error) {
	defer fail.Around(&err)

	common.TimelineBegin("holotree export start")
	defer common.TimelineEnd()

	handle, err := pathlib.Create(archive)
	fail.On(err != nil, "Could not create archive %q.", archive)
	writer := zip.NewWriter(handle)
	defer writer.Close()

	zipper := &zipseen{
		writer,
		make(map[string]bool),
	}

	exported := false

	for _, name := range known {
		catalog := filepath.Join(common.HololibCatalogLocation(), name)
		fs, err := NewRoot(".")
		fail.On(err != nil, "Could not create root location -> %v.", err)
		err = fs.LoadFrom(catalog)
		if err != nil {
			continue
		}
		err = fs.Treetop(ZipIgnore(it, fs, zipper))
		fail.On(err != nil, "Could not ignore catalog %s -> %v.", catalog, err)
	}

	for _, name := range catalogs {
		catalog := filepath.Join(common.HololibCatalogLocation(), name)

		fs, err := NewRoot(".")
		fail.On(err != nil, "Could not create root location -> %v.", err)
		err = fs.LoadFrom(catalog)
		fail.On(err != nil, "Could not load catalog from %s -> %v.", catalog, err)
		err = fs.Treetop(ZipRoot(it, fs, zipper))
		fail.On(err != nil, "Could not zip catalog %s -> %v.", catalog, err)

		relative, err := filepath.Rel(common.HololibLocation(), catalog)
		fail.On(err != nil, "Could not get relative location for catalog -> %v.", err)
		err = zipper.Add(catalog, relative)
		fail.On(err != nil, "Could not add catalog to zip -> %v.", err)

		exported = true
	}
	fail.On(!exported, "None of given catalogs were available for export!")
	return nil
}

func (it *hololib) Record(blueprint []byte) error {
	defer common.Stopwatch("Holotree recording took:").Debug()
	err := it.WriteIdentity(blueprint)
	if err != nil {
		return err
	}
	key := common.BlueprintHash(blueprint)
	common.TimelineBegin("holotree record start %s", key)
	defer common.TimelineEnd()
	fs, err := NewRoot(it.Stage())
	if err != nil {
		return err
	}
	err = fs.Lift()
	if err != nil {
		return err
	}
	common.Timeline("holotree (re)locator start")
	err = fs.AllFiles(Locator(it.Identity()))
	if err != nil {
		return err
	}
	common.Timeline("holotree (re)locator done")
	fs.Blueprint = key
	catalog := it.CatalogPath(key)
	err = fs.SaveAs(catalog)
	if err != nil {
		return err
	}
	score := &stats{}
	common.Timeline("holotree lift start %q", catalog)
	err = fs.Treetop(ScheduleLifters(it, score))
	common.Timeline("holotree lift done")
	defer common.Timeline("- new %d/%d", score.dirty, score.total)
	common.Debug("Holotree new workload: %d/%d\n", score.dirty, score.total)
	return err
}

func CatalogName(key string) string {
	return fmt.Sprintf("%sv12.%s", key, common.Platform())
}

func (it *hololib) CatalogPath(key string) string {
	return filepath.Join(common.HololibCatalogLocation(), CatalogName(key))
}

func (it *hololib) ValidateBlueprint(blueprint []byte) error {
	return nil
}

func (it *hololib) HasBlueprint(blueprint []byte) bool {
	key := common.BlueprintHash(blueprint)
	found, ok := it.queryCache[key]
	if !ok {
		found = it.queryBlueprint(key)
		it.queryCache[key] = found
	}
	return found
}

func (it *hololib) queryBlueprint(key string) bool {
	common.Timeline("holotree blueprint query")
	catalog := it.CatalogPath(key)
	if !pathlib.IsFile(catalog) {
		return false
	}
	tempdir := filepath.Join(common.RobocorpTemp(), key)
	shadow, err := NewRoot(tempdir)
	if err != nil {
		return false
	}
	err = shadow.LoadFrom(catalog)
	if err != nil {
		common.Debug("Catalog load failed, reason: %v", err)
		return false
	}
	common.TimelineBegin("holotree content check start")
	err = shadow.Treetop(CatalogCheck(it, shadow))
	common.TimelineEnd()
	if err != nil {
		cloud.BackgroundMetric(common.ControllerIdentity(), "rcc.holotree.catalog.failure", common.Version)
		common.Debug("Catalog check failed, reason: %v", err)
		return false
	}
	return pathlib.IsFile(catalog)
}

func CatalogNames() []string {
	result := make([]string, 0, 10)
	for _, catalog := range pathlib.Glob(common.HololibCatalogLocation(), "[0-9a-f]*v12.*") {
		result = append(result, filepath.Base(catalog))
	}
	return set.Set(result)
}

func ControllerSpaceName(client, tag []byte) string {
	prefix := common.Textual(common.Sipit(client), 7)
	suffix := common.Textual(common.Sipit(tag), 8)
	return common.UserHomeIdentity() + "_" + prefix + "_" + suffix
}

func touchUsedHash(hash string) {
	filename := fmt.Sprintf("%s.%s", hash, common.UserHomeIdentity())
	fullpath := filepath.Join(common.HololibUsageLocation(), filename)
	pathlib.ForceTouchWhen(fullpath, common.ProgressMark)
}

func (it *hololib) TargetDir(blueprint, controller, space []byte) (result string, err error) {
	defer fail.Around(&err)
	key := common.BlueprintHash(blueprint)
	catalog := it.CatalogPath(key)
	fs, err := NewRoot(it.Stage())
	fail.On(err != nil, "Failed to create stage -> %v", err)
	name := ControllerSpaceName(controller, space)
	err = fs.LoadFrom(catalog)
	if err != nil {
		return filepath.Join(common.HolotreeLocation(), name), nil
	}
	return filepath.Join(fs.HolotreeBase(), name), nil
}

func UserHolotreeLockfile() string {
	name := ControllerSpaceName([]byte(common.ControllerIdentity()), []byte(common.HolotreeSpace))
	return filepath.Join(common.HolotreeLocation(), fmt.Sprintf("%s.lck", name))
}

func (it *hololib) Restore(blueprint, client, tag []byte) (result string, err error) {
	defer fail.Around(&err)
	defer common.Stopwatch("Holotree restore took:").Debug()
	key := common.BlueprintHash(blueprint)
	catalog := it.CatalogPath(key)
	common.TimelineBegin("holotree space restore start [%s]", key)
	defer common.TimelineEnd()
	name := ControllerSpaceName(client, tag)
	fs, err := NewRoot(it.Stage())
	fail.On(err != nil, "Failed to create stage -> %v", err)
	err = fs.LoadFrom(catalog)
	fail.On(err != nil, "Failed to load catalog %s -> %v", catalog, err)
	metafile := filepath.Join(fs.HolotreeBase(), fmt.Sprintf("%s.meta", name))
	targetdir := filepath.Join(fs.HolotreeBase(), name)
	lockfile := filepath.Join(fs.HolotreeBase(), fmt.Sprintf("%s.lck", name))
	completed := pathlib.LockWaitMessage(lockfile, "Serialized holotree restore [holotree base lock]")
	locker, err := pathlib.Locker(lockfile, 30000)
	completed()
	fail.On(err != nil, "Could not get lock for %s. Quiting.", targetdir)
	defer locker.Release()
	journal.Post("space-used", metafile, "normal holotree with blueprint %s from %s", key, catalog)
	currentstate := make(map[string]string)
	mode := fmt.Sprintf("new space for %q", key)
	shadow, err := NewRoot(targetdir)
	if err == nil {
		err = shadow.LoadFrom(metafile)
	}
	if err == nil {
		if key == shadow.Blueprint {
			mode = fmt.Sprintf("cleaned up space for %q", key)
		} else {
			mode = fmt.Sprintf("coverted space from %q to %q", shadow.Blueprint, key)
		}
		common.TimelineBegin("holotree digest start [%q -> %q]", shadow.Blueprint, key)
		shadow.Treetop(DigestRecorder(currentstate))
		common.TimelineEnd()
	}
	common.Timeline("mode: %s", mode)
	common.Debug("Holotree operating mode is: %s", mode)
	err = fs.Relocate(targetdir)
	fail.On(err != nil, "Failed to relocate %s -> %v", targetdir, err)
	common.TimelineBegin("holotree make branches start")
	err = fs.Treetop(MakeBranches)
	common.TimelineEnd()
	fail.On(err != nil, "Failed to make branches -> %v", err)
	score := &stats{}
	common.TimelineBegin("holotree restore start")
	err = fs.AllDirs(RestoreDirectory(it, fs, currentstate, score))
	fail.On(err != nil, "Failed to restore directories -> %v", err)
	common.TimelineEnd()
	defer common.Timeline("- dirty %d/%d", score.dirty, score.total)
	common.Debug("Holotree dirty workload: %d/%d\n", score.dirty, score.total)
	journal.CurrentBuildEvent().Dirty(score.Dirtyness())
	fs.Controller = string(client)
	fs.Space = string(tag)
	err = fs.SaveAs(metafile)
	fail.On(err != nil, "Failed to save metafile %q -> %v", metafile, err)
	pathlib.TouchWhen(catalog, time.Now())
	planfile := filepath.Join(targetdir, "rcc_plan.log")
	if pathlib.FileExist(planfile) {
		common.Log("%sInstallation plan is: %v%s", pretty.Yellow, planfile, pretty.Reset)
	}
	identityfile := filepath.Join(targetdir, "identity.yaml")
	if pathlib.FileExist(identityfile) {
		common.Log("%sEnvironment configuration descriptor is: %v%s", pretty.Yellow, identityfile, pretty.Reset)
	}
	touchUsedHash(key)
	return targetdir, nil
}

func makedirs(prefix string, suffixes ...string) error {
	if common.Liveonly {
		return nil
	}
	for _, suffix := range suffixes {
		fullpath := filepath.Join(prefix, suffix)
		_, err := pathlib.MakeSharedDir(fullpath)
		if err != nil {
			return err
		}
	}
	return nil
}

func New() (MutableLibrary, error) {
	err := makedirs(common.HololibLocation(), "library", "catalog")
	if err != nil {
		return nil, err
	}
	basedir := common.RobocorpHome()
	identity := strings.ToLower(fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH))
	return &hololib{
		identity:   common.Sipit([]byte(identity)),
		basedir:    basedir,
		queryCache: make(map[string]bool),
	}, nil
}
