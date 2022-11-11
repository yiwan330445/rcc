package journal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/fail"
	"github.com/robocorp/rcc/pretty"
)

type (
	acceptor func(BuildEvent) bool
	picker   func(BuildEvent) float64
	prettify func(float64) string

	Numbers     []float64
	BuildEvents []BuildEvent
	BuildEvent  struct {
		When          int64  `json:"when"`
		What          string `json:"what"`
		Force         bool   `json:"force"`
		Build         bool   `json:"build"`
		Success       bool   `json:"success"`
		Retry         bool   `json:"retry"`
		Run           bool   `json:"run"`
		Controller    string `json:"controller"`
		Space         string `json:"space"`
		BlueprintHash string `json:"blueprint"`

		Started         float64 `json:"started"`
		Prepared        float64 `json:"prepared"`
		MicromambaDone  float64 `json:"micromamba"`
		PipDone         float64 `json:"pip"`
		PostInstallDone float64 `json:"postinstall"`
		RecordDone      float64 `json:"record"`
		RestoreDone     float64 `json:"restore"`
		PreRunDone      float64 `json:"prerun"`
		RobotStart      float64 `json:"robotstart"`
		RobotEnd        float64 `json:"robotend"`
		Finished        float64 `json:"finished"`
		Dirtyness       float64 `json:"dirtyness"`
	}
)

var (
	buildevent *BuildEvent
)

func init() {
	buildevent = NewBuildEvent()
}

func asPercent(value float64) string {
	return fmt.Sprintf("%5.1f%%", value)
}

func asSecond(value float64) string {
	return fmt.Sprintf("%7.3fs", value)
}

func started(the BuildEvent) float64 {
	return the.Started
}

func prepared(the BuildEvent) float64 {
	return the.Prepared
}

func micromamba(the BuildEvent) float64 {
	return the.MicromambaDone
}

func pip(the BuildEvent) float64 {
	return the.PipDone
}

func postinstall(the BuildEvent) float64 {
	return the.PostInstallDone
}

func record(the BuildEvent) float64 {
	return the.RecordDone
}

func restore(the BuildEvent) float64 {
	return the.RestoreDone
}

func prerun(the BuildEvent) float64 {
	return the.PreRunDone
}

