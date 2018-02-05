package fsmark

import (
	"os"
	"strconv"
	"testing"
	"time"
)

func TestLock(t *testing.T) {
	t.Parallel()

	l := New("./tmp/TestLock", time.Hour)
	key := "TestLock"

	t.Log(l.BuildPath(key))

	e := l.Create(key)
	if e != nil {
		t.Fatal(e)
	}

	if !l.Exist(key) {
		t.Error("Marker not exist: " + key)
	}

	if !l.ExistUnix(key, 1) {
		t.Error("Marker not exist: " + key)
	}

	if !l.ExistUnix(key, time.Second) {
		t.Error("Marker not exist: " + key)
	}

	e = l.Remove(key)
	if e != nil {
		t.Fatal(e)
	}

	if l.Exist(key) {
		t.Error("Marker should not exist: " + key)
	}

	key = "Test"
	if l.Exist("Test") {
		t.Error("Marker should not exist: " + key)
	}

	key = ""
	if l.Exist(key) {
		t.Error("Marker should not exist: " + key)
	}

	e = l.Clear()
	if e != nil {
		t.Error(e)
	}

	os.RemoveAll(l.path)
}

func BenchmarkFSMark_Create(b *testing.B) {
	l := New("./tmp/BenchmarkFSMark_Create", time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Create(strconv.Itoa(i))
	}
	b.StopTimer()

	os.RemoveAll(l.path)
}

func BenchmarkFSMark_Exist(b *testing.B) {
	l := New("./tmp/BenchmarkFSMark_Exist", time.Hour)

	for i := 0; i < b.N; i++ {
		l.Create(strconv.Itoa(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Exist(strconv.Itoa(i))
	}
	b.StopTimer()

	os.RemoveAll(l.path)
}

func BenchmarkFSMark_CreateExist(b *testing.B) {
	l := New("./tmp/BenchmarkFSMark_CreateExist", time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Create(strconv.Itoa(i))
		l.Exist(strconv.Itoa(i))
	}
	b.StopTimer()

	os.RemoveAll(l.path)
}
