package peercc

import (
	"fmt"
	"net/url"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/set"
)

func knownSpec(knowledge []string, spec *htfs.ExportSpec) (*htfs.ExportSpec, *htfs.ExportSpec) {
	if !set.Member(knowledge, spec.Wants) {
		return nil, nil
	}
	delta := htfs.NewExportSpec(spec.Domain, spec.Wants, set.Intersect(knowledge, spec.Knows))
	if delta.IsRoot() {
		return nil, delta
	}
	return delta, delta.RootSpec()
}

func processQuery(available []string, query *Query) *htfs.ExportSpec {
	defer close(query.Reply)

	delta, root := knownSpec(available, query.Specification)
	if root == nil && delta == nil {
		return nil
	}
	selected := delta
	if selected == nil {
		selected = root
	}
	link, err := url.Parse(fmt.Sprintf("/hold/%s", selected.HoldName()))
	if err != nil {
		return delta
	}
	query.Reply <- link
	return delta
}

func frontdesk(catalogs Catalogs, holds Holdfiles, queries Queries, specs Specs) {
	common.Debug("Frontdesk starting ...")
	available := []string{}
	holding := []string{}
forever:
	for {
		select {
		case catalog, ok := <-catalogs:
			if !ok {
				break forever
			}
			available, _ = set.Update(available, catalog)
		case hold, ok := <-holds:
			if !ok {
				break forever
			}
			holding, _ = set.Update(holding, hold)
		case query, ok := <-queries:
			if !ok {
				break forever
			}
			delta := processQuery(available, query)
			if delta != nil {
				specs <- delta
			}
		}
	}
	common.Debug("Frontdesk stopping!")
}
