package blobs_test

import (
	"testing"

	"github.com/robocorp/rcc/blobs"
	"github.com/robocorp/rcc/hamlet"
	"github.com/robocorp/rcc/operations"
)

func TestCanSeeBaseZipAsset(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	must_be.Panic(func() { blobs.MustAsset("assets/missing.zip") })
	wont_be.Panic(func() { blobs.MustAsset("assets/templates.yaml") })
	wont_be.Panic(func() { blobs.MustAsset("assets/standard.zip") })
	wont_be.Panic(func() { blobs.MustAsset("assets/python.zip") })
	wont_be.Panic(func() { blobs.MustAsset("assets/settings.yaml") })

	_, err := blobs.Asset("assets/missing.zip")
	wont_be.Nil(err)

	asset, err := blobs.Asset("assets/standard.zip")
	must_be.Nil(err)
	wont_be.Nil(asset)
}

func TestCanGetTemplateNamesThruOperations(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	assets := operations.ListTemplates(true)
	wont_be.Nil(assets)
	must_be.True(len(assets) == 3)
	must_be.Equal([]string{"extended", "python", "standard"}, assets)
}
