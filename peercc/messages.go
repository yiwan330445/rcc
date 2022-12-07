package peercc

import (
	"net/url"

	"github.com/robocorp/rcc/htfs"
)

type (
	URLs  chan *url.URL
	Query struct {
		Specification *htfs.ExportSpec
		Reply         URLs
	}
	Queries   chan *Query
	Catalogs  chan string
	Holdfiles chan string
	Specs     chan *htfs.ExportSpec
	Partquery struct {
		Catalog string
		Reply   chan string
	}
	Partqueries chan *Partquery
)
