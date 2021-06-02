package htfs

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dchest/siphash"
	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/fail"
	"github.com/robocorp/rcc/pathlib"
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
	HasBlueprint([]byte) bool
	Open(string) (io.Reader, Closer, error)
	Restore([]byte, []byte, []byte) (string, error)
}

type MutableLibrary interface {
	Library

	Identity() string
	ExactLocation(string) string
	Export([]string, string) error
	Location(string) string
	Record([]byte) error
	Stage() string
}

type hololib struct {
	identity   uint64
	basedir    string
	queryCache map[string]bool
}

func (it *hololib) Open(digest string) (readable io.Reader, closer Closer, err error) {
	return delegateOpen(it, digest)
}

func (it *hololib) Location(digest string) string {
	return filepath.Join(common.HololibLibraryLocation(), digest[:2], digest[2:4], digest[4:6])
}

func (it *hololib) ExactLocation(digest string) string {
	return filepath.Join(common.HololibLibraryLocation(), digest[:2], digest[2:4], digest[4:6], digest)
}

func (it *hololib) Identity() string {
	return fmt.Sprintf("h%016xt", it.identity)
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

func (it *hololib) Export(catalogs []string, archive string) (err error) {
	defer fail.Around(&err)

	common.Timeline("holotree export start")
	defer common.Timeline("holotree export done")

	handle, err := os.Create(archive)
	fail.On(err != nil, "Could not create archive %q.", archive)
	writer := zip.NewWriter(handle)
	defer writer.Close()

	zipper := &zipseen{
		writer,
		make(map[string]bool),
	}

	for _, name := range catalogs {
		catalog := filepath.Join(common.HololibCatalogLocation(), name)
		relative, err := filepath.Rel(common.HololibLocation(), catalog)
		fail.On(err != nil, "Could not get relative location for catalog -> %v.", err)
		err = zipper.Add(catalog, relative)
		fail.On(err != nil, "Could not add catalog to zip -> %v.", err)

		fs, err := NewRoot(".")
		fail.On(err != nil, "Could not create root location -> %v.", err)
		err = fs.LoadFrom(catalog)
		fail.On(err != nil, "Could not load catalog from %s -> %v.", catalog, err)
		err = fs.Treetop(ZipRoot(it, fs, zipper))
		fail.On(err != nil, "Could not zip catalog %s -> %v.", catalog, err)
	}
	return nil
}

func (it *hololib) Record(blueprint []byte) error {
	defer common.Stopwatch("Holotree recording took:").Debug()
	key := BlueprintHash(blueprint)
	common.Timeline("holotree record start %s", key)
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

func (it *hololib) CatalogPath(key string) string {
	name := fmt.Sprintf("%s.%s", key, Platform())
	return filepath.Join(common.HololibCatalogLocation(), name)
}

func (it *hololib) HasBlueprint(blueprint []byte) bool {
	key := BlueprintHash(blueprint)
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
	tempdir := filepath.Join(conda.RobocorpTemp(), key)
	shadow, err := NewRoot(tempdir)
	if err != nil {
		return false
	}
	err = shadow.LoadFrom(catalog)
	if err != nil {
		common.Debug("Catalog load failed, reason: %v", err)
		return false
	}
	common.Timeline("holotree content check start")
	err = shadow.Treetop(CatalogCheck(it, shadow))
	common.Timeline("holotree content check done")
	if err != nil {
		cloud.BackgroundMetric(common.ControllerIdentity(), "rcc.holotree.catalog.failure", common.Version)
		common.Debug("Catalog check failed, reason: %v", err)
		return false
	}
	return pathlib.IsFile(catalog)
}

func Catalogs() []string {
	result := make([]string, 0, 10)
	for _, catalog := range pathlib.Glob(common.HololibCatalogLocation(), "[0-9a-f]*.*") {
		result = append(result, catalog)
	}
	sort.Strings(result)
	return result
}

func Spacemap() map[string]string {
	result := make(map[string]string)
	basedir := common.HolotreeLocation()
	for _, metafile := range pathlib.Glob(basedir, "*.meta") {
		fullpath := filepath.Join(basedir, metafile)
		result[fullpath[:len(fullpath)-5]] = fullpath
	}
	return result
}

func Spaces() []*Root {
	roots := make([]*Root, 0, 20)
	for directory, metafile := range Spacemap() {
		root, err := NewRoot(directory)
		if err != nil {
			continue
		}
		err = root.LoadFrom(metafile)
		if err != nil {
			continue
		}
		roots = append(roots, root)
	}
	return roots
}

func (it *hololib) Restore(blueprint, client, tag []byte) (result string, err error) {
	defer fail.Around(&err)
	defer common.Stopwatch("Holotree restore took:").Debug()
	key := BlueprintHash(blueprint)
	common.Timeline("holotree restore start %s", key)
	prefix := textual(sipit(client), 9)
	suffix := textual(sipit(tag), 8)
	name := prefix + "_" + suffix
	metafile := filepath.Join(common.HolotreeLocation(), fmt.Sprintf("%s.meta", name))
	targetdir := filepath.Join(common.HolotreeLocation(), name)
	currentstate := make(map[string]string)
	shadow, err := NewRoot(targetdir)
	if err == nil {
		err = shadow.LoadFrom(metafile)
	}
	if err == nil {
		common.Timeline("holotree digest start")
		shadow.Treetop(DigestRecorder(currentstate))
		common.Timeline("holotree digest done")
	}
	fs, err := NewRoot(it.Stage())
	fail.On(err != nil, "Failed to create stage -> %v", err)
	err = fs.LoadFrom(it.CatalogPath(key))
	fail.On(err != nil, "Failed to load catalog %s -> %v", it.CatalogPath(key), err)
	err = fs.Relocate(targetdir)
	fail.On(err != nil, "Failed to relocate %s -> %v", targetdir, err)
	common.Timeline("holotree make branches start")
	err = fs.Treetop(MakeBranches)
	common.Timeline("holotree make branches done")
	fail.On(err != nil, "Failed to make branches -> %v", err)
	score := &stats{}
	common.Timeline("holotree restore start")
	err = fs.AllDirs(RestoreDirectory(it, fs, currentstate, score))
	fail.On(err != nil, "Failed to restore directories -> %v", err)
	common.Timeline("holotree restore done")
	defer common.Timeline("- dirty %d/%d", score.dirty, score.total)
	common.Debug("Holotree dirty workload: %d/%d\n", score.dirty, score.total)
	fs.Controller = string(client)
	fs.Space = string(tag)
	err = fs.SaveAs(metafile)
	fail.On(err != nil, "Failed to save metafile %q -> %v", metafile, err)
	return targetdir, nil
}

func BlueprintHash(blueprint []byte) string {
	return textual(sipit(blueprint), 0)
}

func sipit(key []byte) uint64 {
	return siphash.Hash(9007199254740993, 2147483647, key)
}

func textual(key uint64, size int) string {
	text := fmt.Sprintf("%016x", key)
	if size > 0 {
		return text[:size]
	}
	return text
}

func makedirs(prefix string, suffixes ...string) error {
	if common.Liveonly {
		return nil
	}
	for _, suffix := range suffixes {
		fullpath := filepath.Join(prefix, suffix)
		err := os.MkdirAll(fullpath, 0o755)
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
		identity:   sipit([]byte(identity)),
		basedir:    basedir,
		queryCache: make(map[string]bool),
	}, nil
}
