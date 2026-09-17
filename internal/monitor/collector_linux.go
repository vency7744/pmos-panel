//go:build linux

package monitor

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Collector struct {
	prevCPU    cpuTimes
	prevNet    map[string]netCounters
	storagePath string
}

type cpuTimes struct {
	user, nice, system, idle, iowait, irq, softirq, steal uint64
}

type netCounters struct {
	rxBytes, txBytes, rxPkts, txPkts uint64
}

func NewCollector(storagePath string) *Collector {
	return &Collector{
		prevNet:     make(map[string]netCounters),
		storagePath: storagePath,
	}
}

func (c *Collector) Collect() (*Stats, error) {
	stats := &Stats{
		Timestamp: time.Now(),
	}

	stats.CPU = c.collectCPU()
	stats.Memory = c.collectMemory()
	stats.Storage = c.collectStorage()
	stats.Network = c.collectNetwork()
	stats.Temp = c.collectTemp()
	stats.Uptime = c.collectUptime()
	stats.LoadAvg = c.collectLoadAvg()

	return stats, nil
}

func (c *Collector) collectCPU() CPUInfo {
	info := CPUInfo{
		NumCPU: runtime.NumCPU(),
		Arch:   runtime.GOARCH,
	}

	if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "model name") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					info.Model = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}

	times := c.readCPUTimes()
	if c.prevCPU.idle > 0 {
		prev := c.prevCPU
		dUser := times.user - prev.user
		dNice := times.nice - prev.nice
		dSystem := times.system - prev.system
		dIdle := times.idle - prev.idle
		dIowait := times.iowait - prev.iowait
		dIrq := times.irq - prev.irq
		dSoftirq := times.softirq - prev.softirq
		dSteal := times.steal - prev.steal

		total := dUser + dNice + dSystem + dIdle + dIowait + dIrq + dSoftirq + dSteal
		if total > 0 {
			idle := dIdle + dIowait
			info.Usage = float64(total-idle) / float64(total) * 100
		}
	}
	c.prevCPU = times

	return info
}

func (c *Collector) readCPUTimes() cpuTimes {
	var t cpuTimes
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return t
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) >= 9 {
				t.user, _ = strconv.ParseUint(fields[1], 10, 64)
				t.nice, _ = strconv.ParseUint(fields[2], 10, 64)
				t.system, _ = strconv.ParseUint(fields[3], 10, 64)
				t.idle, _ = strconv.ParseUint(fields[4], 10, 64)
				t.iowait, _ = strconv.ParseUint(fields[5], 10, 64)
				t.irq, _ = strconv.ParseUint(fields[6], 10, 64)
				t.softirq, _ = strconv.ParseUint(fields[7], 10, 64)
				t.steal, _ = strconv.ParseUint(fields[8], 10, 64)
			}
			break
		}
	}
	return t
}

func (c *Collector) collectMemory() MemoryInfo {
	info := MemoryInfo{}
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return info
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		val, _ := strconv.ParseUint(fields[1], 10, 64)
		val *= 1024

		switch fields[0] {
		case "MemTotal:":
			info.Total = val
		case "MemAvailable:":
			info.Available = val
		}
	}

	info.Used = info.Total - info.Available
	if info.Total > 0 {
		info.Usage = float64(info.Used) / float64(info.Total) * 100
	}

	return info
}

func (c *Collector) collectStorage() StorageInfo {
	info := StorageInfo{Path: c.storagePath}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(c.storagePath, &stat); err != nil {
		return info
	}

	info.Total = stat.Blocks * uint64(stat.Bsize)
	info.Free = stat.Bfree * uint64(stat.Bsize)
	info.Used = info.Total - info.Free
	if info.Total > 0 {
		info.Usage = float64(info.Used) / float64(info.Total) * 100
	}

	return info
}

func (c *Collector) collectNetwork() NetworkInfo {
	info := NetworkInfo{}

	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return info
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum <= 2 {
			continue
		}

		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		fields := strings.Fields(parts[1])
		if len(fields) < 10 {
			continue
		}

		absRXBytes, _ := strconv.ParseUint(fields[0], 10, 64)
		absRXPkts, _ := strconv.ParseUint(fields[1], 10, 64)
		absTXBytes, _ := strconv.ParseUint(fields[8], 10, 64)
		absTXPkts, _ := strconv.ParseUint(fields[9], 10, 64)

		statusPath := fmt.Sprintf("/sys/class/net/%s/operstate", name)
		status := ""
		if statusData, err := os.ReadFile(statusPath); err == nil {
			status = strings.TrimSpace(string(statusData))
		}

		iface := NetworkInterface{
			Name:   name,
			Status: status,
		}

		if prev, ok := c.prevNet[name]; ok {
			iface.RXBytes = absRXBytes - prev.rxBytes
			iface.TXBytes = absTXBytes - prev.txBytes
			iface.RXPkts = absRXPkts - prev.rxPkts
			iface.TXPkts = absTXPkts - prev.txPkts
		}

		c.prevNet[name] = netCounters{
			rxBytes: absRXBytes,
			txBytes: absTXBytes,
			rxPkts:  absRXPkts,
			txPkts:  absTXPkts,
		}

		info.Interfaces = append(info.Interfaces, iface)
	}

	return info
}

func (c *Collector) collectTemp() TempInfo {
	info := TempInfo{}

	matches, err := filepath.Glob("/sys/class/thermal/thermal_zone*")
	if err != nil {
		return info
	}

	for _, zonePath := range matches {
		zone := ThermalZone{}

		if data, err := os.ReadFile(filepath.Join(zonePath, "type")); err == nil {
			zone.Type = strings.TrimSpace(string(data))
		}

		zone.Name = filepath.Base(zonePath)

		if data, err := os.ReadFile(filepath.Join(zonePath, "temp")); err == nil {
			temp, _ := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
			zone.Temp = temp / 1000.0
		}

		info.Zones = append(info.Zones, zone)
	}

	return info
}

func (c *Collector) collectUptime() float64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}

	fields := strings.Fields(string(data))
	if len(fields) > 0 {
		uptime, _ := strconv.ParseFloat(fields[0], 64)
		return uptime
	}

	return 0
}

func (c *Collector) collectLoadAvg() LoadAvgInfo {
	info := LoadAvgInfo{}
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return info
	}

	fields := strings.Fields(string(data))
	if len(fields) >= 3 {
		info.Load1, _ = strconv.ParseFloat(fields[0], 64)
		info.Load5, _ = strconv.ParseFloat(fields[1], 64)
		info.Load15, _ = strconv.ParseFloat(fields[2], 64)
	}

	return info
}
