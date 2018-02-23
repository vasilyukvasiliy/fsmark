// Copyright 2018 Vasiliy Vasilyuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fsmark

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"
)

func testNew() (*FSMark, error) {
	tmp, e := ioutil.TempDir("", "FSMark__")
	if e != nil {
		return nil, e
	}

	return New(tmp, os.ModePerm, time.Hour, time.Minute), nil
}

func TestGC(t *testing.T) {
	t.Parallel()

	fsm, e := testNew()
	if e != nil {
		t.Fatal(e)
	}

	key := "TestGC"
	fsm.CreateUnixNano(key, time.Hour)
	for i := 0; i < 16; i++ {
		fsm.CreateUnixNano(strconv.Itoa(i), time.Millisecond)
	}

	time.Sleep(time.Millisecond * 2)

	fsm.GC()

	for i := 0; i < 16; i++ {
		if fsm.Exist(strconv.Itoa(i)) {
			t.Fatal("Marker should not exist: ", i)
		}
	}

	if !fsm.Exist(key) {
		t.Error("Marker not exist: " + key)
	}

	os.RemoveAll(fsm.path)
}

func TestFSMark(t *testing.T) {
	t.Parallel()

	fsm, e := testNew()
	if e != nil {
		t.Fatal(e)
	}

	key := "TestLock"
	t.Log(fsm.BuildPath(key))

	e = fsm.Create(key)
	if e != nil {
		t.Fatal(e)
	}

	if !fsm.Exist(key) {
		t.Error("Marker not exist: " + key)
	}

	if !fsm.Exist(key) {
		t.Error("Marker not exist: " + key)
	}

	e = fsm.CreateUnixNano(key, 1)
	if e != nil {
		t.Fatal(e)
	}

	if fsm.Exist(key) {
		t.Error("Marker should not exist: " + key)
	}

	e = fsm.Delete(key)
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
		fsm.CreateUnixNano(strconv.Itoa(i), time.Second*10)
	}

	fsm.GC()

	e = fsm.Clear()
	if e != nil {
		t.Error(e)
	}

	os.RemoveAll(fsm.path)
}

func BenchmarkFSMark_Create(b *testing.B) {
	fsm, e := testNew()
	if e != nil {
		b.Fatal(e)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fsm.Create(strconv.Itoa(i))
	}
	b.StopTimer()

	os.RemoveAll(fsm.path)
}

func BenchmarkFSMark_Exist(b *testing.B) {
	fsm, e := testNew()
	if e != nil {
		b.Fatal(e)
	}

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
	fsm, e := testNew()
	if e != nil {
		b.Fatal(e)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fsm.Create(strconv.Itoa(i))
		fsm.Exist(strconv.Itoa(i))
	}
	b.StopTimer()

	os.RemoveAll(fsm.path)
}
