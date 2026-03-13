package c

/*
#include <string.h>
#include <stdlib.h>
*/
import "C"
import "unsafe"

func Malloc(size int) unsafe.Pointer {
	return C.malloc(C.size_t(size))
}

func Free(ptr unsafe.Pointer) {
	C.free(ptr)
}

func Memmove(dst, src unsafe.Pointer, length int) {
	C.memmove(dst, src, C.size_t(length))
}

func Memcpy(dst unsafe.Pointer, src []byte, length int) {
	srcData := C.CBytes(src)
	C.memcpy(dst, srcData, C.size_t(length))
}
