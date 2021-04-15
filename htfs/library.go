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
	Spaces() []*Root
	Record([]byte) error
	Restore([]byte, []byte, []byte) (string, error)
	Location(string) string
	HasBlueprint([]byte) bool
}

type hololib struct {
	identity uint64
	basedir  string
}

func (it *hololib) HololibDir() string {
	return filepath.Join(it.basedir, "hololib")
}

func (it *hololib) HolotreeDir() string {
	return filepath.Join(it.basedir, "holotree")
}

func (it *hololib) Location(digest string) string {
	return filepath.Join(it.HololibDir(), "library", digest[:2], digest[2:4], digest[4:6])
}

func (it *hololib) Identity() string {
	return fmt.Sprintf("h%016xt", it.identity)
}

func (it *hololib) Stage() string {
	stage := filepath.Join(it.HolotreeDir(), it.Identity())
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
	err = fs.SaveAs(filepath.Join(it.HololibDir(), "catalog", key))
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
	return filepath.Join(it.HololibDir(), "catalog", key)
}

func (it *hololib) HasBlueprint(blueprint []byte) bool {
	key := textual(sipit(blueprint), 0)
	return pathlib.IsFile(it.CatalogPath(key))
}

func (it *hololib) Spaces() []*Root {
	basedir := it.HolotreeDir()
	metafiles := pathlib.Glob(basedir, "*.meta")
	roots := make([]*Root, 0, len(metafiles))
	for _, metafile := range metafiles {
		fullpath := filepath.Join(basedir, metafile)
		root, err := NewRoot(fullpath[:len(fullpath)-5])
		if err != nil {
			continue
		}
		err = root.LoadFrom(fullpath)
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
	metafile := filepath.Join(it.HolotreeDir(), fmt.Sprintf("%s.meta", name))
	targetdir := filepath.Join(it.HolotreeDir(), name)
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
	err = fs.LoadFrom(filepath.Join(it.HololibDir(), "catalog", key))
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
	for _, suffix := range suffixes {
		fullpath := filepath.Join(prefix, suffix)
		err := os.MkdirAll(fullpath, 0o755)
		if err != nil {
			return err
		}
	}
	return nil
}

func New(location string) (Library, error) {
	basedir, err := filepath.Abs(location)
	if err != nil {
		return nil, err
	}
	err = makedirs(basedir, "hololib", "holotree")
	if err != nil {
		return nil, err
	}
	err = makedirs(filepath.Join(basedir, "hololib"), "library", "catalog")
	if err != nil {
		return nil, err
	}
	return &hololib{
		identity: sipit([]byte(basedir)),
		basedir:  basedir,
	}, nil
}
