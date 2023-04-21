package jsonpb

import "unsafe"

func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

func bytesView(s []byte) string {
	return *(*string)(noescape(unsafe.Pointer(&s)))
}

func asBytes(s string) []byte {
	return *(*[]byte)(noescape(unsafe.Pointer(&[3]uintptr{
		*(*uintptr)(noescape(unsafe.Pointer(&s))),
		uintptr(len(s)),
		uintptr(len(s)),
	})))
}
