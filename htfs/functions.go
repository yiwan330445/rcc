package htfs

import (
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/robocorp/rcc/anywork"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/trollhash"
)

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

func ScheduleLifters(library Library, stats *stats) Treetop {
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

func DropFile(sourcename, sinkname string, details *File, rewrite []byte) anywork.Work {
	return func() {
		source, err := os.Open(sourcename)
		if err != nil {
			panic(err)
		}
		defer source.Close()
		reader, err := gzip.NewReader(source)
		if err != nil {
			panic(err)
		}
		defer reader.Close()
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
					stats.Dirty(!ok)
					if !ok {
						anywork.Backlog(RemoveDirectory(directpath))
					}
					continue
				}
				files[part.Name()] = true
				found, ok := it.Files[part.Name()]
				if !ok {
					stats.Dirty(true)
					anywork.Backlog(RemoveFile(directpath))
					continue
				}
				shadow, ok := current[directpath]
				golden := !ok || found.Digest == shadow
				ok = golden && found.Match(info)
				stats.Dirty(!ok)
				if !ok {
					directory := library.Location(found.Digest)
					droppath := filepath.Join(directory, found.Digest)
					anywork.Backlog(DropFile(droppath, directpath, found, fs.Rewrite()))
				}
			}
			for name, found := range it.Files {
				directpath := filepath.Join(path, name)
				_, seen := files[name]
				if !seen {
					stats.Dirty(true)
					directory := library.Location(found.Digest)
					droppath := filepath.Join(directory, found.Digest)
					anywork.Backlog(DropFile(droppath, directpath, found, fs.Rewrite()))
				}
			}
		}
	}
}
