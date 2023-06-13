package operations

import (
	"fmt"
	"os"
	"sort"

	"github.com/mitchellh/go-ps"
	"github.com/robocorp/rcc/pretty"
)

type (
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

func (it *ProcessNode) warnings() {
	pretty.Warning("%q process %d still has running subprocesses:", it.Executable, it.Pid)
	it.warningTree("> ")
	pretty.Note("Depending on OS, above processes may prevent robot to close properly.")
	pretty.Note("Few reasons why this might be happening are:")
	pretty.Note("- robot is not properly releasing all resources that it is using")
	pretty.Note("- there was failure inside robot, which caused robot to exit without proper cleanup")
	pretty.Note("- developer intentionally left processes running, which is not good for repeatable automation")
	pretty.Highlight("So if you see this message, and robot still seems to be running, it is not!")
	pretty.Highlight("You now have to take action and stop those processes that are preventing robot to complete.")
}

func (it *ProcessNode) warningTree(prefix string) {
	kind := "leaf"
	if len(it.Children) > 0 {
		kind = "container"
	}
	pretty.Warning("%s%d  %q <%s>", prefix, it.Pid, it.Executable, kind)
	indent := prefix + "|   "
	for _, key := range it.Children.Keys() {
		it.Children[key].warningTree(indent)
	}
}

func SubprocessWarning() error {
	processes, err := ProcessMapNow()
	if err != nil {
		return err
	}
	self, ok := processes[os.Getpid()]
	if !ok {
		return fmt.Errorf("For some reason, could not find own process in process map.")
	}
	if len(self.Children) > 0 {
		self.warnings()
	}
	return nil
}
