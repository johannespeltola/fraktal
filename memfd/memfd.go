package memfd

import (
	"syscall"
	"unsafe"
)

// Creates an anonymous file and returns a file descriptor (man memfd_create for info)
// Please not that this operation is potentially unsafe and only supported in Linux environments
func Create(name string, flags int) (fd int, err error) {
	// On Linux/amd64, memfd_create is syscall number 319.
	const SYS_MEMFD_CREATE = 319
	// Convert the name to a null-terminated string.
	bName, err := syscall.BytePtrFromString(name)
	if err != nil {
		return -1, err
	}
	r0, _, errno := syscall.Syscall(SYS_MEMFD_CREATE, uintptr(unsafe.Pointer(bName)), uintptr(flags), 0)
	if errno != 0 {
		return -1, errno
	}
	return int(r0), nil
}
