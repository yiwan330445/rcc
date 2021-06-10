package htfs

import (
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/robocorp/rcc/anywork"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/fail"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/trollhash"
)

func CatalogCheck(library MutableLibrary, fs *Root) Treetop {
	var tool Treetop
	tool = func(path string, it *Dir) error {
		for name, file := range it.Files {
			location := library.ExactLocation(file.Digest)
			if !pathlib.IsFile(location) {
				fullpath := filepath.Join(path, name)
				return fmt.Errorf("Content for %q [%s] is missing!", fullpath, file.Digest)
			}
		}
		for name, subdir := range it.Dirs {
			err := tool(filepath.Join(path, name), subdir)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return tool
}

func DigestMapper(target map[string]string) Treetop {
	var tool Treetop
	tool = func(path string, it *Dir) error {
		for name, subdir := range it.Dirs {
			tool(filepath.Join(path, name), subdir)
		}
		for name, file := range it.Files {
			target[file.Digest] = filepath.Join(path, name)
		}
		return nil
	}
	return tool
}

func DigestRecorder(target map[string]string) Treetop {
	var tool Treetop
	tool = func(path string, it *Dir) error {
		for name, subdir := range it.Dirs {
			tool(filepath.Join(path, name), subdir)
		}
		for name, file := range it.Files {
			target[filepath.Join(path, name)] = file.Digest
		}
		return nil
	}
	return tool
}

func Locator(seek string) Filetask {
	return func(fullpath string, details *File) anywork.Work {
		return func() {
			source, err := os.Open(fullpath)
			if err != nil {
				panic(fmt.Sprintf("Open %q, reason: %v", fullpath, err))
			}
			defer source.Close()
			digest := sha256.New()
			locator := trollhash.LocateWriter(digest, seek)
			_, err = io.Copy(locator, source)
			if err != nil {
				panic(fmt.Sprintf("Copy %q, reason: %v", fullpath, err))
			}
			details.Rewrite = locator.Locations()
			details.Digest = fmt.Sprintf("%02x", digest.Sum(nil))
		}
	}
}

func MakeBranches(path string, it *Dir) error {
	for _, subdir := range it.Dirs {
		err := MakeBranches(filepath.Join(path, subdir.Name), subdir)
		if err != nil {
			return err
		}
	}
	if len(it.Dirs) == 0 {
		err := os.MkdirAll(path, 0o750)
		if err != nil {
			return err
		}
	}
	return os.Chtimes(path, motherTime, motherTime)
}

func ScheduleLifters(library MutableLibrary, stats *stats) Treetop {
	var scheduler Treetop
	scheduler = func(path string, it *Dir) error {
		for name, subdir := range it.Dirs {
			scheduler(filepath.Join(path, name), subdir)
		}
		for name, file := range it.Files {
			directory := library.Location(file.Digest)
			if !pathlib.IsDir(directory) {
				os.MkdirAll(directory, 0o755)
			}
			sinkpath := filepath.Join(directory, file.Digest)
			ok := pathlib.IsFile(sinkpath)
			stats.Dirty(!ok)
			if ok {
				continue
			}
			sourcepath := filepath.Join(path, name)
			anywork.Backlog(LiftFile(sourcepath, sinkpath))
		}
		return nil
	}
	return scheduler
}

func LiftFile(sourcename, sinkname string) anywork.Work {
	return func() {
		source, err := os.Open(sourcename)
		if err != nil {
			panic(err)
		}
		defer source.Close()
		sink, err := os.Create(sinkname)
		if err != nil {
			panic(err)
		}
		defer sink.Close()
		writer, err := gzip.NewWriterLevel(sink, gzip.BestSpeed)
		if err != nil {
			panic(err)
		}
		_, err = io.Copy(writer, source)
		if err != nil {
			panic(err)
		}
		err = writer.Close()
		if err != nil {
			panic(err)
		}
		sink.Sync()
	}
}

func LiftFlatFile(sourcename, sinkname string) anywork.Work {
	return func() {
		source, err := os.Open(sourcename)
		if err != nil {
			panic(err)
		}
		defer source.Close()
		sink, err := os.Create(sinkname)
		if err != nil {
			panic(err)
		}
		defer sink.Close()
		_, err = io.Copy(sink, source)
		if err != nil {
			panic(err)
		}
		sink.Sync()
	}
}

func DropFile(library Library, digest, sinkname string, details *File, rewrite []byte) anywork.Work {
	return func() {
		reader, closer, err := library.Open(digest)
		if err != nil {
			panic(err)
		}
		defer closer()
		sink, err := os.Create(sinkname)
		if err != nil {
			panic(err)
		}
		defer sink.Close()
		_, err = io.Copy(sink, reader)
		if err != nil {
			panic(err)
		}
		sink.Sync()
		for _, position := range details.Rewrite {
			_, err = sink.Seek(position, 0)
			if err != nil {
				panic(fmt.Sprintf("%v %d", err, position))
			}
			_, err = sink.Write(rewrite)
			if err != nil {
				panic(err)
			}
		}
		sink.Sync()
		os.Chmod(sinkname, details.Mode)
		os.Chtimes(sinkname, motherTime, motherTime)
	}
}

func DropFlatFile(sourcename, sinkname string, details *File, rewrite []byte) anywork.Work {
	return func() {
		source, err := os.Open(sourcename)
		if err != nil {
			panic(err)
		}
		defer source.Close()
		sink, err := os.Create(sinkname)
		if err != nil {
			panic(err)
		}
		defer sink.Close()
		_, err = io.Copy(sink, source)
		if err != nil {
			panic(err)
		}
		sink.Sync()
		for _, position := range details.Rewrite {
			_, err = sink.Seek(position, 0)
			if err != nil {
				panic(fmt.Sprintf("%v %d", err, position))
			}
			_, err = sink.Write(rewrite)
			if err != nil {
				panic(err)
			}
		}
		sink.Sync()
		os.Chmod(sinkname, details.Mode)
		os.Chtimes(sinkname, motherTime, motherTime)
	}
}

func RemoveFile(filename string) anywork.Work {
	return func() {
		err := os.Remove(filename)
		if err != nil {
			panic(err)
		}
	}
}

func RemoveDirectory(dirname string) anywork.Work {
	return func() {
		err := os.RemoveAll(dirname)
		if err != nil {
			panic(err)
		}
	}
}

func RestoreDirectory(library Library, fs *Root, current map[string]string, stats *stats) Dirtask {
	return func(path string, it *Dir) anywork.Work {
		return func() {
			content, err := os.ReadDir(path)
			if err != nil {
				panic(err)
			}
			files := make(map[string]bool)
			for _, part := range content {
				directpath := filepath.Join(path, part.Name())
				info, err := os.Stat(directpath)
				if err != nil {
					panic(err)
				}
				if info.IsDir() {
					_, ok := it.Dirs[part.Name()]
					if !ok {
						common.Trace("* Holotree: remove extra directory %q", directpath)
						anywork.Backlog(RemoveDirectory(directpath))
					}
					stats.Dirty(!ok)
					continue
				}
				files[part.Name()] = true
				found, ok := it.Files[part.Name()]
				if !ok {
					common.Trace("* Holotree: remove extra file      %q", directpath)
					anywork.Backlog(RemoveFile(directpath))
					stats.Dirty(true)
					continue
				}
				shadow, ok := current[directpath]
				golden := !ok || found.Digest == shadow
				ok = golden && found.Match(info)
				stats.Dirty(!ok)
				if !ok {
					common.Trace("* Holotree: update changed file    %q", directpath)
					anywork.Backlog(DropFile(library, found.Digest, directpath, found, fs.Rewrite()))
				}
			}
			for name, found := range it.Files {
				directpath := filepath.Join(path, name)
				_, seen := files[name]
				if !seen {
					stats.Dirty(true)
					common.Trace("* Holotree: add missing file       %q", directpath)
					anywork.Backlog(DropFile(library, found.Digest, directpath, found, fs.Rewrite()))
				}
			}
		}
	}
}

type Zipper interface {
	Add(fullpath, relativepath string) error
}

func ZipRoot(library MutableLibrary, fs *Root, sink Zipper) Treetop {
	var tool Treetop
	baseline := common.HololibLocation()
	tool = func(path string, it *Dir) (err error) {
		defer fail.Around(&err)

		for _, file := range it.Files {
			location := library.ExactLocation(file.Digest)
			relative, err := filepath.Rel(baseline, location)
			fail.On(err != nil, "Relative path error: %s -> %s -> %v", baseline, location, err)
			err = sink.Add(location, relative)
			fail.On(err != nil, "%v", err)
		}
		for name, subdir := range it.Dirs {
			err := tool(filepath.Join(path, name), subdir)
			fail.On(err != nil, "%v", err)
		}
		return nil
	}
	return tool
}
