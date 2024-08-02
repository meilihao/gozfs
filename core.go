package gozfs

// #cgo CFLAGS: -I/usr/include/libzfs -I/usr/include/libspl
// #cgo LDFLAGS: -lzfs
// #define __USE_LARGEFILE64 // [Required for spl stat.h](https://github.com/johnramsden/zectl/issues/33)
// #define _LARGEFILE_SOURCE
// #define _LARGEFILE64_SOURCE
// #include <libzfs.h>
// #include <errno.h>
// int getErrno() {
//      return errno;
// }
import "C"
import (
	"errors"
	"fmt"
)

var (
	g_zfs *C.struct_libzfs_handle
)

func Init() error {
	// https://github.com/openzfs/zfs/blob/master/cmd/zpool/zpool_main.c#L11787
	g_zfs = C.libzfs_init()
	if g_zfs == nil {
		// [no C.errno](https://github.com/golang/go/commit/02327a72d7ae27c16ab4ed702138ca6a818e6123)
		return fmt.Errorf("libzfs_init failed: %s", C.GoString(C.libzfs_error_init(C.getErrno())))
	}

	return nil
}

func LastError() error {
	return errors.New(C.GoString(C.libzfs_error_description(g_zfs)))
}

type Property struct {
	Value  string
	Source string
}
