package monitor

import "time"

type Stats struct {
	CPU       CPUInfo       `json:"cpu"`
	Memory    MemoryInfo    `json:"memory"`
	Storage   StorageInfo   `json:"storage"`
	Network   NetworkInfo   `json:"network"`
	Temp      TempInfo      `json:"temperature"`
	Uptime    float64       `json:"uptime"`
	LoadAvg   LoadAvgInfo   `json:"load_avg"`
	Timestamp time.Time     `json:"timestamp"`
}

type CPUInfo struct {
	Usage    float64 `json:"usage"`
	NumCPU   int     `json:"num_cpu"`
	Model    string  `json:"model"`
	Arch     string  `json:"arch"`
	CoreTemp float64 `json:"core_temp"`
}

type MemoryInfo struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Available uint64  `json:"available"`
	Usage     float64 `json:"usage"`
}

type StorageInfo struct {
	Total uint64  `json:"total"`
	Used  uint64  `json:"used"`
	Free  uint64  `json:"free"`
	Usage float64 `json:"usage"`
	Path  string  `json:"path"`
}

type NetworkInfo struct {
	Interfaces []NetworkInterface `json:"interfaces"`
}

type NetworkInterface struct {
	Name    string `json:"name"`
	RXBytes uint64 `json:"rx_bytes"`
	TXBytes uint64 `json:"tx_bytes"`
	RXPkts  uint64 `json:"rx_pkts"`
	TXPkts  uint64 `json:"tx_pkts"`
	Status  string `json:"status"`
}

type TempInfo struct {
	Zones []ThermalZone `json:"zones"`
}

type ThermalZone struct {
	Name string  `json:"name"`
	Type string  `json:"type"`
 Temp float64 `json:"temp"`
}

type LoadAvgInfo struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}
