package xviper

import (
	"math"
	"time"
)

const (
	runMinutesStats = `stats.rccminutes`
)

type runMarker time.Time

func RunMinutes() runMarker {
	return runMarker(time.Now())
}

func (it runMarker) Done() uint64 {
	delta := time.Now().Sub(time.Time(it))
	minutes := uint64(math.Max(1.0, math.Ceil(delta.Minutes())))
	previous := GetUint64(runMinutesStats)
	total := previous + minutes
	Set(runMinutesStats, total)
	return total
}
