package shell

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/robocorp/rcc/common"
)

type Task struct {
	environment []string
	directory   string
	executable  string
	args        []string
	stderronly  bool
}

func New(environment []string, directory string, task ...string) *Task {
	executable, args := task[0], task[1:]
	return &Task{
		environment: environment,
		directory:   directory,
		executable:  executable,
		args:        args,
		stderronly:  false,
	}
}

func (it *Task) StderrOnly() *Task {
	it.stderronly = true
	return it
}

func (it *Task) stdout() io.Writer {
	if it.stderronly {
		return os.Stderr
	}
	return os.Stdout
}

func (it *Task) execute(stdin io.Reader, stdout, stderr io.Writer) (int, error) {
	common.Trace("Execute %q with arguments %q", it.executable, it.args)
	command := exec.Command(it.executable, it.args...)
	command.Env = it.environment
	command.Dir = it.directory
	command.Stdin = stdin
	command.Stdout = stdout
	command.Stderr = stderr
	err := command.Start()
	if err != nil {
		return -500, err
	}
	common.Debug("PID #%d is %q.", command.Process.Pid, command)
	defer func() {
		common.Debug("PID #%d finished: %v.", command.Process.Pid, command.ProcessState)
	}()
	err = command.Wait()
	exit, ok := err.(*exec.ExitError)
	if ok {
		return exit.ExitCode(), err
	}
	if err != nil {
		return -500, err
	}
	return 0, nil
}

func (it *Task) Transparent() (int, error) {
	return it.execute(os.Stdin, it.stdout(), os.Stderr)
}

func (it *Task) Tee(folder string, interactive bool) (int, error) {
	err := os.MkdirAll(folder, 0755)
	if err != nil {
		return -600, err
	}
	outfile, err := os.Create(filepath.Join(folder, "stdout.log"))
	if err != nil {
		return -601, err
	}
	defer outfile.Close()
	errfile, err := os.Create(filepath.Join(folder, "stderr.log"))
	if err != nil {
		return -602, err
	}
	defer errfile.Close()
	stdout := io.MultiWriter(it.stdout(), outfile)
	stderr := io.MultiWriter(os.Stderr, errfile)
	var stdin io.Reader = os.Stdin
	if !interactive {
		stdin = bytes.NewReader([]byte{})
	}
	return it.execute(stdin, stdout, stderr)
}

func (it *Task) Observed(sink io.Writer, interactive bool) (int, error) {
	stdout := io.MultiWriter(it.stdout(), sink)
	stderr := io.MultiWriter(os.Stderr, sink)
	var stdin io.Reader = os.Stdin
	if !interactive {
		stdin = bytes.NewReader([]byte{})
	}
	return it.execute(stdin, stdout, stderr)
}

func (it *Task) CaptureOutput() (string, int, error) {
	stdin := bytes.NewReader([]byte{})
	stdout := bytes.NewBuffer(nil)
	code, err := it.execute(stdin, stdout, os.Stderr)
	return stdout.String(), code, err
}
