package fsmark

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"sync"
	"time"
)

func New(s string) FSMark {
	return FSMark{
		FileMode: os.ModePerm,
		Path:     s,
		mutex:    sync.Mutex{},
	}
}

type FSMark struct {
	Path     string
	FileMode os.FileMode
	mutex    sync.Mutex
}

func (f *FSMark) BuildPath(s string) string {
	bytes := sha256.Sum256([]byte(s))
	s = hex.EncodeToString(bytes[:])

	return f.Path + "/" + s[0:4] + "/" + s[4:8] + "/" + s[8:]
}

func (f *FSMark) Clear() (e error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return os.RemoveAll(f.Path)
}

func (f *FSMark) Create(s string) (e error) {
	sb := f.BuildPath(s)
	e = f.Remove(s)
	if e != nil {
		return
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	return os.MkdirAll(sb, f.FileMode)
}

func (f *FSMark) Exist(s string) (is bool) {
	s = f.BuildPath(s)

	f.mutex.Lock()
	defer f.mutex.Unlock()

	if _, e := os.Stat(s); e == nil {
		is = true
	}

	return
}

func (f *FSMark) ExistUnix(s string, i time.Duration) (is bool) {
	sb := f.BuildPath(s)

	f.mutex.Lock()
	if oss, e := os.Stat(sb); e == nil {
		is = (time.Now().UnixNano() - oss.ModTime().UnixNano()) <= int64(i)
		f.mutex.Unlock()
		if !is {
			f.Remove(s)
		}
	}

	return
}

func (f *FSMark) Remove(s string) (e error) {
	s = f.BuildPath(s)
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return os.RemoveAll(s)
}
