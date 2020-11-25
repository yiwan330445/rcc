package shell

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type Task struct {
	environment []string
	directory   string
	executable  string
	args        []string
}

func New(environment []string, directory string, task ...string) *Task {
	executable, args := task[0], task[1:]
	return &Task{
		environment: environment,
		directory:   directory,
		executable:  executable,
		args:        args,
	}
}

func (it *Task) execute(stdin io.Reader, stdout, stderr io.Writer) (int, error) {
	command := exec.Command(it.executable, it.args...)
	command.Env = it.environment
	command.Dir = it.directory
	command.Stdin = stdin
	command.Stdout = stdout
	command.Stderr = stderr
	err := command.Run()
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
	return it.execute(os.Stdin, os.Stdout, os.Stderr)
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
	stdout := io.MultiWriter(os.Stdout, outfile)
	stderr := io.MultiWriter(os.Stderr, errfile)
	if !interactive {
		os.Stdin.Close()
	}
	return it.execute(os.Stdin, stdout, stderr)
}

func (it *Task) Observed(sink io.Writer, interactive bool) (int, error) {
	stdout := io.MultiWriter(os.Stdout, sink)
	stderr := io.MultiWriter(os.Stderr, sink)
	if !interactive {
		os.Stdin.Close()
	}
	return it.execute(os.Stdin, stdout, stderr)
}
