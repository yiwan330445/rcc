package robot_test

import (
	"strings"
	"testing"

	"github.com/robocorp/rcc/hamlet"
	"github.com/robocorp/rcc/robot"
)

const (
	minimalActivity = `
activities:
  Main activity:
    output: output
    activityRoot: .`
)

func TestCannotReadMissingActivityConfig(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	sut, err := robot.LoadYamlConfiguration("testdata/badmissing.yaml")
	wont_be.Nil(err)
	must_be.Nil(sut)

	sut, err = robot.LoadActivityPackage("testdata/bad.yaml")
	wont_be.Nil(err)
	must_be.Nil(sut)
}

func TestCanReadTemplateActivityConfig(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	raw, err := robot.LoadActivityPackage("testdata/template.yaml")
	must_be.Nil(err)
	wont_be.Nil(raw)
	sut := raw.(*robot.Config)
	must_be.Equal("config/conda.yaml", sut.Conda)
	must_be.True(strings.HasSuffix(sut.CondaConfigFile(), "config/conda.yaml"))
	must_be.Equal(1, len(sut.Activities))
	activity := sut.DefaultTask().(*robot.Activity)
	wont_be.Nil(activity)
	must_be.True(strings.HasSuffix(activity.WorkingDirectory(sut), "/testdata"))
	must_be.Nil(sut.TaskByName("Missing Activity Name"))
	must_be.Same(activity, sut.TaskByName("My activity"))
	must_be.Same(activity, sut.TaskByName("my Activity"))
	must_be.Equal([]string{"My activity"}, sut.AvailableTasks())
	must_be.True(strings.HasSuffix(activity.ArtifactDirectory(sut), "output"))
	must_be.Equal(".", activity.Root)
	wont_be.Nil(activity.Environment)
	wont_be.Nil(activity.Action)
	must_be.Equal(10, len(activity.Commandline()))
	must_be.Equal(1, len(activity.Paths(sut)))
	must_be.True(len(activity.ExecutionEnvironment(sut, "foobar", []string{}, false)) > 3)
	must_be.True(len(activity.ExecutionEnvironment(sut, "foobar", []string{}, true)) > len(activity.ExecutionEnvironment(sut, "foobar", []string{}, false)))
	must_be.Equal(3, len(activity.PythonPaths(sut)))
	ok, err := sut.Validate()
	must_be.True(ok)
	must_be.Nil(err)
}

func TestCanReadComplexActivityConfig(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	raw, err := robot.LoadActivityPackage("testdata/complex.yaml")
	must_be.Nil(err)
	wont_be.Nil(raw)
	sut := raw.(*robot.Config)
	wont_be.Nil(sut.IgnoreFiles())
	must_be.Equal(2, len(sut.IgnoreFiles()))
	must_be.Nil(sut.DefaultTask())
	must_be.Equal("conda.yaml", sut.Conda)
	must_be.Equal(2, len(sut.Activities))
	wont_be.Nil(sut.TaskByName("Read Excel to work item"))
	wont_be.Nil(sut.TaskByName("Generate PDFs from work item"))
	must_be.Equal([]string{"Generate PDFs from work item", "Read Excel to work item"}, sut.AvailableTasks())
	ok, err := sut.Validate()
	must_be.True(ok)
	must_be.Nil(err)
}

func TestCanReadEnvironmentlessActivityConfig(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	raw, err := robot.LoadActivityPackage("testdata/complex.yaml")
	must_be.Nil(err)
	wont_be.Nil(raw)
	sut := raw.(*robot.Config)
	activity := sut.DefaultTask()
	must_be.Nil(activity)
	ok, err := sut.Validate()
	must_be.True(ok)
	must_be.Nil(err)
}

func TestCanReadActivityConfigFromText(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	sut, err := robot.ActivityPackageFrom([]byte(""))
	must_be.Nil(err)
	wont_be.Nil(sut)
	must_be.Nil(sut.DefaultTask())
	must_be.Nil(sut.TaskByName(""))
	must_be.Nil(sut.TaskByName("foo"))
	must_be.Equal("", sut.CondaConfigFile())
	must_be.Equal("", sut.CondaConfigFile())
	ok, err := sut.Validate()
	wont_be.True(ok)
	wont_be.Nil(err)
}

func TestCanParseMinimalActivityConfig(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	sut, err := robot.ActivityPackageFrom([]byte(minimalActivity))
	must_be.Nil(err)
	wont_be.Nil(sut)
	wont_be.Nil(sut.DefaultTask())
	wont_be.Nil(sut.TaskByName(""))
	must_be.Nil(sut.TaskByName("foo"))
	must_be.Equal("", sut.CondaConfigFile())
	must_be.Equal("", sut.CondaConfigFile())
	ok, err := sut.Validate()
	wont_be.True(ok)
	wont_be.Nil(err)
}

func TestCanReadActivityConfigFromBadText(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	sut, err := robot.ActivityPackageFrom([]byte(":"))
	wont_be.Nil(err)
	must_be.Nil(sut)
}
