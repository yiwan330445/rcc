package anywork

import (
	"fmt"
	"os"
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
	Scale(16)
	go watcher(failpipe, errcount)
}

func Scale(limit uint64) {
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

func Done() error {
	close(pipeline)
	return Sync()
}
