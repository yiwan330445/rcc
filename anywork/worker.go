package anywork

import (
	"fmt"
	"os"
	"sync"
)

var (
	group     *sync.WaitGroup
	pipeline  WorkQueue
	headcount uint64
)

type Work func()
type WorkQueue chan Work

func catcher(title string, identity uint64) {
	catch := recover()
	if catch != nil {
		fmt.Fprintf(os.Stderr, "Recovering %q #%d: %v\n", title, identity, catch)
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

func init() {
	group = &sync.WaitGroup{}
	pipeline = make(WorkQueue, 100000)
	headcount = 1
	go member(headcount)
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

func Sync() {
	group.Wait()
}

func Done() {
	close(pipeline)
	group.Wait()
}
