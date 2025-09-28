//go:build !(windows || linux || darwin || freebsd || openbsd || netbsd)

package process

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// ListProcesses returns a list of running processes using external commands
func ListProcesses() ([]Process, error) {
	switch runtime.GOOS {
	case "windows":
		return listProcessesWindows()
	default:
		return listProcessesUnix()
	}
}

func listProcessesWindows() ([]Process, error) {
	// Use tasklist command on Windows
	cmd := exec.Command("tasklist", "/fo", "csv", "/nh")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute tasklist: %w", err)
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var processes []Process
	
	for _, line := range lines {
		// Parse CSV line: "name","pid","session","session#","mem usage"
		fields := strings.Split(line, "\",\"")
		if len(fields) < 2 {
			continue
		}
		
		// Remove quotes
		name := strings.Trim(fields[0], "\"")
		pidStr := strings.Trim(fields[1], "\"")
		
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}
		
		processes = append(processes, Process{PID: pid, Name: name})
	}
	
	return processes, nil
}

func listProcessesUnix() ([]Process, error) {
	// Use ps command on Unix systems
	cmd := exec.Command("ps", "-eo", "pid,comm")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute ps: %w", err)
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	// Skip header line
	if len(lines) < 2 {
		return []Process{}, nil
	}
	
	var processes []Process
	for _, line := range lines[1:] { // Skip header
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		
		name := strings.Join(fields[1:], " ")
		processes = append(processes, Process{PID: pid, Name: name})
	}
	
	return processes, nil
}