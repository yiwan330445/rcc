package journal

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/fail"
	"github.com/robocorp/rcc/pretty"
)

type (
	BuildEvent struct {
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
