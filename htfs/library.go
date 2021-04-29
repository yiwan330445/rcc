package htfs

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dchest/siphash"
	"github.com/robocorp/rcc/common"
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

type Library interface {
	Identity() string
	Stage() string
	Record([]byte) error
	Restore([]byte, []byte, []byte) (string, error)
	Location(string) string
	ExactLocation(string) string
	HasBlueprint([]byte) bool
}

type hololib struct {
	identity uint64
	basedir  string
}

func (it *hololib) Location(digest string) string {
	return filepath.Join(common.HololibLocation(), "library", digest[:2], digest[2:4], digest[4:6])
}

func (it *hololib) ExactLocation(digest string) string {
	return filepath.Join(common.HololibLocation(), "library", digest[:2], digest[2:4], digest[4:6], digest)
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

func (it *hololib) Record(blueprint []byte) error {
	defer common.Stopwatch("Holotree recording took:").Debug()
	key := textual(sipit(blueprint), 0)
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
	fs.AllFiles(Locator(it.Identity()))
	common.Timeline("holotree (re)locator done")
	fs.Blueprint = key
	err = fs.SaveAs(filepath.Join(common.HololibLocation(), "catalog", key))
	if err != nil {
		return err
	}
	score := &stats{}
	common.Timeline("holotree lift start")
	err = fs.Treetop(ScheduleLifters(it, score))
	common.Timeline("holotree lift done")
	defer common.Timeline("- new %d/%d", score.dirty, score.total)
	common.Debug("Holotree new workload: %d/%d\n", score.dirty, score.total)
	return err
}

func (it *hololib) CatalogPath(key string) string {
	return filepath.Join(common.HololibLocation(), "catalog", key)
}

func (it *hololib) HasBlueprint(blueprint []byte) bool {
	key := textual(sipit(blueprint), 0)
	return pathlib.IsFile(it.CatalogPath(key))
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

func (it *hololib) Restore(blueprint, client, tag []byte) (string, error) {
	defer common.Stopwatch("Holotree restore took:").Debug()
	key := textual(sipit(blueprint), 0)
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
	if err != nil {
		return "", err
	}
	err = fs.LoadFrom(filepath.Join(common.HololibLocation(), "catalog", key))
	if err != nil {
		return "", err
	}
	err = fs.Relocate(targetdir)
	if err != nil {
		return "", err
	}
	common.Timeline("holotree make branches start")
	err = fs.Treetop(MakeBranches)
	common.Timeline("holotree make branches done")
	if err != nil {
		return "", err
	}
	score := &stats{}
	common.Timeline("holotree restore start")
	fs.AllDirs(RestoreDirectory(it, fs, currentstate, score))
	common.Timeline("holotree restore done")
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

func New() (Library, error) {
	err := makedirs(common.HololibLocation(), "library", "catalog")
	if err != nil {
		return nil, err
	}
	basedir := common.RobocorpHome()
	return &hololib{
		identity: sipit([]byte(basedir)),
		basedir:  basedir,
	}, nil
}

type virtual struct {
	identity uint64
	root     *Root
	registry map[string]string
	key      string
}

func Virtual() Library {
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

func (it *virtual) Record(blueprint []byte) error {
	defer common.Stopwatch("Holotree recording took:").Debug()
	key := textual(sipit(blueprint), 0)
	common.Timeline("holotree record start %s (virtual)", key)
	fs, err := NewRoot(it.Stage())
	if err != nil {
		return err
	}
	err = fs.Lift()
	if err != nil {
		return err
	}
	common.Timeline("holotree (re)locator start (virtual)")
	fs.AllFiles(Locator(it.Identity()))
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
	key := textual(sipit(blueprint), 0)
	common.Timeline("holotree restore start %s (virtual)", key)
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
	fs.AllDirs(RestoreDirectory(it, fs, currentstate, score))
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

func (it *virtual) ExactLocation(key string) string {
	return it.registry[key]
}

func (it *virtual) Location(key string) string {
	panic("Location is not supported on virtual holotree.")
}

func (it *virtual) HasBlueprint(blueprint []byte) bool {
	return it.key == textual(sipit(blueprint), 0)
}
