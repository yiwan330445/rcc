package operations

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/robot"
)

type WriteTarget struct {
	Source *zip.File
	Target string
}

type Command interface {
	Execute() bool
}

type CommandChannel chan Command
type CompletedChannel chan bool

func (it *WriteTarget) Execute() bool {
	source, err := it.Source.Open()
	if err != nil {
		return false
	}
	defer source.Close()
	err = os.MkdirAll(filepath.Dir(it.Target), 0o750)
	if err != nil {
		return false
	}
	target, err := os.Create(it.Target)
	if err != nil {
		return false
	}
	defer target.Close()
	common.Debug("- %v", it.Target)
	_, err = io.Copy(target, source)
	if err != nil {
		common.Debug("  - failure: %v", err)
	}
	os.Chtimes(it.Target, it.Source.Modified, it.Source.Modified)
	return err == nil
}

type unzipper struct {
	reader *zip.Reader
	closer io.Closer
}

func (it *unzipper) Close() {
	it.closer.Close()
}

func newPayloadUnzipper(filename string) (*unzipper, error) {
	payloader, err := PayloadReaderAt(filename)
	if err != nil {
		return nil, err
	}
	reader, err := zip.NewReader(payloader, payloader.Limit())
	if err != nil {
		return nil, err
	}
	return &unzipper{
		reader: reader,
		closer: payloader,
	}, nil
}

func newUnzipper(filename string) (*unzipper, error) {
	reader, err := zip.OpenReader(filename)
	if err != nil {
		return nil, err
	}
	return &unzipper{
		reader: &reader.Reader,
		closer: reader,
	}, nil
}

func loopExecutor(work CommandChannel, done CompletedChannel) {
	// This is PoC code, for parallel extraction
	for {
		task, ok := <-work
		if !ok {
			break
		}
		task.Execute()
	}
	done <- true
}

func (it *unzipper) Explode(workers int, directory string) error {
	// This is PoC code, for parallel extraction
	common.Debug("Exploding:")

	todo := make(CommandChannel)
	done := make(CompletedChannel)

	for step := 0; step < workers; step++ {
		go loopExecutor(todo, done)
	}

	for _, entry := range it.reader.File {
		if entry.FileInfo().IsDir() {
			continue
		}
		todo <- &WriteTarget{
			Source: entry,
			Target: filepath.Join(directory, entry.Name),
		}
	}

	close(todo)

	for step := 0; step < workers; step++ {
		<-done
	}

	common.Debug("Done.")

	return nil
}

func (it *unzipper) Extract(directory string) error {
	common.Debug("Extracting:")
	success := true
	for _, entry := range it.reader.File {
		if entry.FileInfo().IsDir() {
			continue
		}
		target := filepath.Join(directory, entry.Name)
		todo := WriteTarget{
			Source: entry,
			Target: target,
		}
		success = todo.Execute() && success
	}
	common.Debug("Done.")
	if !success {
		return fmt.Errorf("Problems while unwrapping robot. Use --debug to see details.")
	}
	return nil
}

type zipper struct {
	handle   *os.File
	writer   *zip.Writer
	failures []error
}

func newZipper(filename string) (*zipper, error) {
	handle, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	writer := zip.NewWriter(handle)
	return &zipper{
		handle:   handle,
		writer:   writer,
		failures: make([]error, 2),
	}, nil
}

func (it *zipper) Note(err error) {
	it.failures = append(it.failures, err)
	common.Debug("Warning! %v", err)
}

func (it *zipper) Add(fullpath, relativepath string, details os.FileInfo) {
	if details != nil {
		common.Debug("- %v size %v", relativepath, details.Size())
	} else {
		common.Debug("- %v", relativepath)
	}
	source, err := os.Open(fullpath)
	if err != nil {
		it.Note(err)
		return
	}
	defer source.Close()
	target, err := it.writer.Create(relativepath)
	if err != nil {
		it.Note(err)
		return
	}
	_, err = io.Copy(target, source)
	if err != nil {
		it.Note(err)
	}
}

