//go:build windows

package main

import (
	"log"
	"os"
	"syscall"
	"unsafe"
)

// webview_go selalu bikin jendela pakai icon generik Windows (IDI_APPLICATION)
// - lihat libs/webview/include/webview.h di dependency-nya - dan gak pernah
// baca resource icon custom yang di-embed lewat go-winres ke exe. Icon di
// File Explorer/Taskbar shortcut/installer sudah benar (baca resource exe
// langsung), tapi title bar jendela yang lagi jalan tetap harus di-set manual
// lewat WM_SETICON supaya ikut pakai icon yang sama.
const (
	wmSetIcon = 0x0080
	iconSmall = 0
	iconBig   = 1
)

var (
	shell32          = syscall.NewLazyDLL("shell32.dll")
	user32           = syscall.NewLazyDLL("user32.dll")
	procExtractIconW = shell32.NewProc("ExtractIconW")
	procSendMessageW = user32.NewProc("SendMessageW")
)

func setWindowIconFromExe(hwnd unsafe.Pointer) {
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("gagal ambil path exe buat set icon jendela: %v", err)
		return
	}
	pathPtr, err := syscall.UTF16PtrFromString(exePath)
	if err != nil {
		return
	}

	hIcon, _, _ := procExtractIconW.Call(0, uintptr(unsafe.Pointer(pathPtr)), 0)
	if hIcon == 0 || hIcon == 1 {
		return
	}
	procSendMessageW.Call(uintptr(hwnd), wmSetIcon, iconBig, hIcon)
	procSendMessageW.Call(uintptr(hwnd), wmSetIcon, iconSmall, hIcon)
}
