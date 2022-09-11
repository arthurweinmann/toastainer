package atomicpath

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/toastate/toastcloud/internal/utils"
)

type File struct {
	name     string
	refCount int64
	modTime  int64
	id       uint32
	deleted  uint32
	hash     uint16

	next *File
}

type FileSystem struct {
	gcStart *File
	gcLast  *File
	gcMu    sync.Mutex

	creationLock   []map[string]*sync.WaitGroup
	creationLockMu []*sync.Mutex

	paths   []map[string]*File
	pathsMu []*sync.RWMutex

	modulo uint16

	gc func(id uint32, name string)
}

func NewFileSystem(count int, gc func(id uint32, name string)) *FileSystem {
	if count < 2 {
		panic("count must be >= 2")
	}
	if !utils.IsPowerOf2(count) {
		panic("count must be a power of 2")
	}

	h := &FileSystem{}

	h.paths = make([]map[string]*File, count)
	h.pathsMu = make([]*sync.RWMutex, count)
	h.creationLock = make([]map[string]*sync.WaitGroup, count)
	h.creationLockMu = make([]*sync.Mutex, count)

	for i := 0; i < count; i++ {
		h.paths[i] = make(map[string]*File)
		h.pathsMu[i] = &sync.RWMutex{}
		h.creationLock[i] = make(map[string]*sync.WaitGroup)
		h.creationLockMu[i] = &sync.Mutex{}
	}

	h.modulo = uint16(count - 1)

	h.gc = gc

	return h
}

// Open returns nil if the file does not exist
func (fs *FileSystem) Open(name string, hash uint16) (f *File) {
	modulo := hash & fs.modulo

	fs.pathsMu[modulo].RLock()
	f, _ = fs.paths[modulo][name]
	if f != nil {
		atomic.AddInt64(&f.refCount, 1)
	}
	fs.pathsMu[modulo].RUnlock()

	if f != nil {
		atomic.StoreInt64(&f.modTime, time.Now().Unix())
	}

	return
}

func (fs *FileSystem) Close(f *File) {
	if atomic.AddInt64(&f.refCount, -1) == 0 {
		if atomic.LoadUint32(&f.deleted) == 1 {
			go fs.gc(f.id, f.name)
		}
	}
}

// Swap returns the new file
func (fs *FileSystem) Swap(name string, hash uint16, newid uint32, openNew bool) (f *File) {
	modulo := hash & fs.modulo

	f = &File{id: newid, name: name, modTime: time.Now().Unix(), hash: hash}
	if openNew {
		f.refCount = 1
	}

	fs.pathsMu[modulo].Lock()
	swapped, _ := fs.paths[modulo][name]
	fs.paths[modulo][name] = f
	fs.pathsMu[modulo].Unlock()

	if swapped != nil {
		atomic.StoreUint32(&swapped.deleted, 1)
		if atomic.LoadInt64(&swapped.refCount) == 0 {
			go fs.gc(swapped.id, swapped.name)
		}
	}

	go fs.pushGC(f)

	return
}

func (fs *FileSystem) deleteLungo(name string, hash uint16) {
	modulo := hash & fs.modulo

	fs.pathsMu[modulo].Lock()
	deleted, _ := fs.paths[modulo][name]
	if deleted != nil {
		delete(fs.paths[modulo], name)
	}
	fs.pathsMu[modulo].Unlock()

	if deleted != nil {
		atomic.StoreUint32(&deleted.deleted, 1)
		if atomic.LoadInt64(&deleted.refCount) == 0 {
			fs.gc(deleted.id, deleted.name)
		}
	}
}

func (fs *FileSystem) deleteIfMod(name string, hash uint16, mod int64) bool {
	modulo := hash & fs.modulo

	fs.pathsMu[modulo].Lock()
	deleted, _ := fs.paths[modulo][name]
	if deleted != nil {
		if deleted.modTime != mod {
			deleted = nil
		} else {
			delete(fs.paths[modulo], name)
		}
	}
	fs.pathsMu[modulo].Unlock()

	if deleted != nil {
		atomic.StoreUint32(&deleted.deleted, 1)
		if atomic.LoadInt64(&deleted.refCount) == 0 {
			fs.gc(deleted.id, deleted.name)
		}

		return true
	}

	return false
}

// CreationLock - if wait is true and the lock was already taken, then it waits for the other thread to finish creating but it doesn't retry to lock after that
func (fs *FileSystem) CreationLock(name string, hash uint16, wait bool) bool {
	modulo := hash & fs.modulo

	w := &sync.WaitGroup{}
	w.Add(1)

	fs.creationLockMu[modulo].Lock()
	wg, _ := fs.creationLock[modulo][name]
	if wg == nil {
		fs.creationLock[modulo][name] = w
	}
	fs.creationLockMu[modulo].Unlock()

	if wg != nil {
		if wait {
			wg.Wait()
		}

		return false
	}

	return true
}

func (fs *FileSystem) CreationUnlock(name string, hash uint16) {
	modulo := hash & fs.modulo

	fs.creationLockMu[modulo].Lock()
	wg, _ := fs.creationLock[modulo][name]
	delete(fs.creationLock[modulo], name)
	fs.creationLockMu[modulo].Unlock()

	if wg == nil {
		panic("unlocked not locked file")
	}

	wg.Done()

	return
}

func (fs *FileSystem) pushGC(f *File) {
	fs.gcMu.Lock()
	if fs.gcStart == nil {
		fs.gcStart = f
		fs.gcLast = f
	} else {
		fs.gcLast.next = f
		fs.gcLast = f
	}
	fs.gcMu.Unlock()
}

// GC must be called only once at a time
func (fs *FileSystem) GC(ttlSec int64, sleepBetweenDeletes time.Duration) {
	fs.gcMu.Lock()
	start := fs.gcStart
	last := fs.gcLast
	fs.gcMu.Unlock()

	t := time.Now().Unix() - ttlSec

	if start != nil {
		if start == last {
			return
		}

		prev := start
		subj := start

		var mod int64

		for {
			mod = atomic.LoadInt64(&subj.modTime)

			if t > mod {
				if fs.deleteIfMod(subj.name, subj.hash, mod) {
					if subj == start {
						start = subj.next

						fs.gcMu.Lock()
						fs.gcStart = start
						fs.gcMu.Unlock()

						prev = start
						subj = start

						continue
					} else if subj == last {
						prev.next = nil

						fs.gcMu.Lock()
						fs.gcLast = prev
						fs.gcMu.Unlock()

						break
					}

					if subj.next == last {
						fs.gcMu.Lock()
						prev.next = subj.next
						subj = prev
						fs.gcMu.Unlock()
					} else {
						prev.next = subj.next
						subj = prev
					}
				}
			}

			if subj.next == nil {
				break
			}

			prev = subj
			subj = subj.next
		}
	}
}

func (h *FileSystem) Refresh(ttlSec int64, sleepBetweenDeletes time.Duration, cb func(id uint32, name string)) {
	return
}

func (f *File) GetID() uint32 {
	return f.id
}

func (f *File) GetIDStr() string {
	return strconv.Itoa(int(f.id))
}

func (f *File) ModTime() int64 {
	return atomic.LoadInt64(&f.modTime)
}

func (f *File) IsZombie() bool {
	if atomic.LoadInt64(&f.refCount) == 0 {
		if atomic.LoadUint32(&f.deleted) == 1 {
			return true
		}
	}

	return false
}