func ShowStatistics() {
	stats, err := Stats()
	if err != nil {
		pretty.Warning("Loading statistics failed, reason: %v", err)
		return
	}
	tabbed := tabwriter.NewWriter(os.Stderr, 2, 4, 2, ' ', tabwriter.AlignRight)
	tabbed.Write(sprint("Selected statistics:\t%d samples\t\n", len(stats)))
	tabbed.Write([]byte("\n"))
	tabbed.Write([]byte("Name \tAverage\t10%\tMedian\t90%\tMAX\t\n"))
	stats.Statsline(tabbed, "Dirty", asPercent, func(the BuildEvent) float64 {
		return the.Dirtyness
	})
	tabbed.Write([]byte("\n"))
	tabbed.Write([]byte("Name                  \tAverage\t10%\tMedian\t90%\tMAX\t\n"))
	stats.Statsline(tabbed, "Lead time             ", asSecond, func(the BuildEvent) float64 {
		return the.Started
	})
	stats.Statsline(tabbed, "Setup time            ", asSecond, func(the BuildEvent) float64 {
		if the.RobotStart > 0 {
			return the.RobotStart - the.Started
		}
		return the.Finished - the.Started
	})
	stats.Statsline(tabbed, "Holospace restore time", asSecond, func(the BuildEvent) float64 {
		if the.RestoreDone > 0 {
			return the.RestoreDone - the.Started
		}
		return the.Finished - the.Started
	})
	stats.Statsline(tabbed, "Pre-run               ", asSecond, func(the BuildEvent) float64 {
		if the.PreRunDone > 0 {
			return the.PreRunDone - the.RestoreDone
		}
		return 0
	})
	stats.Statsline(tabbed, "Robot startup delay   ", asSecond, func(the BuildEvent) float64 {
		return the.RobotStart
	})
	stats.Statsline(tabbed, "Robot execution time  ", asSecond, func(the BuildEvent) float64 {
		return the.RobotEnd - the.RobotStart
	})
	stats.Statsline(tabbed, "Total execution time  ", asSecond, func(the BuildEvent) float64 {
		return the.Finished
	})
	onlyBuilds := stats.filter(func(the BuildEvent) bool {
		return the.Build
	})
	tabbed.Write([]byte("\n\n"))
	percentage := 100.0 * float64(len(onlyBuilds)) / float64(len(stats))
	tabbed.Write(sprint("%d\tsamples with environment builds\t(%3.1f%% from selected)\t\n", len(onlyBuilds), percentage))
	tabbed.Write([]byte("\n"))
	tabbed.Write([]byte("Name                  \tAverage\t10%\tMedian\t90%\tMAX\t\n"))
	onlyBuilds.Statsline(tabbed, "Phase: prepare        ", asSecond, func(the BuildEvent) float64 {
		return the.Prepared - the.Started
	})
	onlyBuilds.Statsline(tabbed, "Phase: micromamba     ", asSecond, func(the BuildEvent) float64 {
		return the.MicromambaDone - the.Prepared
	})
	onlyBuilds.Statsline(tabbed, "Phase: pip            ", asSecond, func(the BuildEvent) float64 {
		if the.PipDone > 0 {
			return the.PipDone - the.MicromambaDone
		}
		return 0.0
	})
	onlyBuilds.Statsline(tabbed, "Phase: post install   ", asSecond, func(the BuildEvent) float64 {
		if the.PostInstallDone > 0 {
			return the.PostInstallDone - the.first(pip, micromamba)
		}
		return 0.0
	})
	onlyBuilds.Statsline(tabbed, "Phase: record         ", asSecond, func(the BuildEvent) float64 {
		return the.RecordDone - the.first(postinstall, pip, micromamba)
	})
	onlyBuilds.Statsline(tabbed, "To hololib            ", asSecond, func(the BuildEvent) float64 {
		if the.RecordDone > 0 {
			return the.RecordDone - the.Started
		}
		return 0.0
	})
	tabbed.Flush()
}

func CurrentBuildEvent() *BuildEvent {
	return buildevent
}

func CurrentEventFilename() string {
	year, week := common.Clock.Time().ISOWeek()
	filename := fmt.Sprintf("stats_%s_%04d_%02d.log", common.UserHomeIdentity(), year, week)
	return filepath.Join(common.JournalLocation(), filename)
}

func BuildEventStats(label string) {
	err := serialize(buildevent.finished(label))
	if err != nil {
		pretty.Warning("build stats for %q failed, reason: %v", label, err)
	}
}

func serialize(event *BuildEvent) (err error) {
	defer fail.Around(&err)

	blob, err := json.Marshal(event)
	fail.On(err != nil, "Could not serialize event: %v -> %v", event.What, err)
	return appendJournal(CurrentEventFilename(), blob)
}

func NewBuildEvent() *BuildEvent {
	return &BuildEvent{
		When: common.Clock.When(),
	}
}

func (it *BuildEvent) stowatch() float64 {
	return common.Clock.Elapsed().Seconds()
}

func (it *BuildEvent) finished(label string) *BuildEvent {
	it.What = label
	it.Finished = it.stowatch()
	it.Controller = common.ControllerType
	it.Space = common.HolotreeSpace
	return it
}

func (it *BuildEvent) Successful() {
	it.Success = true
}

func (it *BuildEvent) StartNow(force bool) {
	buildevent.Started = it.stowatch()
	buildevent.Force = force
}

func (it *BuildEvent) Blueprint(blueprint string) {
	buildevent.BlueprintHash = blueprint
}

