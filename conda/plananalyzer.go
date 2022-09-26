package conda

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pretty"
)

const (
	newline = '\n'
	spacing = "\r\n\t "
)

var (
	planPattern = regexp.MustCompile("^---  (.+?) plan @\\d+.\\d+s  ---$")
)

type (
	AnalyzerStrategy func(*PlanAnalyzer, string)
	StrategyMap      map[string]AnalyzerStrategy
	RepeatCache      map[string]bool

	PlanAnalyzer struct {
		Strategies StrategyMap
		Active     AnalyzerStrategy
		Notes      []string
		Pending    []byte
		Repeats    RepeatCache
		Realtime   bool
		Details    bool
		Started    time.Time
	}
)

func NewPlanAnalyzer(realtime bool) *PlanAnalyzer {
	strategies := make(StrategyMap)
	strategies["micromamba"] = ignoreStrategy
	strategies["post install"] = ignoreStrategy
	strategies["activation"] = ignoreStrategy
	strategies["pip check"] = ignoreStrategy
	strategies["pip"] = pipStrategy
	return &PlanAnalyzer{
		Strategies: strategies,
		Active:     ignoreStrategy,
		Notes:      []string{},
		Pending:    nil,
		Repeats:    make(RepeatCache),
		Realtime:   realtime,
		Details:    false,
	}
}

func pipStrategy(ref *PlanAnalyzer, event string) {
	low := strings.ToLower(event)
	note := ""
	detail := ""
	if strings.HasPrefix(low, "info:") || strings.HasPrefix(low, "error:") {
		note = event
	}
	if strings.Contains(low, "using cached") {
		if strings.Contains(low, ".tar.gz") {
			detail = fmt.Sprintf("%s [missing wheel file?]", event)
		} else {
			detail = event
		}
	}
	elapsed := time.Since(ref.Started).Round(1 * time.Second)
	if len(note) > 0 {
		ref.Notes = append(ref.Notes, note)
		if ref.Realtime {
			pretty.Warning("%s  @%s", note, elapsed)
		}
		ref.Details = true
		return
	}
	if ref.Details && len(detail) > 0 {
		ref.Notes = append(ref.Notes, detail)
		if ref.Realtime {
			pretty.Note("%s  @%s", detail, elapsed)
		}
		return
	}
	if ref.Realtime {
		common.Trace("PIP: %s", event)
	}
}

func ignoreStrategy(ref *PlanAnalyzer, event string) {
	// does nothing by default
}

func (it *PlanAnalyzer) Observe(event string) {
	found := planPattern.FindStringSubmatch(event)
	if len(found) > 1 {
		it.Active = ignoreStrategy
		strategy, ok := it.Strategies[found[1]]
		if ok {
			it.Active = strategy
		}
		it.Repeats = make(RepeatCache)
		it.Details = false
		it.Started = time.Now()
	}
	it.Active(it, event)
}

func (it *PlanAnalyzer) Write(blob []byte) (int, error) {
	old := len(it.Pending)
	update := len(blob)
	var total uint64 = uint64(old) + uint64(update)
	body := make([]byte, 0, total)
	if old > 0 {
		body = append(body, it.Pending...)
	}
	if update > 0 {
		body = append(body, blob...)
	}
	terminator := []byte{newline}
	parts := bytes.SplitAfter(body, terminator)
	size := len(parts)
	last := parts[size-1]
	terminated := bytes.HasSuffix(last, terminator)
	if !terminated {
		it.Pending = last
		parts = parts[:size-1]
	} else {
		it.Pending = nil
	}
	for _, part := range parts {
		it.Observe(strings.TrimRight(string(part), spacing))
	}
	return update, nil
}

func (it *PlanAnalyzer) Close() {
	if len(it.Notes) == 0 || it.Realtime {
		return
	}
	pretty.Warning("Analyzing installation plan revealed following findings:")
	for _, note := range it.Notes {
		common.Log("  %s* %s%s%s", pretty.Cyan, pretty.Bold, note, pretty.Reset)
	}
}
