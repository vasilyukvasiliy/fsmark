package fsmark

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"sync"
	"time"
)

func New(path string, duration time.Duration) FSMark {
	return FSMark{
		mode:     os.ModePerm,
		path:     path,
		mutex:    sync.Mutex{},
		duration: duration,
	}
}

type FSMark struct {
	path     string
	mode     os.FileMode
	mutex    sync.Mutex
	duration time.Duration
}

func (fsm *FSMark) BuildPath(key string) string {
	bytes := sha256.Sum256([]byte(key))
	key = hex.EncodeToString(bytes[:])

	return fsm.path + "/" + key[0:4] + "/" + key[4:8] + "/" + key[8:]
}

func (fsm *FSMark) Clear() (e error) {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()

	return os.RemoveAll(fsm.path)
}

func (fsm *FSMark) Create(key string) (e error) {
	path := fsm.BuildPath(key)
	e = fsm.Remove(key)
	if e != nil {
		return
	}

	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()

	return os.MkdirAll(path, fsm.mode)
}

func (fsm *FSMark) Exist(key string) (is bool) {
	return fsm.ExistUnixNano(key, 0)
}

func (fsm *FSMark) ExistUnix(key string, duration time.Duration) (is bool) {
	return fsm.ExistUnixNano(key, duration*time.Second)
}

func (fsm *FSMark) ExistUnixNano(key string, duration time.Duration) (is bool) {
	path := fsm.BuildPath(key)

	fsm.mutex.Lock()
	if oss, e := os.Stat(path); e == nil {
		is = (time.Now().UnixNano()-oss.ModTime().UnixNano()) <= int64(duration) || (duration == 0)
		fsm.mutex.Unlock()
		if !is {
			fsm.Remove(key)
		}
	} else {
		fsm.mutex.Unlock()
	}

	return
}

func (fsm *FSMark) Remove(key string) (e error) {
	key = fsm.BuildPath(key)
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()

	return os.RemoveAll(key)
}

func (fsm *FSMark) GCDemon(duration time.Duration) {
	for {
		fsm.GCDemonCustomDuration(duration, fsm.duration)
	}
}

func (fsm *FSMark) GCDemonCustomDuration(durationSleep time.Duration, durationGC time.Duration) {
	for {
		time.Sleep(durationSleep)
		fsm.GCUnixNano(durationGC)
	}
}

func (fsm *FSMark) GCUnix(duration time.Duration) {
	fsm.GCUnixNano(duration * time.Second)
}

func (fsm *FSMark) GCUnixNano(duration time.Duration) {
	//ToDo: Garbage Collection
}
