// Copyright (c) 2018 Vasiliy Vasilyuk <vasilyukvasiliy@gmail.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fsmark

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

func New(path string, mode os.FileMode, duration time.Duration, durationGC time.Duration) *FSMark {
	path, _ = filepath.Abs(path)
	return &FSMark{
		mode:       mode,
		path:       path,
		duration:   duration,
		durationGC: durationGC,
	}
}

type FSMark struct {
	path       string
	mode       os.FileMode
	mutex      sync.RWMutex
	duration   time.Duration
	durationGC time.Duration
}

func (fsm *FSMark) BuildPath(key string) string {
	bytes := sha256.Sum256([]byte(key))
	key = hex.EncodeToString(bytes[:])

	return fsm.path + "/" + key[0:4] + "/" + key[4:]
}

func (fsm *FSMark) Clear() (e error) {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()

	return os.RemoveAll(fsm.path)
}

func (fsm *FSMark) Create(key string) (e error) {
	return fsm.CreateUnixNano(key, fsm.duration)
}

func (fsm *FSMark) CreateUnix(key string, duration time.Duration) (e error) {
	return fsm.CreateUnixNano(key, fsm.duration*time.Second)
}

func (fsm *FSMark) CreateUnixNano(key string, duration time.Duration) (e error) {
	if fsm.Exist(key) {
		fsm.Delete(key)
	}

	path := fsm.BuildPath(key)
	dir := filepath.Dir(path)

	if duration.Nanoseconds() > time.Millisecond.Nanoseconds() {
		fsm.mutex.Lock()
		defer fsm.mutex.Unlock()

		os.MkdirAll(dir, fsm.mode)
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, fsm.mode)
		if err != nil {
			return err
		}
		f.Write([]byte(strconv.FormatInt(timeNowUTCUnixNano()+int64(duration), 10)))
		f.Close()
	}

	return
}

func (fsm *FSMark) Exist(key string) (is bool) {
	path := fsm.BuildPath(key)

	fsm.mutex.RLock()
	exist := false
	defer func() {
		fsm.mutex.RUnlock()
		if !is && exist {
			fsm.Delete(key)
		}
	}()

	if fileInfo, e := os.Stat(path); e == nil {
		exist = true
		if fileInfo.IsDir() {
			return
		}

		bytes, e := ioutil.ReadFile(path)
		if e != nil {
			return
		}

		timeUtcUnixNano, e := strconv.ParseInt(string(bytes), 10, 64)
		if e != nil {
			return
		}

		if timeUtcUnixNano < timeNowUTCUnixNano() {
			return
		}

		is = true
	}

	return
}

func (fsm *FSMark) Delete(key string) (e error) {
	key = fsm.BuildPath(key)
	return fsm.remove(key)
}

func (fsm *FSMark) remove(path string) (e error) {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()

	return os.RemoveAll(path)
}

func (fsm *FSMark) GCDemon() {
	fsm.GCDemonCustomDuration(fsm.durationGC)
}

func (fsm *FSMark) GCDemonCustomDuration(durationSleep time.Duration) {
	for {
		time.Sleep(durationSleep)
		fsm.GC()
	}
}

func (fsm *FSMark) GC() {
	filepath.Walk(fsm.path, func(path string, info os.FileInfo, err error) (e error) {
		if err != nil || fsm.path == path {
			return
		}

		length := len(info.Name())
		if (length != 4 && info.IsDir()) || (length == 4 && !info.IsDir()) {
			fsm.mutex.Lock()
			os.RemoveAll(path)
			fsm.mutex.Unlock()
			return
		}

		if length == 4 {
			i, err := ioutil.ReadDir(path)
			if err != nil {
				return err
			}

			if len(i) == 0 {
				fsm.remove(path)
			}
		} else if length == 36 {
			bytes, e := ioutil.ReadFile(path)
			if e != nil {
				return fsm.remove(path)
			}

			timeUtcUnixNano, e := strconv.ParseInt(string(bytes), 10, 64)
			if e != nil {
				return fsm.remove(path)
			}

			if timeUtcUnixNano < timeNowUTCUnixNano() {
				return fsm.remove(path)
			}
		}

		return
	})
}

func timeNowUTCUnixNano() int64 {
	return time.Now().UTC().UnixNano()
}
