//go:build windows

package process

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Windows API constants
const (
	TH32CS_SNAPPROCESS = 0x00000002
)

// Windows API structures
type PROCESSENTRY32 struct {
	DwSize              uint32
	CntUsage            uint32
	Th32ProcessID       uint32
	Th32DefaultHeapID   uintptr
	Th32ModuleID        uint32
	CntThreads          uint32
	Th32ParentProcessID uint32
	PcPriClassBase      int32
	DwFlags             uint32
	SzExeFile           [260]uint16
}

// Windows API functions
var (
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	CreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	Process32First      = kernel32.NewProc("Process32FirstW")
	Process32Next       = kernel32.NewProc("Process32NextW")
	CloseHandle         = kernel32.NewProc("CloseHandle")
)

// ListProcesses returns a list of running processes on Windows
func ListProcesses() ([]Process, error) {
	// Create snapshot of processes
	snapshot, _, _ := CreateToolhelp32Snapshot.Call(TH32CS_SNAPPROCESS, 0)
	if snapshot == 0 {
		return nil, fmt.Errorf("failed to create process snapshot")
	}
	defer CloseHandle.Call(snapshot)

	// Initialize process entry structure
	var entry PROCESSENTRY32
	entry.DwSize = uint32(unsafe.Sizeof(entry))

	// Get first process
	ret, _, _ := Process32First.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return nil, fmt.Errorf("failed to get first process")
	}

	var processes []Process
	
	// Add first process
	processes = append(processes, Process{
		PID:  int(entry.Th32ProcessID),
		Name: syscall.UTF16ToString(entry.SzExeFile[:]),
	})

	// Get remaining processes
	for {
		ret, _, _ := Process32Next.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			// Check if we've reached the end
			break
		}
		
		processes = append(processes, Process{
			PID:  int(entry.Th32ProcessID),
			Name: syscall.UTF16ToString(entry.SzExeFile[:]),
		})
	}

	return processes, nil
}