package utils

import (
	"sync"
)

// LimitedBuffer is like a bytes.Buffer but dumps everything that exceeds maxLen
type LimitedBuffer struct {
	sync.Mutex
	i      int64 // current reading index
	idata  []byte
	maxLen int
}

func (b *LimitedBuffer) Write(in []byte) (int, error) {
	b.Lock()
	defer b.Unlock()

	if len(b.idata)+len(in) > b.maxLen {
		rem := b.maxLen - len(b.idata)
		if rem > 0 {
			b.idata = append(b.idata, in[:rem]...)
		}
	} else {
		b.idata = append(b.idata, in...)
	}

	return len(in), nil
}

func (b *LimitedBuffer) String() string {
	b.Lock()
	defer b.Unlock()
	return string(b.idata)
}

func (b *LimitedBuffer) Bytes() []byte {
	b.Lock()
	defer b.Unlock()
	return b.idata
}

func (b *LimitedBuffer) CopyBytes() []byte {
	cop := make([]byte, 0, b.maxLen/10)

	b.Lock()
	defer b.Unlock()

	cop = append(cop, b.idata...)

	return cop
}

// NewLimitedBuffer returns a new Limitedbuffer with a defined maxLen
func NewLimitedBuffer(maxLen int) *LimitedBuffer {
	return &LimitedBuffer{
		maxLen: maxLen,
		idata:  make([]byte, 0, 4096),
	}
}
