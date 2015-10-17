package plugin

/*
#cgo LDFLAGS: -lrt
#include <stdlib.h>
#include <sys/mman.h>
*/
import "C"

import (
	"os"
	"syscall"
	"unsafe"
)

type region struct {
	name   string
	bytes  []byte
	fd     *os.File
	unlink bool
}

// Unmaps the shared memory, closes the region and - if the region owns the
// underlying shm object - signals that the shm object can be unlinked
// when its unmapped by all processes.
func (r *region) Close() error {
	err := syscall.Munmap(r.bytes)
	if err != nil {
		rpcLog.Error("Failed to unmap memory: %s", err.Error())
		return err
	}

	err = r.fd.Close()
	if err != nil {
		rpcLog.Error("Failed to close shared memory handle: %s", err.Error())
		return err
	}

	if r.unlink {
		cname := C.CString(r.name)
		defer C.free(unsafe.Pointer(cname))
		_, err = C.shm_unlink(cname)
		if err != nil {
			rpcLog.Error("Failed to mark shm block for deletion: %s", err.Error())
			return err
		}
	}

	return nil
}

// Creates a new region instance by attaching to an existing shm object
// and mapping it into the process address space. This region does not "own"
// the underlying shm object, and will not call shm_unlink on it when the
// region is closed.
func OpenRegion(name string) (result region, err error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	fd, err := C.shm_open(cname, C.int(os.O_RDWR), 0)
	if err != nil {
		return
	}

	file := os.NewFile(uintptr(fd), name)
	defer func() {
		if err != nil {
			file.Close()
		}
	}()

	shmInfo, err := file.Stat()
	if err != nil {
		rpcLog.Error("Failed to stat the shared memory block: %s", err.Error())
		return
	}

	buf, err := syscall.Mmap(int(fd), 0, int(shmInfo.Size()),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED)
	if err != nil {
		return
	}

	result = region{
		bytes:  buf,
		fd:     file,
		name:   name,
		unlink: false,
	}

	return
}

// Creates a new region of shared memory by creating and mapping in a new
// shm object. The returned region "owns" the shm object, and will mark it for
// unlinking when the region is closed.
func NewRegion(name string, length int) (result region, err error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	fd, err := C.shm_open(
		cname,
		C.int(os.O_CREATE|os.O_EXCL|os.O_RDWR),
		0600)
	if err != nil {
		return
	}

	file := os.NewFile(uintptr(fd), name)
	defer func() {
		if err != nil {
			file.Close()
		}
	}()

	err = syscall.Ftruncate(int(fd), int64(length))
	if err != nil {
		return
	}

	buf, err := syscall.Mmap(int(fd), 0, length,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED)
	if err != nil {
		return
	}

	result = region{
		bytes:  buf,
		fd:     file,
		name:   name,
		unlink: true}
	return
}
