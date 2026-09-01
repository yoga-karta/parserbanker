//go:build !windows

package main

import "unsafe"

// Cuma relevan di Windows (lihat window_icon_windows.go) - platform lain
// gak butuh WM_SETICON manual.
func setWindowIconFromExe(hwnd unsafe.Pointer) {}
