package fsmark

import (
	"os"
	"strconv"
	"testing"
	"time"
)

func TestLock(t *testing.T) {
	t.Parallel()

	l := New("./TestLock")
	key := "TestLock"
	e := l.Create(key)
	if e != nil {
		t.Fatal(e)
	}

	t.Log(l.Exist(key))
	t.Log(l.ExistUnix(key, time.Second))
	t.Log(l.Exist("Test"))
	t.Log(l.Exist(""))

	e = l.Clear()
	if e != nil {
		t.Error(e)
	}

	os.RemoveAll(l.Path)
}

func BenchmarkFSMark(b *testing.B) {
	l := New("./BenchmarkFSMark")

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		l.Create(strconv.Itoa(i))
	}
	b.StopTimer()

	os.RemoveAll(l.Path)
}
