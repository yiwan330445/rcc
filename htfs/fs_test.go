package htfs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/robocorp/rcc/hamlet"
	"github.com/robocorp/rcc/htfs"
)

func TestHTFSspecification(t *testing.T) {
	must, wont := hamlet.Specifications(t)

	filename := filepath.Join(os.TempDir(), "htfs_test.json")

	fs, err := htfs.NewRoot("..")
	must.Nil(err)
	wont.Nil(fs)
	wont.Nil(fs.Tree)

	must.Nil(fs.Lift())

	content, err := fs.AsJson()
	must.Nil(err)
	must.True(len(content) > 50000)

	must.Nil(fs.SaveAs(filename))

	reloaded, err := htfs.NewRoot(".")
	must.Nil(err)
	wont.Nil(reloaded)
	before, err := reloaded.AsJson()
	must.Nil(err)
	must.True(len(before) < 300)
	wont.Equal(fs.Path, reloaded.Path)

	must.Nil(reloaded.LoadFrom(filename))
	after, err := reloaded.AsJson()
	must.Nil(err)
	must.Equal(len(after), len(content))
	must.Equal(fs.Path, reloaded.Path)
}

func TestZipLibrary(t *testing.T) {
	must, wont := hamlet.Specifications(t)

	_, blueprint, err := htfs.ComposeFinalBlueprint([]string{"testdata/simple.yaml"}, "")
	must.Nil(err)
	wont.Nil(blueprint)
	sut, err := htfs.ZipLibrary("testdata/simple.zip")
	must.Nil(err)
	wont.Nil(sut)
	must.True(sut.HasBlueprint(blueprint))
}
