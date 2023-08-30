package operations

import (
	"fmt"
	"os"
	"time"

	"github.com/mitchellh/go-ps"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/set"
)

type (
	ChildMap     map[int]string
	ProcessMap   map[int]*ProcessNode
	ProcessNodes []*ProcessNode
	ProcessNode  struct {
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
	return set.Keys(it)
}

func (it ProcessMap) Roots() []int {
	roots := []int{}
	for candidate, node := range it {
		_, ok := it[node.Parent]
		if !ok {
			roots = append(roots, candidate)
		}
	}
	return set.Sort(roots)
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
		for _, key := range additional.Roots() {
			additional[key].warningTree("| ", true)
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
	pretty.Highlight("Example cleanup command: %s", common.GenerateKillCommand(additional.Keys()))
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
	common.RunJournal("orphan process", fmt.Sprintf("parent=%d pid=%d name=%s", it.Parent, it.Pid, it.Executable), "process pollution")
	indent := prefix + "|   "
	for _, key := range it.Children.Keys() {
		it.Children[key].warningTree(indent, false)
	}
}

func SubprocessWarning(seen ChildMap, use bool) error {
	before := len(seen)
	if before == 0 {
		common.Debug("No tracked subprocesses, which is a good thing.")
		return nil
	}
	processes, err := ProcessMapNow()
	if err != nil {
		return err
	}
	removeStaleChildren(processes, seen)
	after := len(seen)
	pretty.DebugNote("Final subprocess count %d -> %d. %v", before, after, seen)
	if after == 0 {
		common.Debug("No active tracked subprocesses anymore, and that is a good thing.")
		return nil
	}
	self, ok := processes[os.Getpid()]
	if !ok {
		return fmt.Errorf("For some reason, could not find own process in process map.")
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

func removeStaleChildren(processes ProcessMap, seen ChildMap) bool {
	removed := false
	for key, name := range seen {
		found, ok := processes[key]
		if !ok || found.Executable != name {
			delete(seen, key)
			removed = true
		}
	}
	return removed
}

func updateActiveChildrenLoop(start *ProcessNode, seen ChildMap) bool {
	updated := false
	counted := make(map[int]bool)
	counted[start.Pid] = true
	at, todo := 0, ProcessNodes{start}
	for at < len(todo) {
		for pid, child := range todo[at].Children {
			if counted[pid] {
				continue
			}
			counted[pid] = true
			_, previously := seen[pid]
			seen[pid] = child.Executable
			todo = append(todo, child)
			if !previously {
				updated = true
			}
		}
		at += 1
	}
	return updated
}

func updateSeenChildren(pid int, processes ProcessMap, seen ChildMap) bool {
	source, ok := processes[pid]
	if ok {
		removed := removeStaleChildren(processes, seen)
		updated := updateActiveChildrenLoop(source, seen)
		return removed || updated
	}
	return false
}

func WatchChildren(pid int, delay time.Duration) chan ChildMap {
	pipe := make(chan ChildMap)
	go babySitter(pid, pipe, delay)
	return pipe
}

func babySitter(pid int, reply chan ChildMap, delay time.Duration) {
	defer close(reply)
	seen := make(ChildMap)
	failures, broadcasted := 0, 0
	defer common.RunJournal("processes", "final", "count: %d", broadcasted)
forever:
	for failures < 10 {
		updated := false
		processes, err := ProcessMapNow()
		if err == nil {
			updated = updateSeenChildren(pid, processes, seen)
			failures = 0
		} else {
			common.Debug("Process snapshot failure: %v", err)
		}
		if updated {
			active := len(seen)
			pretty.DebugNote("Active subprocess count %d -> %d. %v", broadcasted, active, seen)
			common.RunJournal("processes", "updated", "count from %d to %d ... %v", broadcasted, active, seen)
			broadcasted = active
		}
		select {
		case reply <- seen:
			break forever
		case <-time.After(delay):
			continue forever
		}
	}
	common.Debug("Final active subprocess count was %d.", broadcasted)
}
