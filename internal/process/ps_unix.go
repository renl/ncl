//go:build (linux || darwin || freebsd || openbsd || netbsd) && !windows

package process

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

// ListProcesses returns a list of running processes on Unix-like systems
func ListProcesses() ([]Process, error) {
	var processes []Process
	
	// Read /proc directory
	procDir, err := ioutil.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc directory: %w", err)
	}
	
	for _, entry := range procDir {
		// Only consider directories with numeric names (PIDs)
		if !entry.IsDir() {
			continue
		}
		
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue // Not a PID directory
		}
		
		// Read process name from /proc/[pid]/comm
		commPath := fmt.Sprintf("/proc/%d/comm", pid)
		commBytes, err := ioutil.ReadFile(commPath)
		if err != nil {
			// Try to read from cmdline if comm is not available
			cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
			cmdlineBytes, err := ioutil.ReadFile(cmdlinePath)
			if err != nil {
				continue
			}
			
			// cmdline is null-separated, take first argument
			cmdline := strings.Split(string(cmdlineBytes), "\x00")[0]
			if len(cmdline) > 0 {
				// Extract just the executable name
				parts := strings.Split(cmdline, "/")
				name := parts[len(parts)-1]
				processes = append(processes, Process{PID: pid, Name: name})
			}
			continue
		}
		
		name := strings.TrimSpace(string(commBytes))
		processes = append(processes, Process{PID: pid, Name: name})
	}
	
	return processes, nil
}