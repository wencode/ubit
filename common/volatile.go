package common

import (
	"runtime/volatile"
)

func Volatile32_GetAndClear(reg32 *volatile.Register32) uint32 {
	v := reg32.Get()
	if v != 0 {
		reg32.Set(0)
	}
	return v
}
