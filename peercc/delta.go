package peercc

import (
	"archive/zip"
	"bufio"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/operations"
	"github.com/robocorp/rcc/set"
)

func makeDeltaHandler(queries Partqueries) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		catalog := filepath.Base(request.URL.Path)
		defer common.Stopwatch("Delta of catalog %q took", catalog).Debug()
		if request.Method != http.MethodPost {
			response.WriteHeader(http.StatusMethodNotAllowed)
			common.Trace("Delta: rejecting request %q for catalog %q.", request.Method, catalog)
			return
		}
		reply := make(chan string)
		queries <- &Partquery{
			Catalog: catalog,
			Reply:   reply,
		}
		known, ok := <-reply
		common.Debug("query handler: %q -> %v", catalog, ok)
		if !ok {
			response.WriteHeader(http.StatusNotFound)
			response.Write([]byte("404 not found, sorry"))
			return
		}

		members := strings.Split(known, "\n")

		approved := make([]string, 0, 1000)
		todo := bufio.NewReader(request.Body)
	todoloop:
		for {
			line, err := todo.ReadString('\n')
			stopping := err == io.EOF
			candidate := filepath.Base(strings.TrimSpace(line))
			if len(candidate) > 0 {
				if set.Member(members, candidate) {
					approved = append(approved, candidate)
				} else {
					common.Trace("DELTA: ignoring extra %q entry, not part of set!", candidate)
					if !stopping {
						continue todoloop
					}
				}
			}
			if stopping {
				break todoloop
			}
			if err != nil {
				common.Trace("DELTA: error %v with line %q", err, line)
				break todoloop
			}
		}

		headers := response.Header()
		headers.Add("Content-Type", "application/zip")
		response.WriteHeader(http.StatusOK)

		sink := zip.NewWriter(response)
		defer sink.Close()
		fullpath := filepath.Join(common.HololibCatalogLocation(), catalog)
		relative, err := filepath.Rel(common.HololibLocation(), fullpath)
		if err != nil {
			common.Debug("DELTA: error %v", err)
			return
		}
		err = operations.ZipAppend(sink, fullpath, relative)
		if err != nil {
			common.Debug("DELTA: error %v", err)
			return
		}

		for _, member := range approved {
			relative := htfs.RelativeDefaultLocation(member)
			fullpath := htfs.ExactDefaultLocation(member)
			err = operations.ZipAppend(sink, fullpath, relative)
			if err != nil {
				common.Debug("DELTA: error %v with %v -> %v", err, fullpath, relative)
				return
			}
		}
	}
}
