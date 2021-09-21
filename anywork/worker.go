package anywork

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
)

var (
	group     *sync.WaitGroup
	pipeline  WorkQueue
	failpipe  Failures
	errcount  Counters
	headcount uint64
)

type Work func()
type WorkQueue chan Work
type Failures chan string
type Counters chan uint64

func catcher(title string, identity uint64) {
	catch := recover()
	if catch != nil {
		failpipe <- fmt.Sprintf("Recovering %q #%d: %v", title, identity, catch)
	}
}

func process(fun Work, identity uint64) {
	defer group.Done()
	defer catcher("process", identity)
	fun()
}

func member(identity uint64) {
	defer catcher("member", identity)
	for {
		work, ok := <-pipeline
		if !ok {
			break
		}
		process(work, identity)
	}
}

func watcher(failures Failures, counters Counters) {
	counter := uint64(0)
	for {
		select {
		case fail := <-failures:
			counter += 1
			fmt.Fprintln(os.Stderr, fail)
		case counters <- counter:
			counter = 0
		}
	}
}

func init() {
	group = &sync.WaitGroup{}
	pipeline = make(WorkQueue, 100000)
	failpipe = make(Failures)
	errcount = make(Counters)
	headcount = 0
	AutoScale()
	go watcher(failpipe, errcount)
}

func Scale() uint64 {
	return headcount
}

func AutoScale() {
	limit := uint64(runtime.NumCPU() - 1)
	if limit > 96 {
		limit = 96
	}
	if limit < 4 {
		limit = 4
	}
	for headcount < limit {
		go member(headcount)
		headcount += 1
	}
}

func Backlog(todo Work) {
	if todo != nil {
		group.Add(1)
		pipeline <- todo
	}
}

func Sync() error {
	group.Wait()
	count := <-errcount
	if count > 0 {
		return fmt.Errorf("There has been %d failures. See messages above.", count)
	}
	return nil
}

func OnErrPanicCloseAll(err error, closers ...io.Closer) {
	if err != nil {
		for _, closer := range closers {
			if closer != nil {
				closer.Close()
			}
		}
		panic(err)
	}
}

func Done() error {
	close(pipeline)
	return Sync()
}
