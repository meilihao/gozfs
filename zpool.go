package gozfs

// #define __USE_LARGEFILE64
// #define _LARGEFILE_SOURCE
// #define _LARGEFILE64_SOURCE
// #include <libzfs.h>
// #include <sys/fs/zfs.h>
// #include <sys/stdtypes.h>
// #include <sys/nvpair.h>
// #include <stdlib.h>
// nvlist_t *nvlist_array_at(nvlist_t **a, uint_t i) {
// 	return a[i];
// }
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

type Pool struct {
	zph        *C.struct_zpool_handle
	Properties []Property
	Features   map[string]string
}

func NewPool(name string) (*Pool, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	zph := C.zpool_open_canfail(g_zfs, cName)
	if zph == nil {
		return nil, errors.New("pool not found")
	}

	return &Pool{
		zph: zph,
	}, LastError()
}

func (p *Pool) Free() {
	if p.zph != nil {
		C.zpool_close(p.zph)
	}
}

func (p *Pool) Name() string {
	return C.GoString(C.zpool_get_name(p.zph))
}

func (p *Pool) State() int {
	// for_each_pool + status_callback
	return int(C.zpool_get_state(p.zph))
}

func (p *Pool) GetProp(name string) (value string, err error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	zprop := C.zpool_name_to_prop(cName)

	var fprop [512]C.char
	r := C.zpool_get_prop(p.zph, zprop, &(fprop[0]), 512, nil, C.B_TRUE) // tree=p, C.B_TRUE=C.boolean_t(1)
	if r != 0 {
		err = errors.New(fmt.Sprint("invalid zpool property: ", name))
		return
	}

	value = C.GoString(&(fprop[0]))
	return
}

func (p *Pool) GetFeature(name string) (value string, err error) {
	cName := C.CString(fmt.Sprint("feature@", name))
	defer C.free(unsafe.Pointer(cName))

	var fvalue [512]C.char
	r := C.zpool_prop_get_feature(p.zph, cName, &(fvalue[0]), 512)
	if r != 0 {
		err = errors.New(fmt.Sprint("invalid zpool feature: ", name))
		return
	}

	value = C.GoString(&(fvalue[0]))
	return
}

func (p *Pool) VDevTree(name string) (vdevs any, err error) {
	config := C.zpool_get_config(p.zph, nil)
	if config == nil {
		err = errors.New("zpool_get_config failed")
		return
	}

	cCVT := C.CString(C.ZPOOL_CONFIG_VDEV_TREE)
	defer C.free(unsafe.Pointer(cCVT))

	nvroot := C.fnvlist_lookup_nvlist(config, cCVT) // fnvlist_lookup_nvlist: 查找名为 "cCVT" 的子 nvlist
	if nvroot == nil {
		err = errors.New("fnvlist_lookup_nvlist failed")
		return
	}

	return
}

func (p *Pool) poolGetSpares(nvroot *C.struct_nvlist) any {
	cS := C.CString(C.ZPOOL_CONFIG_SPARES)
	defer C.free(unsafe.Pointer(cS))

	var spares **C.struct_nvlist
	var nspares C.uint

	if C.nvlist_lookup_nvlist_array(nvroot, cS, &spares, &nspares) != 0 {
		return nil
	}

	if nspares == 0 {
		return nil
	}

	for i := C.uint(0); i < nspares; i++ {
		name := C.GoString(C.zpool_vdev_name(g_zfs, p.zph, C.nvlist_array_at(spares, i), C.B_TRUE))
		fmt.Println(name)
	}

	return nil
}
