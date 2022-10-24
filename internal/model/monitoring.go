package model

import (
	"encoding/binary"
	"fmt"
)

type ServerMonitoring struct {
	TotalMemoryBytes uint64
	FreeMemoryBytes  uint64

	TotalCPUSeconds   uint64
	FreeCPUPercentage uint64 // base 10 000 not 100!!!

	TotalDiskBytes uint64
	FreeDiskBytes  uint64

	ToasterStartupLatencyNanoseconds uint64
}

func (sm *ServerMonitoring) Marshal() []byte {
	b := make([]byte, 56)

	binary.BigEndian.PutUint64(b[0:8], sm.TotalMemoryBytes)
	binary.BigEndian.PutUint64(b[8:16], sm.FreeMemoryBytes)
	binary.BigEndian.PutUint64(b[16:24], sm.TotalCPUSeconds)
	binary.BigEndian.PutUint64(b[24:32], sm.FreeCPUPercentage)
	binary.BigEndian.PutUint64(b[32:40], sm.TotalDiskBytes)
	binary.BigEndian.PutUint64(b[40:48], sm.FreeDiskBytes)
	binary.BigEndian.PutUint64(b[48:56], sm.ToasterStartupLatencyNanoseconds)

	return b
}

func UnmarshalServerMonitoring(b []byte) (*ServerMonitoring, error) {
	if len(b) != 56 {
		return nil, fmt.Errorf("invalid length")
	}

	return &ServerMonitoring{
		TotalMemoryBytes:                 binary.BigEndian.Uint64(b[0:8]),
		FreeMemoryBytes:                  binary.BigEndian.Uint64(b[8:16]),
		TotalCPUSeconds:                  binary.BigEndian.Uint64(b[16:24]),
		FreeCPUPercentage:                binary.BigEndian.Uint64(b[24:32]),
		TotalDiskBytes:                   binary.BigEndian.Uint64(b[32:40]),
		FreeDiskBytes:                    binary.BigEndian.Uint64(b[40:48]),
		ToasterStartupLatencyNanoseconds: binary.BigEndian.Uint64(b[48:56]),
	}, nil
}
