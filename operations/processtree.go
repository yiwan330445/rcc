package operations

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/mitchellh/go-ps"
	"github.com/robocorp/rcc/pretty"
)

type (
	ChildMap    map[int]string
	ProcessMap  map[int]*ProcessNode
	ProcessNode struct {
		Pid        int
		Parent     int
		Executable string
		Children   ProcessMap
	}
)

func NewProcessNode(core ps.Process) *ProcessNode {
	return &ProcessNode{
		Pid:        core.Pid(),
		Parent:     core.PPid(),
		Executable: core.Executable(),
		Children:   make(ProcessMap),
	}
}

func ProcessMapNow() (ProcessMap, error) {
	processes, err := ps.Processes()
	if err != nil {
		return nil, err
	}
	result := make(ProcessMap)
	for _, process := range processes {
		result[process.Pid()] = NewProcessNode(process)
	}
	for pid, process := range result {
		parent, ok := result[process.Parent]
		if ok {
			parent.Children[pid] = process
		}
	}
	return result, nil
}

func (it ProcessMap) Keys() []int {
	keys := make([]int, 0, len(it))
	for key, _ := range it {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	return keys
}

func (it *ProcessNode) warnings(additional ProcessMap) {
	if len(it.Children) > 0 {
		pretty.Warning("%q process %d still has running subprocesses:", it.Executable, it.Pid)
		it.warningTree("> ", false)
	} else {
		pretty.Warning("%q process %d still has running migrated processes:", it.Executable, it.Pid)
	}
	if len(additional) > 0 {
		pretty.Warning("+ migrated process still running:")
		for _, zombie := range additional {
			zombie.warningTree("| ", true)
		}
	}
	pretty.Note("Depending on OS, above processes may prevent robot to close properly.")
	pretty.Note("Few reasons why this might be happening are:")
	pretty.Note("- robot is not properly releasing all resources that it is using")
	pretty.Note("- robot is generating background processes that don't complete before robot tries to exit")
	pretty.Note("- there was failure inside robot, which caused robot to exit without proper cleanup")
	pretty.Note("- developer intentionally left processes running, which is not good for repeatable automation")
	pretty.Highlight("So if you see this message, and robot still seems to be running, it is not!")
	pretty.Highlight("You now have to take action and stop those processes that are preventing robot to complete.")
}

func (it *ProcessNode) warningTree(prefix string, newparent bool) {
	kind := "leaf"
	if len(it.Children) > 0 {
		kind = "container"
	}
	if newparent {
		kind = fmt.Sprintf("%s -> new parent PID: #%d", kind, it.Parent)
	}
	pretty.Warning("%s#%d  %q <%s>", prefix, it.Pid, it.Executable, kind)
	indent := prefix + "|   "
	for _, key := range it.Children.Keys() {
		it.Children[key].warningTree(indent, false)
	}
}

func SubprocessWarning(seen ChildMap, use bool) error {
	processes, err := ProcessMapNow()
	if err != nil {
		return err
	}
	self, ok := processes[os.Getpid()]
	if !ok {
		return fmt.Errorf("For some reason, could not find own process in process map.")
	}
	masked := make(ChildMap)
	if use {
		for pid, executable := range seen {
			ref, ok := processes[pid]
			if ok {
				updateActiveChildren(ref, masked, 70)
			} else {
				masked[pid] = executable
			}
		}
	}
	for key, _ := range masked {
		delete(seen, key)
	}
	additional := make(ProcessMap)
	for pid, executable := range seen {
		ref, ok := processes[pid]
		if ok && executable == ref.Executable {
			additional[pid] = ref
		}
	}
	if len(self.Children) > 0 || len(additional) > 0 {
		self.warnings(additional)
	}
	return nil
}

func removeStaleChildren(processes ProcessMap, seen ChildMap) {
	for key, name := range seen {
		found, ok := processes[key]
		if !ok || found.Executable != name {
			delete(seen, key)
		}
	}
}

func updateActiveChildren(host *ProcessNode, seen ChildMap, maxDepth int) {
	if maxDepth < 0 {
		return
	}
	for pid, child := range host.Children {
		seen[pid] = child.Executable
		updateActiveChildren(child, seen, maxDepth-1)
	}
}

func updateSeenChildren(pid int, processes ProcessMap, seen ChildMap) {
	source, ok := processes[pid]
	if ok {
		removeStaleChildren(processes, seen)
		updateActiveChildren(source, seen, 70)
	}
}

func WatchChildren(pid int, delay time.Duration) chan ChildMap {
	pipe := make(chan ChildMap)
	go babySitter(pid, pipe, delay)
	return pipe
}

func babySitter(pid int, reply chan ChildMap, delay time.Duration) {
	defer close(reply)
	seen := make(ChildMap)
	failures := 0
forever:
	for failures < 10 {
		processes, err := ProcessMapNow()
		if err == nil {
			updateSeenChildren(pid, processes, seen)
			failures = 0
		}
		select {
		case reply <- seen:
			break forever
		case <-time.After(delay):
			continue forever
		}
	}
}
