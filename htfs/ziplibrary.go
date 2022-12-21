package htfs

import (
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/fail"
	"github.com/robocorp/rcc/journal"
	"github.com/robocorp/rcc/pathlib"
)

type ziplibrary struct {
	content  *zip.ReadCloser
	identity uint64
	root     *Root
	lookup   map[string]*zip.File
}

func ZipLibrary(zipfile string) (Library, error) {
	content, err := zip.OpenReader(zipfile)
	if err != nil {
		return nil, err
	}
	lookup := make(map[string]*zip.File)
	for _, entry := range content.File {
		lookup[entry.Name] = entry
	}
	identity := strings.ToLower(fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH))
	return &ziplibrary{
		content:  content,
		identity: sipit([]byte(identity)),
		lookup:   lookup,
	}, nil
}

func (it *ziplibrary) ValidateBlueprint(blueprint []byte) error {
	return nil
}

func (it *ziplibrary) HasBlueprint(blueprint []byte) bool {
	key := BlueprintHash(blueprint)
	_, ok := it.lookup[it.CatalogPath(key)]
	return ok
}

func (it *ziplibrary) openFile(filename string) (readable io.Reader, closer Closer, err error) {
	content, ok := it.lookup[filename]
	if !ok {
		return nil, nil, fmt.Errorf("Missing file: %q", filename)
	}
	file, err := content.Open()
	if err != nil {
		return nil, nil, err
	}
	wrapper, err := gzip.NewReader(file)
	if err != nil {
		return nil, nil, err
	}
	closer = func() error {
		wrapper.Close()
		return file.Close()
	}
	return wrapper, closer, nil
}

func (it *ziplibrary) Open(digest string) (readable io.Reader, closer Closer, err error) {
	filename := filepath.Join("library", digest[:2], digest[2:4], digest[4:6], digest)
	return it.openFile(filename)
}

func (it *ziplibrary) CatalogPath(key string) string {
	return filepath.Join("catalog", CatalogName(key))
}

func (it *ziplibrary) TargetDir(blueprint, client, tag []byte) (path string, err error) {
	defer fail.Around(&err)
	key := BlueprintHash(blueprint)
	name := ControllerSpaceName(client, tag)
	fs, err := NewRoot(".")
	fail.On(err != nil, "Failed to create root -> %v", err)
	catalog := it.CatalogPath(key)
	reader, closer, err := it.openFile(catalog)
	fail.On(err != nil, "Failed to open catalog %q -> %v", catalog, err)
	defer closer()
	err = fs.ReadFrom(reader)
	fail.On(err != nil, "Failed to read catalog %q -> %v", catalog, err)
	return filepath.Join(fs.HolotreeBase(), name), nil
}

func (it *ziplibrary) Restore(blueprint, client, tag []byte) (result string, err error) {
	defer fail.Around(&err)
	defer common.Stopwatch("Holotree restore took:").Debug()
	key := BlueprintHash(blueprint)
	common.Timeline("holotree restore start %s (zip)", key)
	name := ControllerSpaceName(client, tag)
	fs, err := NewRoot(".")
	fail.On(err != nil, "Failed to create root -> %v", err)
	catalog := it.CatalogPath(key)
	reader, closer, err := it.openFile(catalog)
	fail.On(err != nil, "Failed to open catalog %q -> %v", catalog, err)
	defer closer()
	err = fs.ReadFrom(reader)
	fail.On(err != nil, "Failed to read catalog %q -> %v", catalog, err)
	metafile := filepath.Join(fs.HolotreeBase(), fmt.Sprintf("%s.meta", name))
	targetdir := filepath.Join(fs.HolotreeBase(), name)
	lockfile := filepath.Join(fs.HolotreeBase(), fmt.Sprintf("%s.lck", name))
	completed := pathlib.LockWaitMessage(lockfile, "Serialized holotree restore [holotree base lock]")
	locker, err := pathlib.Locker(lockfile, 30000)
	completed()
	fail.On(err != nil, "Could not get lock for %s. Quiting.", targetdir)
	defer locker.Release()
	journal.Post("space-used", metafile, "zipped holotree with blueprint %s from %s", key, catalog)
	currentstate := make(map[string]string)
	shadow, err := NewRoot(targetdir)
	if err == nil {
		err = shadow.LoadFrom(metafile)
	}
	if err == nil {
		common.TimelineBegin("holotree digest start (zip)")
		shadow.Treetop(DigestRecorder(currentstate))
		common.TimelineEnd()
	}
	err = fs.Relocate(targetdir)
	fail.On(err != nil, "Failed to relocate %q -> %v", targetdir, err)
	common.TimelineBegin("holotree make branches start (zip)")
	err = fs.Treetop(MakeBranches)
	common.TimelineEnd()
	fail.On(err != nil, "Failed to make branches %q -> %v", targetdir, err)
	score := &stats{}
	common.TimelineBegin("holotree restore start (zip)")
	err = fs.AllDirs(RestoreDirectory(it, fs, currentstate, score))
	fail.On(err != nil, "Failed to restore directory %q -> %v", targetdir, err)
	common.TimelineEnd()
	defer common.Timeline("- dirty %d/%d", score.dirty, score.total)
	common.Debug("Holotree dirty workload: %d/%d\n", score.dirty, score.total)
	journal.CurrentBuildEvent().Dirty(score.Dirtyness())
	fs.Controller = string(client)
	fs.Space = string(tag)
	err = fs.SaveAs(metafile)
	fail.On(err != nil, "Failed to save metafile %q -> %v", metafile, err)
	return targetdir, nil
}
