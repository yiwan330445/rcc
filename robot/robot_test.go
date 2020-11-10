package robot_test

import (
	"strings"
	"testing"

	"github.com/robocorp/rcc/hamlet"
	"github.com/robocorp/rcc/robot"
)

func TestCannotReadMissingRobotYaml(t *testing.T) {
	must, wont := hamlet.Specifications(t)

	sut, err := robot.LoadRobotYaml("testdata/badmissing.yaml")
	wont.Nil(err)
	must.Nil(sut)
}

func TestCanReadRealRobotYaml(t *testing.T) {
	must, wont := hamlet.Specifications(t)

	sut, err := robot.LoadRobotYaml("testdata/robot.yaml")
	must.Nil(err)
	wont.Nil(sut)
	must.Equal(1, len(sut.IgnoreFiles()))
	must.Equal(3, len(sut.AvailableTasks()))
	must.Nil(sut.DefaultTask())
	must.Nil(sut.TaskByName(""))
	must.Nil(sut.TaskByName("missing task"))
	wont.Nil(sut.TaskByName("task form name"))
	wont.Nil(sut.TaskByName("Shell Form Name"))
	wont.Nil(sut.TaskByName("  Old command form name "))
	must.Equal(1, len(sut.Paths("")))
	must.Equal(3, len(sut.PythonPaths("")))
	must.True(2 < len(sut.SearchPath("", ".")))
	must.True(strings.HasSuffix(sut.CondaConfigFile(), "conda.yaml"))
	must.True(strings.HasSuffix(sut.WorkingDirectory(""), "testdata"))
	must.True(strings.HasSuffix(sut.ArtifactDirectory(""), "output"))
	valid, err := sut.Validate()
	must.True(valid)
	must.Nil(err)
}

func TestCanGetShellFormCommand(t *testing.T) {
	must, wont := hamlet.Specifications(t)

	sut, err := robot.LoadRobotYaml("testdata/robot.yaml")
	must.Nil(err)
	wont.Nil(sut)
	task := sut.TaskByName("Shell Form Name")
	wont.Nil(task)

	command := task.Commandline()
	wont.Nil(command)
	must.Equal(8, len(command))
	must.Equal("tasks/shilling.robot", command[7])
}

func TestCanGetTaskFormCommand(t *testing.T) {
	must, wont := hamlet.Specifications(t)

	sut, err := robot.LoadRobotYaml("testdata/robot.yaml")
	must.Nil(err)
	wont.Nil(sut)
	task := sut.TaskByName("Task Form Name")
	wont.Nil(task)

	command := task.Commandline()
	wont.Nil(command)
	must.Equal(12, len(command))
}
