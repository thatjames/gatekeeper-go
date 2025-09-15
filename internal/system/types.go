package system

type SystemInfo map[string]string

type InterfaceStatistics struct {
	TxBytes uint64
	RxBytes uint64
}

var Version = "development-build"
