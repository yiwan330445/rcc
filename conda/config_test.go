package conda_test

import (
	"os"
	"testing"

	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/hamlet"
)

func TestReadingMissingFileProducesError(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)
	text, err := conda.ReadConfig("some.yaml")
	wont_be.Nil(err)
	must_be.Text("", text)
	must_be.True(true)
	cwd, err := os.Getwd()
	wont_be.Nil(cwd)
	must_be.Nil(err)
}

func TestReadingCorrectFileProducesText(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)
	text, err := conda.ReadConfig("testdata/conda.yaml")
	must_be.Nil(err)
	wont_be.Text("", text)
	must_be.Equal(167, len(text))
}

func TestUnifyLineWorksCorrectly(t *testing.T) {
	must_be, _ := hamlet.Specifications(t)

	must_be.Text("foo", conda.UnifyLine("foo"))
	must_be.Text("foo", conda.UnifyLine(" foo"))
	must_be.Text("foo", conda.UnifyLine(" foo "))
	must_be.Text("foo", conda.UnifyLine(" foo \t"))
	must_be.Text("foo", conda.UnifyLine(" \tfoo \r\n \r\n"))
}

func TestSplitsLinesCorrectly(t *testing.T) {
	must_be, _ := hamlet.Specifications(t)

	must_be.Equal(1, len(conda.SplitLines("")))
	must_be.Equal(2, len(conda.SplitLines("\r\n")))
	must_be.Equal(2, len(conda.SplitLines("\n")))
	must_be.Equal([]string{"", ""}, conda.SplitLines("\r\n"))
	must_be.Equal([]string{"a", "b", "c", "d"}, conda.SplitLines("a\r\nb\r\nc\r\nd"))
}

func TestGetsLinesAsUnifiedCorrectly(t *testing.T) {
	must_be, _ := hamlet.Specifications(t)

	must_be.Equal(0, len(conda.AsUnifiedLines("")))
	must_be.Equal([]string{"a"}, conda.AsUnifiedLines("a\r\n\r\n"))
	must_be.Equal([]string{"a", "b"}, conda.AsUnifiedLines(" \r\n\tb \r\na\r\n\ta\t\r\n"))
}

func TestCanCalculateHash(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	expected := "8a900000303c000003000003030000000000033cc000000c030cf00000c03000000000"
	actual, err := conda.LocalitySensitiveHash([]string{"a", "b", "c"})
	must_be.Nil(err)
	wont_be.Nil(actual)
	must_be.Equal(expected, actual)
}

func TestCanCalculateHashEvenOnEmptySet(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	expected := "aa900000000c000000000000000000000000000c000000000000000000003000000000"
	actual, err := conda.LocalitySensitiveHash([]string{})
	must_be.Nil(err)
	wont_be.Nil(actual)
	must_be.Equal(expected, actual)
}

func TestCanCalculateHashEvenOnEmptyString(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	expected := "aa900000000c000000000000000000000000000c000000000000000000003000000000"
	actual, err := conda.LocalitySensitiveHash([]string{""})
	must_be.Nil(err)
	wont_be.Nil(actual)
	must_be.Equal(expected, actual)
}

func TestCanCalculateHashForConfig(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	actual, err := conda.HashConfig("missing/bad.yaml")
	wont_be.Nil(err)
	must_be.Equal("", actual)

	expected := "ded08c86224cc710b22228d3a1aa1a074bdf1a44f01be819c0a816044eebb80242030a"
	actual, err = conda.HashConfig("testdata/conda.yaml")
	must_be.Nil(err)
	must_be.Equal(expected, actual)

	other := "59c02b47324cc310a3332cc3a19a160b4bef0a04f02ff415c0f410044ddb780342030a"
	actual, err = conda.HashConfig("testdata/other.yaml")
	must_be.Nil(err)
	must_be.Equal(other, actual)

	third := "a8d08c86224cc710b22228c3a1aa1a0b4bef1a44f01fa815c0a412044aaa780242030a"
	actual, err = conda.HashConfig("testdata/third.yaml")
	must_be.Nil(err)
	must_be.Equal(third, actual)

	distance, err := conda.Distance(expected, other)
	must_be.Nil(err)
	must_be.Equal(94, distance)

	distance, err = conda.Distance(expected, third)
	must_be.Nil(err)
	must_be.Equal(13, distance)

	alien := "8a900000303c000003000003030000000000033cc000000c030cf00000c03000000000"

	distance, err = conda.Distance(expected, alien)
	must_be.Nil(err)
	must_be.Equal(384, distance)

	distance, err = conda.Distance("", alien)
	wont_be.Nil(err)
	must_be.Equal(999999, distance)

	distance, err = conda.Distance("iiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiii", alien)
	wont_be.Nil(err)
	must_be.Equal(0, distance)

	distance, err = conda.Distance(alien, "iiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiii")
	wont_be.Nil(err)
	must_be.Equal(0, distance)
}
