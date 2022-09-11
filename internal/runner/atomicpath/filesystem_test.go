package atomicpath

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestAtomicFileSystem(t *testing.T) {
	filesystem := NewFileSystem(8, gc)

	if !filesystem.CreationLock("bob", 4, false) {
		panic("could not acquire lock")
	}

	if filesystem.CreationLock("bob", 4, false) {
		panic("should not been able to acquire lock")
	}

	var exited uint32

	for i := 0; i < 10; i++ {
		go func() {
			filesystem.CreationLock("bob", 4, true)
			atomic.AddUint32(&exited, 1)
		}()
	}

	time.Sleep(3 * time.Second)
	if atomic.LoadUint32(&exited) != 0 {
		panic("atomic.LoadUint32(&exited) != 0")
	}

	filesystem.CreationUnlock("bob", 4)
	time.Sleep(3 * time.Second)

	if atomic.LoadUint32(&exited) != 10 {
		panic("atomic.LoadUint32(&exited) != 10")
	}

	if filesystem.Open("bob", 4) != nil {
		panic("should be nil")
	}

	f1 := filesystem.Swap("bob", 4, 1, true)
	if f1 == nil {
		panic("f1 == nil")
	}
	if f1.refCount != 1 {
		panic("f1.refCount != 1")
	}

	f2 := filesystem.Open("bob", 4)
	if f2 == nil {
		panic("f2 == nil")
	}
	if f2.refCount != 2 {
		panic("f2.refCount != 2")
	}
	if f1.refCount != 2 {
		panic("f1.refCount != 2")
	}

	filesystem.Close(f1)
	if f2.refCount != 1 {
		panic("f2.refCount != 1")
	}
	if f1.refCount != 1 {
		panic("f1.refCount != 1")
	}

	filesystem.Close(f2)
	if f2.refCount != 0 {
		panic("f2.refCount != 0")
	}
	if f1.refCount != 0 {
		panic("f1.refCount != 0")
	}

	f3 := filesystem.Swap("bob", 4, 3, true)
	if f3 == nil {
		panic("f3 == nil")
	}
	if f3.refCount != 1 {
		panic("f3.refCount != 1")
	}
	if f3.id != 3 {
		panic("f3.id != 3")
	}

	filesystem.deleteLungo("bob", 4)

	if f3.deleted != 1 {
		panic("f3.deleted != 1")
	}

	select {
	case sl := <-gced:
		panic(fmt.Errorf("%v", sl))
	default:
	}

	filesystem.Close(f3)
	if f3.refCount != 0 {
		panic("f3.refCount != 0")
	}

	time.Sleep(1 * time.Second)
	select {
	case sl := <-gced:
		if sl[0].(uint32) != 3 {
			panic("sl[0].(uint32) != 3")
		}
		if sl[1].(string) != "bob" {
			panic("sl[1].(string) != bob")
		}

	default:
		panic("did not gc")
	}
}

var gced = make(chan []interface{}, 32)

func gc(id uint32, name string) {
	gced <- []interface{}{id, name}
	return
}
