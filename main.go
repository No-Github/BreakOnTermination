package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	ProcessBreakOnTermination = 29
	PROCESS_ALL_ACCESS        = 0x1F0FFF // 手动定义 PROCESS_ALL_ACCESS 常量
	STATUS_SUCCESS            = 0x00000000
)

var (
	modntdll                    = windows.NewLazySystemDLL("ntdll.dll")
	procNtSetInformationProcess = modntdll.NewProc("NtSetInformationProcess")
)

func main() {
	// 获取要设置断点的进程名称
	processName := "notepad.exe"

	// 获取进程ID
	pid, err := findProcessIDByName(processName)
	if err != nil {
		fmt.Println("获取进程ID失败:", err)
		return
	}

	// 获取进程句柄
	hProcess, err := windows.OpenProcess(PROCESS_ALL_ACCESS, false, uint32(pid))
	if err != nil {
		fmt.Println("获取进程句柄失败:", err)
		return
	}
	defer windows.CloseHandle(hProcess)

	// 设置断点
	breakOnTermination := uint32(1)
	status, _, err := procNtSetInformationProcess.Call(
		uintptr(hProcess),
		uintptr(ProcessBreakOnTermination),
		uintptr(unsafe.Pointer(&breakOnTermination)),
		uintptr(unsafe.Sizeof(breakOnTermination)),
	)
	if status != STATUS_SUCCESS {
		fmt.Printf("设置进程断点失败: NTSTATUS=0x%x, 错误: %v\n", status, err)
		return
	}

	fmt.Println("设置进程断点成功")
}

// findProcessIDByName 查找指定进程名的进程ID
func findProcessIDByName(name string) (int, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(snapshot)

	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	if err := windows.Process32First(snapshot, &pe); err != nil {
		return 0, err
	}

	for {
		if windows.UTF16ToString(pe.ExeFile[:]) == name {
			return int(pe.ProcessID), nil
		}
		if err := windows.Process32Next(snapshot, &pe); err != nil {
			break
		}
	}

	return 0, fmt.Errorf("process not found")
}
