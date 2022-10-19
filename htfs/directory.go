package htfs

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/robocorp/rcc/anywork"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/fail"
	"github.com/robocorp/rcc/pathlib"
)

var (
	killfile map[string]bool
)

func init() {
	killfile = make(map[string]bool)
	killfile["__MACOSX"] = true
	killfile["__pycache__"] = true
	killfile[".pyc"] = true
	killfile[".git"] = true
	killfile[".hg"] = true
	killfile[".svn"] = true
	killfile[".gitignore"] = true

	pathlib.MakeSharedDir(common.HoloLocation())
	pathlib.MakeSharedDir(common.HololibCatalogLocation())
	pathlib.MakeSharedDir(common.HololibLibraryLocation())
	pathlib.MakeSharedDir(common.HololibUsageLocation())
	pathlib.MakeSharedDir(common.HololibPids())
}

type Filetask func(string, *File) anywork.Work
type Dirtask func(string, *Dir) anywork.Work
type Treetop func(string, *Dir) error

type Root struct {
	RccVersion string `json:"rcc"`
	Identity   string `json:"identity"`
	Path       string `json:"path"`
	Controller string `json:"controller"`
	Space      string `json:"space"`
	Platform   string `json:"platform"`
	Blueprint  string `json:"blueprint"`
	Lifted     bool   `json:"lifted"`
	Tree       *Dir   `json:"tree"`
	source     string
}

func NewRoot(path string) (*Root, error) {
	fullpath, err := pathlib.Abs(path)
	if err != nil {
		return nil, err
	}
	basename := filepath.Base(fullpath)
	return &Root{
		Identity:   basename,
		Path:       fullpath,
		Platform:   common.Platform(),
		Lifted:     false,
		Tree:       newDir("", "", false),
		RccVersion: common.Version,
		source:     fullpath,
	}, nil
}

func (it *Root) Show(filename string) ([]byte, error) {
	return it.Tree.Show(filepath.SplitList(filename), filename)
}

func (it *Root) Source() string {
	return it.source
}

func (it *Root) Touch() {
	touchUsedHash(it.Blueprint)
}

func (it *Root) HolotreeBase() string {
	return filepath.Dir(it.Path)
}

func (it *Root) Signature() uint64 {
	return sipit([]byte(strings.ToLower(fmt.Sprintf("%s %q", it.Platform, it.Path))))
}

func (it *Root) Rewrite() []byte {
	return []byte(it.Identity)
}

func (it *Root) Relocate(target string) error {
	locate := filepath.Dir(target)
	if it.HolotreeBase() != locate {
		return fmt.Errorf("Base directory mismatch: %q vs %q.", it.HolotreeBase(), locate)
	}
	basename := filepath.Base(target)
	if len(it.Identity) != len(basename) {
		return fmt.Errorf("Base name length mismatch: %q vs %q.", it.Identity, basename)
	}
	if len(it.Path) != len(target) {
		return fmt.Errorf("Path length mismatch: %q vs %q.", it.Path, target)
	}
	it.Path = target
	it.Identity = basename
	return nil
}

func (it *Root) Lift() error {
	if it.Lifted {
		return nil
	}
	it.Lifted = true
	return it.Tree.Lift(it.Path)
}

func (it *Root) Treetop(task Treetop) error {
	common.TimelineBegin("holotree treetop sync start")
	defer common.TimelineEnd()
	err := task(it.Path, it.Tree)
	if err != nil {
		return err
	}
	return anywork.Sync()
}

