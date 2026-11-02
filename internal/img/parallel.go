package img

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
)

// ParallelOptions 控制批量并发。Jobs<=0 时使用 runtime.NumCPU()。
type ParallelOptions struct {
	Jobs     int
	Continue bool
}

const maxWorkers = 1024

// WorkerCount 返回实际 worker 数。jobs<=0 表示自动取逻辑 CPU 数。
func WorkerCount(jobs int) int {
	if jobs <= 0 {
		n := runtime.NumCPU()
		if n < 1 {
			return 1
		}
		return n
	}
	if jobs > maxWorkers {
		return maxWorkers
	}
	return jobs
}

// DescribeWorkers 用于日志。jobs<=0 表示自动。
func DescribeWorkers(jobs int) string {
	if jobs <= 0 {
		return fmt.Sprintf("%d (cpu)", WorkerCount(0))
	}
	return fmt.Sprintf("%d", WorkerCount(jobs))
}

// Parallel 以有限并发执行 n 个任务。Continue 为 false 时，首个错误后不再启动新任务。
func Parallel(n int, opt ParallelOptions, fn func(i int) error) error {
	if n <= 0 {
		return nil
	}
	workers := WorkerCount(opt.Jobs)
	if workers > n {
		workers = n
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	sem := make(chan struct{}, workers)
	var mu sync.Mutex
	var errs []error

loop:
	for i := 0; i < n; i++ {
		if !opt.Continue {
			select {
			case <-ctx.Done():
				break loop
			default:
			}
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()
			if !opt.Continue {
				select {
				case <-ctx.Done():
					return
				default:
				}
			}
			if err := fn(i); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				if !opt.Continue {
					cancel()
				}
			}
		}(i)
	}
	wg.Wait()
	return errors.Join(errs...)
}
