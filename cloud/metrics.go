package cloud

import (
	"fmt"
	"net/url"
	"runtime"
	"sync"
	"time"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/settings"
	"github.com/robocorp/rcc/xviper"
)

var (
	telemetryBarrier = sync.WaitGroup{}
)

const (
	trackingUrl = `/metric-v1/%v/%v/%v/%v/%v`
)

func sendMetric(metricsHost, kind, name, value string) {
	common.Timeline("%s:%s = %s", kind, name, value)
	defer func() {
		status := recover()
		if status != nil {
			common.Debug("Telemetry panic recovered: %v", status)
		}
		telemetryBarrier.Done()
	}()
	client, err := NewClient(metricsHost)
	if err != nil {
		common.Debug("ERROR: %v", err)
		return
	}
	timestamp := time.Now().UnixNano()
	url := fmt.Sprintf(trackingUrl, url.PathEscape(kind), timestamp, url.PathEscape(xviper.TrackingIdentity()), url.PathEscape(name), url.PathEscape(value))
	common.Debug("Sending metric as %v%v", metricsHost, url)
	client.Put(client.NewRequest(url))
}

func BackgroundMetric(kind, name, value string) {
	metricsHost := settings.Global.TelemetryURL()
	if len(metricsHost) == 0 {
		return
	}
	common.Debug("BackgroundMetric kind:%v name:%v value:%v send:%v", kind, name, value, xviper.CanTrack())
	if xviper.CanTrack() {
		telemetryBarrier.Add(1)
		go sendMetric(metricsHost, kind, name, value)
		runtime.Gosched()
	}
}

func WaitTelemetry() {
	common.Debug("wait telemetry to complete")
	telemetryBarrier.Wait()
	common.Debug("telemetry sending completed")
}