func (it *BuildEvent) Rebuild() {
	buildevent.Retry = true
	buildevent.Build = true
}

func (it *BuildEvent) PrepareComplete() {
	buildevent.Build = true
	buildevent.Prepared = it.stowatch()
}

func (it *BuildEvent) MicromambaComplete() {
	buildevent.Build = true
	buildevent.MicromambaDone = it.stowatch()
}

func (it *BuildEvent) PipComplete() {
	buildevent.Build = true
	buildevent.PipDone = it.stowatch()
}

func (it *BuildEvent) PostInstallComplete() {
	buildevent.Build = true
	buildevent.PostInstallDone = it.stowatch()
}

func (it *BuildEvent) RecordComplete() {
	buildevent.RecordDone = it.stowatch()
}

func (it *BuildEvent) Dirty(dirtyness float64) {
	buildevent.Dirtyness = dirtyness
}

func (it *BuildEvent) RestoreComplete() {
	buildevent.RestoreDone = it.stowatch()
}

func (it *BuildEvent) PreRunComplete() {
	buildevent.Run = true
	buildevent.PreRunDone = it.stowatch()
}

func (it *BuildEvent) RobotStarts() {
	buildevent.Run = true
	buildevent.RobotStart = it.stowatch()
}

func (it *BuildEvent) RobotEnds() {
	buildevent.Run = true
	buildevent.RobotEnd = it.stowatch()
}

func (it BuildEvent) first(tools ...picker) float64 {
	for _, tool := range tools {
		value := tool(it)
		if value > 0 {
			return value
		}
	}
	return 0
}

func (it BuildEvents) filter(query acceptor) BuildEvents {
	result := make(BuildEvents, 0, len(it))
	for _, event := range it {
		if query(event) {
			result = append(result, event)
		}
	}
	return result
}

func (it BuildEvents) pick(tool picker) Numbers {
	result := make(Numbers, 0, len(it))
	for _, event := range it {
		result = append(result, tool(event))
	}
	return result
}

func sprint(form string, fields ...any) []byte {
	return []byte(fmt.Sprintf(form, fields...))
}

func (it BuildEvents) Statsline(tabbed *tabwriter.Writer, label string, nice prettify, tool picker) {
	numbers := it.pick(tool)
	sort.Float64s(numbers)
	average, low, median, high, last := numbers.Statsline()
	tabbed.Write(sprint("%s\t%s\t%s\t%s\t%s\t%s\t\n", label, nice(average), nice(low), nice(median), nice(high), nice(last)))
}

func (it Numbers) safe(at int) float64 {
	total := len(it)
	if total == 0 {
		return 0.0
	}
	if at < 0 {
		return it[0]
	}
	if at < total {
		return it[at]
	}
	return it[total-1]
}

func (it Numbers) Statsline() (average, low, median, high, worst float64) {
	total := len(it)
	if total < 1 {
		return
	}
	sum := 0.0
	for _, value := range it {
		sum += value
	}
	half := total >> 1
	percentile := total / 10
	last := total - 1
	right := last - percentile
	return sum / float64(total), it.safe(percentile), it.safe(half), it.safe(right), it.safe(last)
}

func Stats() (result BuildEvents, err error) {
	defer fail.Around(&err)
	journalname := CurrentEventFilename()
	handle, err := os.Open(journalname)
	fail.On(err != nil, "Failed to open event journal %v -> %v", journalname, err)
	defer handle.Close()
	source := bufio.NewReader(handle)
	fail.On(err != nil, "Failed to read %s.", journalname)
	result = make(BuildEvents, 0, 100)
	for {
		line, err := source.ReadBytes('\n')
		if err == io.EOF {
			return result, nil
		}
		fail.On(err != nil, "Failed to read %s.", journalname)
		event := BuildEvent{}
		err = json.Unmarshal(line, &event)
		if err != nil {
			continue
		}
		result = append(result, event)
	}
}
