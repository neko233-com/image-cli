package img

import (
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerCountAuto(t *testing.T) {
	if WorkerCount(0) != runtime.NumCPU() && runtime.NumCPU() > 0 {
		t.Fatalf("auto=%d cpu=%d", WorkerCount(0), runtime.NumCPU())
	}
	if WorkerCount(4) != 4 {
		t.Fatal(WorkerCount(4))
	}
	if WorkerCount(100000) != maxWorkers {
		t.Fatal(WorkerCount(100000))
	}
}

func TestParallelRespectsJobs(t *testing.T) {
	var cur, max atomic.Int32
	const jobs = 3
	err := Parallel(12, ParallelOptions{Jobs: jobs}, func(int) error {
		c := cur.Add(1)
		for {
			m := max.Load()
			if c <= m || max.CompareAndSwap(m, c) {
				break
			}
		}
		time.Sleep(15 * time.Millisecond)
		cur.Add(-1)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if max.Load() > jobs {
		t.Fatalf("inflight %d > %d", max.Load(), jobs)
	}
	if max.Load() < 2 {
		t.Fatalf("expected parallelism, inflight %d", max.Load())
	}
}
