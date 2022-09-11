package utils

import "encoding/binary"

func BigEndianUint64(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}

func Varint(n int64) []byte {
	b := make([]byte, binary.MaxVarintLen64)
	return b[:binary.PutVarint(b, n)]
}
