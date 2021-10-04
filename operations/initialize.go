package operations

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/robocorp/rcc/blobs"
	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/fail"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/settings"
	"gopkg.in/yaml.v1"
)

type StringMap map[string]string

type MetaTemplates struct {
	Date      string    `yaml:"date"`
	Hash      string    `yaml:"hash"`
	Templates StringMap `yaml:"templates"`
	Url       string    `yaml:"url"`
}

func TemplateInfo(filename string) (ingore *MetaTemplates, err error) {
	defer fail.Around(&err)

	raw, err := os.ReadFile(filename)
	fail.On(err != nil, "Failure reading %q, reason: %v", filename, err)
	var metadata MetaTemplates
	err = yaml.Unmarshal(raw, &metadata)
	fail.On(err != nil, "Failure parsing %q, reason: %v", filename, err)
	return &metadata, nil
}

func templatesYamlPart() string {
	return filepath.Join(common.TemplateLocation(), "templates.yaml.part")
}

func templatesYamlFinal() string {
	return filepath.Join(common.TemplateLocation(), "templates.yaml")
}

func templatesZipPart() string {
	return filepath.Join(common.TemplateLocation(), "templates.zip.part")
}

func templatesZipFinal() string {
	return filepath.Join(common.TemplateLocation(), "templates.zip")
}

func needNewTemplates() (ignore *MetaTemplates, err error) {
	defer fail.Around(&err)

	metadata := settings.Global.TemplatesYamlURL()
	if len(metadata) == 0 {
		common.Debug("No URL for templates.yaml available.")
		return nil, nil
	}
	partfile := templatesYamlPart()
	err = cloud.Download(metadata, partfile)
	fail.On(err != nil, "Failure loading %q, reason: %s", metadata, err)
	meta, err := TemplateInfo(partfile)
	fail.On(err != nil, "%s", err)
	fail.On(!strings.HasPrefix(meta.Url, "https:"), "Location for templates.zip is not https: %q", meta.Url)
	hash, err := pathlib.Sha256(templatesZipFinal())
	if err != nil || hash != meta.Hash {
		return meta, nil
	}
	return nil, nil
}

func downloadTemplatesZip(meta *MetaTemplates) (err error) {
	defer fail.Around(&err)

	partfile := templatesZipPart()
	err = cloud.Download(meta.Url, partfile)
	fail.On(err != nil, "Failure loading %q, reason: %s", meta.Url, err)
	hash, err := pathlib.Sha256(partfile)
	fail.On(err != nil, "Failure hashing %q, reason: %s", partfile, err)
	fail.On(hash != meta.Hash, "Received broken templates.zip from %q", meta.Hash)
	return nil
}

func updateTemplates() (err error) {
	defer fail.Around(&err)

	defer os.Remove(templatesZipPart())
	defer os.Remove(templatesYamlPart())

	meta, err := needNewTemplates()
	fail.On(err != nil, "%s", err)
	if meta == nil {
		return nil
	}
	err = downloadTemplatesZip(meta)
	fail.On(err != nil, "%s", err)
	err = os.Rename(templatesYamlPart(), templatesYamlFinal())
	alt := os.Rename(templatesZipPart(), templatesZipFinal())
	fail.On(alt != nil, "%s", alt)
	fail.On(err != nil, "%s", err)
	return nil
}

func unpack(content []byte, directory string) error {
	common.Debug("Initializing:")
	size := int64(len(content))
	byter := bytes.NewReader(content)
	reader, err := zip.NewReader(byter, size)
	if err != nil {
		return err
	}
	success := true
	for _, entry := range reader.File {
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
		return fmt.Errorf("Problems while initializing robot. Use --debug to see details.")
	}
	return nil
}

func ListTemplates() []string {
	err := updateTemplates()
	if err != nil {
		pretty.Warning("Problem updating templates.zip, reason: %v", err)
	}
	assets := blobs.AssetNames()
	result := make([]string, 0, len(assets))
	for _, name := range blobs.AssetNames() {
		if !strings.HasPrefix(name, "assets") || !strings.HasSuffix(name, ".zip") {
			continue
		}
		result = append(result, strings.TrimSuffix(filepath.Base(name), filepath.Ext(name)))
	}
	sort.Strings(result)
	return result
}

func InitializeWorkarea(directory, name string, force bool) error {
	content, err := blobs.Asset(fmt.Sprintf("assets/%s.zip", name))
	if err != nil {
		return err
	}
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
	UpdateRobot(fullpath)
	return unpack(content, fullpath)
}