func (it *Root) Stats() (*TreeStats, error) {
	task, stats := CalculateTreeStats()
	err := it.AllDirs(task)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (it *Root) AllDirs(task Dirtask) error {
	common.TimelineBegin("holotree dirs sync start")
	defer common.TimelineEnd()
	it.Tree.AllDirs(it.Path, task)
	return anywork.Sync()
}

func (it *Root) AllFiles(task Filetask) error {
	common.TimelineBegin("holotree files sync start")
	defer common.TimelineEnd()
	it.Tree.AllFiles(it.Path, task)
	return anywork.Sync()
}

func (it *Root) AsJson() ([]byte, error) {
	return json.MarshalIndent(it, "", "  ")
}

func (it *Root) SaveAs(filename string) error {
	content, err := it.AsJson()
	if err != nil {
		return err
	}
	sink, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer sink.Close()
	defer sink.Sync()
	writer, err := gzip.NewWriterLevel(sink, gzip.BestSpeed)
	if err != nil {
		return err
	}
	defer writer.Close()
	_, err = writer.Write(content)
	if err != nil {
		return err
	}
	return nil
}

func (it *Root) ReadFrom(source io.Reader) error {
	decoder := json.NewDecoder(source)
	return decoder.Decode(&it)
}

func (it *Root) LoadFrom(filename string) error {
	common.TimelineBegin("holotree load %q", filename)
	defer common.TimelineEnd()
	source, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer source.Close()
	reader, err := gzip.NewReader(source)
	if err != nil {
		return err
	}
	it.source = filename
	defer reader.Close()
	return it.ReadFrom(reader)
}

type Dir struct {
	Name    string           `json:"name"`
	Symlink string           `json:"symlink,omitempty"`
	Mode    fs.FileMode      `json:"mode"`
	Dirs    map[string]*Dir  `json:"subdirs"`
	Files   map[string]*File `json:"files"`
	Shadow  bool             `json:"shadow,omitempty"`
}

func showFile(filename string) (content []byte, err error) {
	defer fail.Around(&err)

	reader, closer, err := gzDelegateOpen(filename, true)
	fail.On(err != nil, "Failed to open %q, reason: %v", filename, err)
	defer closer()

	sink := bytes.NewBuffer(nil)
	_, err = io.Copy(sink, reader)
	fail.On(err != nil, "Failed to read %q, reason: %v", filename, err)
	return sink.Bytes(), nil
}

func (it *Dir) Show(path []string, fullpath string) ([]byte, error) {
	if len(path) > 1 {
		subtree, ok := it.Dirs[path[0]]
		if !ok {
			return nil, fmt.Errorf("Not found: %s", fullpath)
		}
		return subtree.Show(path[1:], fullpath)
	}
	file, ok := it.Files[path[0]]
	if !ok {
		return nil, fmt.Errorf("Not found: %s", fullpath)
	}
	location := guessLocation(file.Digest)
	rawfile := filepath.Join(common.HololibLibraryLocation(), location)
	return showFile(rawfile)
}

func (it *Dir) IsSymlink() bool {
	return len(it.Symlink) > 0
}

func (it *Dir) AllDirs(path string, task Dirtask) {
	for name, dir := range it.Dirs {
		fullpath := filepath.Join(path, name)
		dir.AllDirs(fullpath, task)
	}
	anywork.Backlog(task(path, it))
}

func (it *Dir) AllFiles(path string, task Filetask) {
	for name, dir := range it.Dirs {
		fullpath := filepath.Join(path, name)
		dir.AllFiles(fullpath, task)
	}
	for name, file := range it.Files {
		fullpath := filepath.Join(path, name)
		anywork.Backlog(task(fullpath, file))
	}
}

func (it *Dir) Lift(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	it.Mode = stat.Mode()
	content, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	shadow := it.Shadow || it.IsSymlink()
	for _, part := range content {
		if killfile[part.Name()] || killfile[filepath.Ext(part.Name())] {
			continue
		}
		fullpath := filepath.Join(path, part.Name())
		// following must be done to get by symbolic links
		info, err := os.Stat(fullpath)
		if err != nil {
			return err
		}
		symlink, _ := pathlib.Symlink(fullpath)
		if info.IsDir() {
			it.Dirs[part.Name()] = newDir(info.Name(), symlink, shadow)
			continue
		}
		it.Files[part.Name()] = newFile(info, symlink)
	}
	for name, dir := range it.Dirs {
		err = dir.Lift(filepath.Join(path, name))
		if err != nil {
			return err
		}
	}
	return nil
}

type File struct {
	Name    string      `json:"name"`
	Symlink string      `json:"symlink,omitempty"`
	Size    int64       `json:"size"`
	Mode    fs.FileMode `json:"mode"`
	Digest  string      `json:"digest"`
	Rewrite []int64     `json:"rewrite"`
}

func (it *File) IsSymlink() bool {
	return len(it.Symlink) > 0
}

func (it *File) Match(info fs.FileInfo) bool {
	name := it.Name == info.Name()
	size := it.Size == info.Size()
	mode := it.Mode == info.Mode()
	return name && size && mode
}

func newDir(name, symlink string, shadow bool) *Dir {
	return &Dir{
		Name:    name,
		Symlink: symlink,
		Dirs:    make(map[string]*Dir),
		Files:   make(map[string]*File),
		Shadow:  shadow,
	}
}

func newFile(info fs.FileInfo, symlink string) *File {
	return &File{
		Name:    info.Name(),
		Symlink: symlink,
		Mode:    info.Mode(),
		Size:    info.Size(),
		Digest:  "N/A",
		Rewrite: make([]int64, 0),
	}
}
