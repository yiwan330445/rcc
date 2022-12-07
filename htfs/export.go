package htfs

import (
	"fmt"
	"sort"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/set"
	"gopkg.in/yaml.v2"
)

type (
	ExportSpec struct {
		Domain string   `yaml:"domain"`
		Wants  string   `yaml:"wants"`
		Knows  []string `yaml:"knows"`
	}
)

func (it *ExportSpec) IsRoot() bool {
	return len(it.Knows) == 0
}

func (it *ExportSpec) RootSpec() *ExportSpec {
	if it.IsRoot() {
		return it
	}
	return NewExportSpec(it.Domain, it.Wants, []string{})
}

func (it *ExportSpec) HoldName() string {
	_, fingerprint, err := it.Fingerprint()
	if err != nil {
		return "broken.hld"
	}
	return fmt.Sprintf("%016x.hld", fingerprint)
}

func (it *ExportSpec) Fingerprint() (string, uint64, error) {
	sort.Strings(it.Knows)
	content, err := yaml.Marshal(it)
	if err != nil {
		return "", 0, err
	}
	fingerprint := common.Siphash(9007199254740993, 2147483647, content)
	return string(content), fingerprint, nil
}

func ParseExportSpec(content []byte) (*ExportSpec, error) {
	result := &ExportSpec{}
	err := yaml.Unmarshal(content, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func NewExportSpec(domain, wants string, knows []string) *ExportSpec {
	return &ExportSpec{
		Domain: domain,
		Wants:  wants,
		Knows:  set.Set(knows),
	}
}
