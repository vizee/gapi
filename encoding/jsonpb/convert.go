package jsonpb

import "unsafe"

func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

func bytesView(s []byte) string {
	return *(*string)(noescape(unsafe.Pointer(&s)))
}
