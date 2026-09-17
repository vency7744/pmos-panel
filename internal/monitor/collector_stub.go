//go:build !linux

package monitor

import (
	"runtime"
	"time"
)

type Collector struct {
	storagePath string
}

func NewCollector(storagePath string) *Collector {
	return &Collector{storagePath: storagePath}
}

func (c *Collector) Collect() (*Stats, error) {
	return &Stats{
		Timestamp: time.Now(),
		CPU: CPUInfo{
			NumCPU: runtime.NumCPU(),
			Arch:   runtime.GOARCH,
			Model:  "unknown (non-linux)",
		},
		Memory: MemoryInfo{},
		Storage: StorageInfo{
			Path: c.storagePath,
		},
		Network: NetworkInfo{},
		Temp:    TempInfo{},
		Uptime:  0,
	}, nil
}
