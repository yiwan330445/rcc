package peercc

import (
	"path/filepath"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/pathlib"
)

func feedInitialSpecs(domain string, specs Specs) {
	for _, catalog := range htfs.Catalogs() {
		specs <- htfs.NewExportSpec(domain, catalog, []string{})
	}
}

func buildSpecToStorage(storage string, spec *htfs.ExportSpec) (string, bool) {
	holdfile := spec.HoldName()

	defer common.Stopwatch("Build of spec %q -> %q took", spec.Wants, holdfile).Debug()
	tree, err := htfs.New()
	if err != nil {
		return "", false
	}
	archive := filepath.Join(storage, holdfile)
	err = tree.Export([]string{spec.Wants}, spec.Knows, archive)
	if err != nil {
		return "", false
	}
	return holdfile, true
}

func builder(storage string, specs Specs, catalogs Catalogs, holds Holdfiles) {
	common.Debug("Builder for %q starting ...", storage)
	pathlib.EnsureDirectoryExists(storage)
forever:
	for {
		todo, ok := <-specs
		if !ok {
			break forever
		}
		if todo == nil {
			continue
		}
		common.Debug("Build for %q requested.", todo.HoldName())
		holdfile, ok := buildSpecToStorage(storage, todo)
		if ok {
			catalogs <- todo.Wants
			holds <- holdfile
			common.Debug("Build for %q done.", todo.HoldName())
		}
	}
	common.Debug("Builder stopped!")
}
