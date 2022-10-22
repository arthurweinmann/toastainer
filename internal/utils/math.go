package utils

import "golang.org/x/exp/constraints"

func Min[T constraints.Ordered](x, y T) T {
	if x < y {
		return x
	}
	return y
}

func Max[T constraints.Ordered](x, y T) T {
	if x > y {
		return x
	}
	return y
}

func IsPowerOf2(x int) bool {
	return (x != 0) && (x&(x-1)) == 0
}

func NextHighestPowerOf2(x uint32) uint32 {
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x++
	return x
}
