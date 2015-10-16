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

func shmCreate(name string, length int) (result *region, err error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	fd, err := C.shm_open(C.CString(name), C.int(os.O_CREATE|os.O_RDWR), 0600)
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

	result = &region{bytes: buf, fd: file, name: name, unlink: true}
	return
}
