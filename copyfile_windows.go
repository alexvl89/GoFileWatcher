//go:build windows
// +build windows

package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
    modkernel32   = windows.NewLazySystemDLL("kernel32.dll")
    procCopyFileW = modkernel32.NewProc("CopyFileW")
)

func copyFile(src, dst string) error {
    srcPtr, err := windows.UTF16PtrFromString(src)
    if err != nil {
        return err
    }
    dstPtr, err := windows.UTF16PtrFromString(dst)
    if err != nil {
        return err
    }

    ret, _, callErr := procCopyFileW.Call(
        uintptr(unsafe.Pointer(srcPtr)),
        uintptr(unsafe.Pointer(dstPtr)),
        uintptr(0),
    )
    if ret == 0 {
        return fmt.Errorf("copy file failed for %s to %s: %w", src, dst, callErr)
    }
    return nil
}