func (it *zipper) AddBlob(relativepath string, blob []byte) {
	target, err := it.writer.Create(relativepath)
	if err != nil {
		it.Note(err)
		return
	}
	_, err = target.Write(blob)
	if err != nil {
		it.Note(err)
	}
}

func (it *zipper) Close() {
	err := it.writer.Close()
	if err != nil {
		common.Log("Problem closing zip writer: %v", err)
	}
	err = it.handle.Close()
	if err != nil {
		common.Log("Problem closing zipfile: %v", err)
	}
}

func defaultIgnores(selfie string) pathlib.Ignore {
	result := make([]pathlib.Ignore, 0, 10)
	result = append(result, pathlib.IgnorePattern(selfie))
	result = append(result, pathlib.IgnorePattern(".git"))
	result = append(result, pathlib.IgnorePattern(".rpa"))
	result = append(result, pathlib.IgnorePattern("rcc"))
	result = append(result, pathlib.IgnorePattern("output/"))
	result = append(result, pathlib.IgnorePattern("temp/"))
	result = append(result, pathlib.IgnorePattern("tmp/"))
	result = append(result, pathlib.IgnorePattern("__pycache__"))
	result = append(result, pathlib.IgnorePattern("__MACOSX"))
	return pathlib.CompositeIgnore(result...)
}

func CarrierUnzip(directory, carrier string, force, temporary bool) error {
	fullpath, err := filepath.Abs(directory)
	if err != nil {
		return err
	}
	if force {
		err = pathlib.EnsureDirectoryExists(fullpath)
	} else {
		err = pathlib.EnsureEmptyDirectory(fullpath)
	}
	if err != nil {
		return err
	}
	unzip, err := newPayloadUnzipper(carrier)
	if err != nil {
		return err
	}
	defer unzip.Close()
	err = unzip.Extract(fullpath)
	if err != nil {
		return err
	}
	if temporary {
		return nil
	}
	err = UpdateRobot(fullpath)
	if err != nil {
		return err
	}
	return FixDirectory(fullpath)
}

func Unzip(directory, zipfile string, force, temporary bool) error {
	common.Timeline("unzip %q to %q", zipfile, directory)
	defer common.Timeline("unzip done")
	fullpath, err := filepath.Abs(directory)
	if err != nil {
		return err
	}
	if force {
		err = pathlib.EnsureDirectoryExists(fullpath)
	} else {
		err = pathlib.EnsureEmptyDirectory(fullpath)
	}
	if err != nil {
		return err
	}
	unzip, err := newUnzipper(zipfile)
	if err != nil {
		return err
	}
	defer unzip.Close()
	err = unzip.Extract(fullpath)
	if err != nil {
		return err
	}
	if temporary {
		return nil
	}
	err = UpdateRobot(fullpath)
	if err != nil {
		return err
	}
	return FixDirectory(fullpath)
}

func Zip(directory, zipfile string, ignores []string) error {
	common.Timeline("zip %q to %q", directory, zipfile)
	defer common.Timeline("zip done")
	common.Debug("Wrapping %v into %v ...", directory, zipfile)
	config, err := robot.LoadRobotYaml(robot.DetectConfigurationName(directory), false)
	if err != nil {
		return err
	}
	ignores = append(ignores, config.IgnoreFiles()...)
	zipper, err := newZipper(zipfile)
	if err != nil {
		return err
	}
	defer zipper.Close()
	ignored, err := pathlib.LoadIgnoreFiles(ignores)
	if err != nil {
		return err
	}
	defaults := defaultIgnores(zipfile)
	pathlib.ForceWalk(directory, pathlib.ForceFilename("hololib.zip"), pathlib.CompositeIgnore(defaults, ignored), zipper.Add)
	return nil
}
