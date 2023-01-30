package remotree

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

func Serve(address string, port int, domain, storage string) error {
	// we need
	// - query handler (for just catalog hashes)
	// - partial content sender (for sending delta catalog)
	// - webserver
	holding := filepath.Join(storage, "hold")
	err := cleanupHoldStorage(holding)
	if err != nil {
		return err
	}
	defer cleanupHoldStorage(holding)

	triggers := make(chan string, 20)
	defer close(triggers)

	partqueries := make(Partqueries)
	defer close(partqueries)

	go listProvider(partqueries)
	go pullProcess(triggers)

	listen := fmt.Sprintf("%s:%d", address, port)
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:           listen,
		Handler:        mux,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   40 * time.Second,
		MaxHeaderBytes: 1 << 14,
	}

	mux.HandleFunc("/parts/", makeQueryHandler(partqueries, triggers))
	mux.HandleFunc("/delta/", makeDeltaHandler(partqueries))
	mux.HandleFunc("/force/", makeTriggerHandler(triggers))

	go server.ListenAndServe()

	return runTillSignal(server)
}

func runTillSignal(server *http.Server) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
	<-signals
	return server.Shutdown(context.TODO())
}
