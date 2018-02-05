package fsmark

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestLock(t *testing.T) {
	t.Parallel()

	tmp, e := ioutil.TempDir("", "TestLock")
	if e != nil {
		t.Fatal(e)
	}

	t.Log("TempDir", tmp)
	fsm := New(tmp, time.Hour)
	key := "TestLock"

	t.Log(fsm.BuildPath(key))

	e = fsm.Create(key)
	if e != nil {
		t.Fatal(e)
	}

	if !fsm.Exist(key) {
		t.Error("Marker not exist: " + key)
	}

	if !fsm.ExistUnix(key, 1) {
		t.Error("Marker not exist: " + key)
	}

	if !fsm.ExistUnix(key, time.Second) {
		t.Error("Marker not exist: " + key)
	}

	e = fsm.Remove(key)
	if e != nil {
		t.Fatal(e)
	}

	if fsm.Exist(key) {
		t.Error("Marker should not exist: " + key)
	}

	key = "Test"
	if fsm.Exist("Test") {
		t.Error("Marker should not exist: " + key)
	}

	key = ""
	if fsm.Exist(key) {
		t.Error("Marker should not exist: " + key)
	}

	for i := 0; i < 4096; i++ {
		fsm.Create(strconv.Itoa(i))
	}

	time.Sleep(1 * time.Second)

	fsm.GCUnixNano(1)

	e = fsm.Clear()
	if e != nil {
		t.Error(e)
	}

	os.RemoveAll(fsm.path)
}

func BenchmarkFSMark_Create(b *testing.B) {
	tmp, e := ioutil.TempDir("", "BenchmarkFSMark_Create")
	if e != nil {
		b.Fatal(e)
	}

	fsm := New(tmp, time.Hour)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fsm.Create(strconv.Itoa(i))
	}
	b.StopTimer()

	os.RemoveAll(fsm.path)
}

func BenchmarkFSMark_Exist(b *testing.B) {
	tmp, e := ioutil.TempDir("", "BenchmarkFSMark_Exist")
	if e != nil {
		b.Fatal(e)
	}

	fsm := New(tmp, time.Hour)

	for i := 0; i < b.N; i++ {
		fsm.Create(strconv.Itoa(i))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fsm.Exist(strconv.Itoa(i))
	}
	b.StopTimer()

	os.RemoveAll(fsm.path)
}

func BenchmarkFSMark_CreateExist(b *testing.B) {
	tmp, e := ioutil.TempDir("", "BenchmarkFSMark_CreateExist")
	if e != nil {
		b.Fatal(e)
	}

	fsm := New(tmp, time.Hour)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fsm.Create(strconv.Itoa(i))
		fsm.Exist(strconv.Itoa(i))
	}
	b.StopTimer()

	os.RemoveAll(fsm.path)
}
