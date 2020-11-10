package common

var (
	Silent    bool
	Debug     bool
	Trace     bool
	Separator bool
	NoCache   bool
)

const (
	DefaultEndpoint = "https://api.eu1.robocloud.eu/"
)

func UnifyVerbosityFlags() {
	if Silent {
		Debug = false
		Trace = false
	}
	if Trace {
		Debug = true
	}
}
