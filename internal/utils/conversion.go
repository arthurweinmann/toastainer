package utils

import "unsafe"

// Do not apply it after a String2ByteSlice was applied
func ByteSlice2String(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}

// Do not apply it after a ByteSlice2String was applied
func String2ByteSlice(str string) []byte {
	return *(*[]byte)(unsafe.Pointer(&str))
}
