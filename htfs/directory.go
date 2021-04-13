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

	"github.com/robocorp/rcc/anywork"
	"github.com/robocorp/rcc/pathlib"
)

var (
	killfile map[string]bool
)

func init() {
	killfile = make(map[string]bool)
	killfile["__pycache__"] = true
	killfile[".pyc"] = true
	killfile[".git"] = true
	killfile[".hg"] = true
	killfile[".svn"] = true
	killfile[".gitignore"] = true
}

type Filetask func(string, *File) anywork.Work
type Dirtask func(string, *Dir) anywork.Work
type Treetop func(string, *Dir) error

type Root struct {
	Identity   string `json:"identity"`
	Path       string `json:"path"`
	Controller string `json:"controller"`
	Space      string `json:"space"`
	Blueprint  string `json:"blueprint"`
	Lifted     bool   `json:"lifted"`
	Tree       *Dir   `json:"tree"`
}

func NewRoot(path string) (*Root, error) {
	fullpath, err := pathlib.Abs(path)
	if err != nil {
		return nil, err
	}
	basename := filepath.Base(fullpath)
	return &Root{
		Identity: basename,
		Path:     fullpath,
		Lifted:   false,
		Tree:     newDir(""),
	}, nil
}

func (it *Root) Rewrite() []byte {
	return []byte(it.Identity)
}

func (it *Root) Relocate(target string) error {
	origin := filepath.Dir(it.Path)
	locate := filepath.Dir(target)
	if origin != locate {
		return fmt.Errorf("Base directory mismatch: %q vs %q.", origin, locate)
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
	err := task(it.Path, it.Tree)
	anywork.Sync()
	return err
}

func (it *Root) AllDirs(task Dirtask) {
	it.Tree.AllDirs(it.Path, task)
	anywork.Sync()
}

func (it *Root) AllFiles(task Filetask) {
	it.Tree.AllFiles(it.Path, task)
	anywork.Sync()
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

func (it *Root) LoadFrom(filename string) error {
	source, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer source.Close()
	reader, err := gzip.NewReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()
	content := bytes.NewBuffer(nil)
	io.Copy(content, reader)
	return json.Unmarshal(content.Bytes(), &it)
}

type Dir struct {
	Name  string           `json:"name"`
	Mode  fs.FileMode      `json:"mode"`
	Dirs  map[string]*Dir  `json:"subdirs"`
	Files map[string]*File `json:"files"`
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
	source, err := os.Open(path)
	if err != nil {
		return err
	}
	defer source.Close()
	content, err := source.ReadDir(-1)
	if err != nil {
		return err
	}
	for _, part := range content {
		if killfile[part.Name()] || killfile[filepath.Ext(part.Name())] {
			continue
		}
		// following must be done to get by symbolic links
		info, err := os.Stat(filepath.Join(path, part.Name()))
		if err != nil {
			return err
		}
		if info.IsDir() {
			it.Dirs[part.Name()] = newDir(info.Name())
			continue
		}
		it.Files[part.Name()] = newFile(info)
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
	Size    int64       `json:"size"`
	Mode    fs.FileMode `json:"mode"`
	Digest  string      `json:"digest"`
	Rewrite []int64     `json:"rewrite"`
}

func (it *File) Match(info fs.FileInfo) bool {
	name := it.Name == info.Name()
	size := it.Size == info.Size()
	mode := it.Mode == info.Mode()
	return name && size && mode
}

func newDir(name string) *Dir {
	return &Dir{
		Name:  name,
		Dirs:  make(map[string]*Dir),
		Files: make(map[string]*File),
	}
}

func newFile(info fs.FileInfo) *File {
	return &File{
		Name:    info.Name(),
		Mode:    info.Mode(),
		Size:    info.Size(),
		Digest:  "N/A",
		Rewrite: make([]int64, 0),
	}
}
