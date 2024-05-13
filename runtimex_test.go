// Copyright 2024 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runtimex_test

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modern-go/reflect2"

	"github.com/cloudwego/runtimex"
)

//go:noinline
func testStackFunction(n int) int {
	var stack [1024 * 4]int64
	for i := 0; i < len(stack); i++ {
		stack[i] = int64(n + i)
	}
	return int(stack[len(stack)/2])
}

func assert(t *testing.T, cond bool, args ...interface{}) {
	t.Helper()
	if cond {
		return
	}
	if len(args) > 0 {
		t.Fatal(args...)
	}
	t.Fatal("assertion failed")
}

func TestReflect2NotFound(t *testing.T) {
	notFound, _ := reflect2.TypeByName("runtime.xxx123").(reflect2.StructType)
	assert(t, notFound == nil)

	gType := reflect2.TypeByName("runtime.g").(reflect2.StructType)
	assert(t, gType != nil, gType)
	mField := gType.FieldByName("xxx123")
	assert(t, mField == nil)
}

func TestRuntimeStatus(t *testing.T) {
	goroutines := 8
	pnum := 4
	oldpnum := runtime.GOMAXPROCS(pnum)
	defer runtime.GOMAXPROCS(oldpnum)

	var (
		stop int32
		wg   sync.WaitGroup
		gm   sync.Map
		pm   sync.Map
	)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 1; atomic.LoadInt32(&stop) == 0; j++ {
				testStackFunction(j % 102400)

				if j%1024 == 0 {
					gid, err := runtimex.GID()
					assert(t, err == nil)
					pid, err := runtimex.PID()
					assert(t, err == nil)
					t.Logf("gid=%d,pid=%d", gid, pid)

					gm.Store(gid, true)
					pm.Store(pid, true)
				}
			}
		}(i)
	}
	time.Sleep(time.Second)
	atomic.StoreInt32(&stop, 1)
	wg.Wait()
	gcount := 0
	gm.Range(func(key, value interface{}) bool {
		gcount++
		return true
	})
	pcount := 0
	pm.Range(func(key, value interface{}) bool {
		pcount++
		return true
	})
	assert(t, gcount == goroutines, gcount, goroutines)
	assert(t, pcount == pnum, pcount, pnum)
}

func BenchmarkGID(b *testing.B) {
	b.ReportAllocs()
	// 0 allocs/op
	for i := 0; i < b.N; i++ {
		id, err := runtimex.GID()
		if id < 0 || err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPID(b *testing.B) {
	b.ReportAllocs()
	// 0 allocs/op
	for i := 0; i < b.N; i++ {
		id, err := runtimex.PID()
		if id < 0 || err != nil {
			b.Fatal(err)
		}
	}
}
