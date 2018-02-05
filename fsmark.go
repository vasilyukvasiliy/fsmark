package fsmark

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func New(path string, duration time.Duration) FSMark {
	path, _ = filepath.Abs(path)
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
	bytes := sha1.Sum([]byte(key))
	key = hex.EncodeToString(bytes[:])

	return fsm.path + "/" + key[0:4] + "/" + key[4:]
}

func (fsm *FSMark) Clear() (e error) {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()

	return os.RemoveAll(fsm.path)
}

func (fsm *FSMark) Create(key string) (e error) {
	path := fsm.BuildPath(key)
	if _, err := os.Stat(path); err == nil {
		e = fsm.Remove(key)
		if e != nil {
			return
		}
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
		is = fsm.CheckExpire(oss.ModTime(), time.Now(), duration)
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
	fsm.GCDemonCustomDuration(duration, fsm.duration)
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

func (fsm *FSMark) CheckExpire(mod, now time.Time, duration time.Duration) bool {
	return now.UnixNano()-mod.UnixNano() <= int64(duration) || (duration == 0)
}

func (fsm *FSMark) GCUnixNano(duration time.Duration) {
	dropPath := ""
	filepath.Walk(fsm.path, func(path string, info os.FileInfo, err error) (e error) {
		if dropPath != path && dropPath != "" {
			fsm.mutex.Lock()
			os.RemoveAll(dropPath)
			fsm.mutex.Unlock()
		}

		if err != nil || fsm.path == path {
			return
		}

		fnLen := len(info.Name())
		if !info.IsDir() || (fnLen != 4 && fnLen != 36) {
			os.RemoveAll(path)
			return
		}

		if fnLen == 36 {
			if !fsm.CheckExpire(info.ModTime(), time.Now(), duration) {
				dropPath = path
			}
		}

		return
	})
}
