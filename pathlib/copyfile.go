package pathlib

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	"github.com/robocorp/rcc/common"
)

type copyfunc func(io.Writer, io.Reader) (int64, error)

func archiver(target io.Writer, source io.Reader) (int64, error) {
	wrapper, err := gzip.NewWriterLevel(target, flate.BestSpeed)
	if err != nil {
		return 0, err
	}
	defer wrapper.Close()
	return io.Copy(wrapper, source)
}

func restorer(target io.Writer, source io.Reader) (int64, error) {
	wrapper, err := gzip.NewReader(source)
	if err != nil {
		return 0, err
	}
	defer wrapper.Close()
	return io.Copy(target, wrapper)
}

type Copier func(string, string, bool) error

func ArchiveFile(source, target string, overwrite bool) error {
	return copyFile(source, target, overwrite, archiver)
}

func RestoreFile(source, target string, overwrite bool) error {
	return copyFile(source, target, overwrite, restorer)
}

func CopyFile(source, target string, overwrite bool) error {
	mark, err := Modtime(source)
	if err != nil {
		return err
	}
	err = copyFile(source, target, overwrite, io.Copy)
	TouchWhen(target, mark)
	return err
}

func copyFile(source, target string, overwrite bool, copier copyfunc) error {
	targetDir := filepath.Dir(target)
	err := os.MkdirAll(targetDir, 0o755)
	if err != nil {
		return err
	}
	if overwrite && Exists(target) {
		err = os.Remove(target)
	}
	if err != nil {
		return err
	}
	readable, err := os.Open(source)
	if err != nil {
		return err
	}
	defer readable.Close()
	stats, err := readable.Stat()
	if err != nil {
		return err
	}
	writable, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_EXCL, stats.Mode())
	if err != nil {
		return err
	}
	defer writable.Close()

	_, err = copier(writable, readable)
	if err != nil {
		common.Error("copy-file", err)
	}

	return err
}
