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
		content, ok := <-reply
		common.Debug("query handler: %q -> %v", catalog, ok)
		if !ok {
			response.WriteHeader(http.StatusNotFound)
			response.Write([]byte("404 not found, sorry"))
			return
		}

		members := strings.Split(content, "\n")

		requested := make([]string, 0, 1000)
		todo := bufio.NewReader(request.Body)
	todoloop:
		for {
			line, err := todo.ReadString('\n')
			if err == io.EOF {
				break todoloop
			}
			if err != nil {
				common.Debug("DELTA: %v with %q", err, line)
				break todoloop
			}
			flat := strings.TrimSpace(line)
			member := set.Member(members, flat)
			if !member {
				common.Trace("DELTA: ignoring extra %q entry, not part of set!", flat)
				continue todoloop
			}
			requested = append(requested, flat)
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

		for _, flat := range requested {
			relative := htfs.RelativeDefaultLocation(flat)
			fullpath := htfs.ExactDefaultLocation(flat)
			err = operations.ZipAppend(sink, fullpath, relative)
			if err != nil {
				common.Debug("DELTA: error %v with %v -> %v", err, fullpath, relative)
				return
			}
		}
	}
}
