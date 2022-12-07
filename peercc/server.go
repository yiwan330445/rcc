package peercc

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func stopper(server *http.Server) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
	<-signals
	return server.Shutdown(context.TODO())
}

func Serve(address string, port int, domain, storage string) error {
	// we need
	// - builder
	// - holder
	// - webserver
	// - download handler (for optimists)
	// - specification handler (for pessimists)
	holding := filepath.Join(storage, "hold")
	err := cleanupHoldStorage(holding)
	if err != nil {
		return err
	}
	defer cleanupHoldStorage(holding)

	holds := make(Holdfiles)
	specs := make(Specs, 30)
	queries := make(Queries)
	catalogs := make(Catalogs)
	partqueries := make(Partqueries)
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	defer close(holds)
	defer close(specs)
	defer close(queries)
	defer close(catalogs)

	go listProvider(partqueries)
	//go feedInitialSpecs(domain, specs)
	//go frontdesk(catalogs, holds, queries, specs)
	//go builder(holding, specs, catalogs, holds)

	listen := fmt.Sprintf("%s:%d", address, port)
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:           listen,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 14,
	}

	mux.HandleFunc("/parts/", makeQueryHandler(partqueries))

	go server.ListenAndServe()

	<-signals

	return server.Shutdown(context.TODO())
}
