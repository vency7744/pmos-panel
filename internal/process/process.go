package process

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

type Process struct {
	PID    int     `json:"pid"`
	User   string  `json:"user"`
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"mem"`
	Command string `json:"command"`
}

func List() ([]Process, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("read /proc: %w", err)
	}

	var processes []Process
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		proc, err := readProcPID(pid)
		if err != nil {
			continue
		}
		processes = append(processes, *proc)
	}
	return processes, nil
}

func readProcPID(pid int) (*Process, error) {
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	data, err := os.ReadFile(statPath)
	if err != nil {
		return nil, err
	}

	p := &Process{PID: pid}

	content := string(data)
	idx := strings.LastIndex(content, ")")
	if idx == -1 {
		return nil, fmt.Errorf("invalid stat format")
	}
	fields := strings.Fields(content[idx+2:])

	commPath := fmt.Sprintf("/proc/%d/comm", pid)
	if commData, err := os.ReadFile(commPath); err == nil {
		p.Command = strings.TrimSpace(string(commData))
	}

	if len(fields) > 11 {
		utime, _ := strconv.ParseFloat(fields[11], 64)
		stime, _ := strconv.ParseFloat(fields[12], 64)
		p.CPU = (utime + stime) / 100.0
	}

	if len(fields) > 20 {
		rss, _ := strconv.ParseUint(fields[20], 10, 64)
		p.Memory = float64(rss*4096) / (1024 * 1024)
	}

	statusPath := fmt.Sprintf("/proc/%d/status", pid)
	if statusData, err := os.ReadFile(statusPath); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(statusData)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "Uid:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					uid, _ := strconv.Atoi(fields[1])
					p.User = resolveUID(uid)
				}
				break
			}
		}
	}

	return p, nil
}

func resolveUID(uid int) string {
	data, err := os.ReadFile("/etc/passwd")
	if err != nil {
		return strconv.Itoa(uid)
	}
	uidStr := strconv.Itoa(uid)
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) >= 3 && fields[2] == uidStr {
			return fields[0]
		}
	}
	return uidStr
}

func Kill(pid int) error {
	path := fmt.Sprintf("/proc/%d", pid)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("process %d not found", pid)
	}
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to kill process %d: %w", pid, err)
	}
	return nil
}
