package common

import (
	"encoding/binary"
	"strings"

	"github.com/toastate/toastainer/internal/model"
	"github.com/toastate/toastainer/internal/utils"
)

func DumpToaterExeInfo(toaster *model.Toaster) []byte {
	var b []byte

	b = append(b, utils.BigEndianUint64(uint64(len(toaster.CodeID)))...)
	b = append(b, toaster.CodeID...)

	b = append(b, utils.BigEndianUint64(uint64(len(toaster.OwnerID)))...)
	b = append(b, toaster.OwnerID...)

	b = append(b, utils.BigEndianUint64(uint64(len(toaster.Image)))...)
	b = append(b, toaster.Image...)

	b = append(b, utils.BigEndianUint64(uint64(len(toaster.ExeCmd)))...)
	for i := 0; i < len(toaster.ExeCmd); i++ {
		b = append(b, utils.BigEndianUint64(uint64(len(toaster.ExeCmd[i])))...)
		b = append(b, toaster.ExeCmd[i]...)
	}

	joined := strings.Join(toaster.Env, ",")
	b = append(b, utils.BigEndianUint64(uint64(len(joined)))...)
	b = append(b, joined...)

	b = append(b, utils.BigEndianUint64(uint64(toaster.JoinableForSec))...)
	b = append(b, utils.BigEndianUint64(uint64(toaster.MaxConcurrentJoiners))...)
	b = append(b, utils.BigEndianUint64(uint64(toaster.TimeoutSec))...)

	return b
}

func ParseToasterExeInfo(dump []byte) *model.Toaster {
	toaster := &model.Toaster{}

	var offset int

	l := int(binary.BigEndian.Uint64(dump[offset : offset+8]))
	offset += 8

	toaster.CodeID = string(dump[offset : offset+l])
	offset += l

	l = int(binary.BigEndian.Uint64(dump[offset : offset+8]))
	offset += 8

	toaster.OwnerID = string(dump[offset : offset+l])
	offset += l

	l = int(binary.BigEndian.Uint64(dump[offset : offset+8]))
	offset += 8

	toaster.Image = string(dump[offset : offset+l])
	offset += l

	l = int(binary.BigEndian.Uint64(dump[offset : offset+8]))
	offset += 8

	toaster.ExeCmd = make([]string, l)
	for i := 0; i < len(toaster.ExeCmd); i++ {
		l = int(binary.BigEndian.Uint64(dump[offset : offset+8]))
		offset += 8

		toaster.ExeCmd[i] = string(dump[offset : offset+l])
		offset += l
	}

	l = int(binary.BigEndian.Uint64(dump[offset : offset+8]))
	offset += 8

	toaster.Env = strings.Split(string(dump[offset:offset+l]), ",")
	offset += l

	toaster.JoinableForSec = int(binary.BigEndian.Uint64(dump[offset : offset+8]))
	offset += 8

	toaster.MaxConcurrentJoiners = int(binary.BigEndian.Uint64(dump[offset : offset+8]))
	offset += 8

	toaster.TimeoutSec = int(binary.BigEndian.Uint64(dump[offset : offset+8]))
	offset += 8

	return toaster
}
