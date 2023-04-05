package operations

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/xviper"

	"gopkg.in/yaml.v2"
)

type Folder struct {
	Path    string `yaml:"path" json:"robot"`
	Created int64  `yaml:"created" json:"created"`
	Updated int64  `yaml:"updated" json:"updated"`
	Deleted int64  `yaml:"deleted" json:"deleted"`
}

type Credential struct {
	Account  string `yaml:"account"`
	Context  string `yaml:"context"`
	Token    string `yaml:"token"`
	Deadline int64  `yaml:"deadline"`
}

type FolderMap map[string]*Folder
type CredentialMap map[string]*Credential
type StampMap map[string]int64

type Cache struct {
	Robots      FolderMap     `yaml:"robots"`
	Credentials CredentialMap `yaml:"credentials"`
	Stamps      StampMap      `yaml:"stamps"`
}

func (it Cache) Ready() *Cache {
	if it.Robots == nil {
		it.Robots = make(FolderMap)
	}
	if it.Credentials == nil {
		it.Credentials = make(CredentialMap)
	}
	if it.Stamps == nil {
		it.Stamps = make(StampMap)
	}
	return &it
}

func cacheLockFile() string {
	return fmt.Sprintf("%s.lck", cacheLocation())
}

func cacheLocation() string {
	reference := xviper.ConfigFileUsed()
	if len(reference) > 0 {
		return filepath.Join(filepath.Dir(reference), "rcccache.yaml")
	} else {
		return filepath.Join(common.RobocorpHome(), "rcccache.yaml")
	}
}

func SummonCache() (*Cache, error) {
	var result Cache
	lockfile := cacheLockFile()
	completed := pathlib.LockWaitMessage(lockfile, "Serialized cache access [cache lock]")
	locker, err := pathlib.Locker(lockfile, 125)
	completed()
	if err != nil {
		return nil, err
	}
	defer locker.Release()

	cacheFile := cacheLocation()
	source, err := os.Open(cacheFile)
	if err != nil {
		return result.Ready(), nil
	}
	defer source.Close()
	defer pathlib.RestrictOwnerOnly(cacheFile)
	decoder := yaml.NewDecoder(source)
	err = decoder.Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Ready(), nil
}

func (it *Cache) Save() error {
	lockfile := cacheLockFile()
	completed := pathlib.LockWaitMessage(lockfile, "Serialized cache access [cache lock]")
	locker, err := pathlib.Locker(lockfile, 125)
	completed()
	if err != nil {
		return err
	}
	defer locker.Release()

	cacheFile := cacheLocation()
	sink, err := pathlib.Create(cacheFile)
	if err != nil {
		return err
	}
	defer sink.Close()
	defer pathlib.RestrictOwnerOnly(cacheFile)
	encoder := yaml.NewEncoder(sink)
	err = encoder.Encode(it)
	if err != nil {
		return err
	}
	return nil
}
