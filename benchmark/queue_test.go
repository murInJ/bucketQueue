package benchmark

import (
	"bucketQueue/core"
	"sync"
	"testing"
)

func TestBucketQueue(t *testing.T) {
	N := 90
	arr := make([]int, 0)
	q := core.NewBucketQueue[int, int](uint32(3), uint32(100))
	for i := 0; i < N; i++ {
		var a int
		a = i % 5
		q.Put(&a, a%5)
	}
	for !q.Empty() {
		a, ok := q.Get()
		if ok {
			arr = append(arr, *a)
		}
		b, ok := q.Gets()
		if ok {
			for _, v := range *b {
				arr = append(arr, *v)
			}

		}
	}

	for _, v := range arr {
		t.Logf("%d ", v)
	}
}

func BenchmarkBucketQueue_Put(b *testing.B) {
	// 初始化 BucketQueue 和其他必要的数据结构
	q := core.NewBucketQueue[int, int](uint32(100), uint32(1000))
	// 使用 b.N 次 Put 操作进行性能测试
	for i := 0; i < b.N; i++ {
		// 执行 Put 操作
		var a int
		a = i
		q.Put(&a, a%50)
	}
}

func BenchmarkBucketQueue_Get(b *testing.B) {
	// 初始化 BucketQueue 和其他必要的数据结构
	q := core.NewBucketQueue[int, int](uint32(100), uint32(1000))
	for i := 0; i < 1e5; i++ {
		var a int
		a = i
		q.Put(&a, a%50)
	}
	// 使用 b.N 次 Gets 操作进行性能测试
	for i := 0; i < b.N; i++ {
		// 执行 Gets 操作
		q.Get()
	}
}

func BenchmarkBucketQueue_Gets(b *testing.B) {
	// 初始化 BucketQueue 和其他必要的数据结构
	q := core.NewBucketQueue[int, int](uint32(100), uint32(1000))
	for i := 0; i < 1e5; i++ {
		var a int
		a = i
		q.Put(&a, a%50)
	}
	// 使用 b.N 次 Gets 操作进行性能测试
	for i := 0; i < b.N; i++ {
		// 执行 Gets 操作
		q.Gets()
	}
}

// 可以添加其他 Benchmark 函数来测试其他操作

func BenchmarkConcurrentQueuePut(b *testing.B) {
	// 初始化 BucketQueue 和其他必要的数据结构
	q := core.NewBucketQueue[int, int](uint32(100), uint32(1000))

	var wg sync.WaitGroup
	b.ResetTimer()
	// 使用多个 goroutine 进行并发 Put 操作
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 执行 Put 操作
			var a int
			a = i
			q.Put(&a, a%50)
		}()
	}

	wg.Wait()
}

func BenchmarkConcurrentQueueGet(b *testing.B) {
	// 初始化 BucketQueue 和其他必要的数据结构
	q := core.NewBucketQueue[int, int](uint32(100), uint32(1000))
	for i := 0; i < 1e5; i++ {
		var a int
		a = i
		q.Put(&a, a%50)
	}
	var wg sync.WaitGroup
	b.ResetTimer()
	// 使用多个 goroutine 进行并发 Gets 操作
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 执行 Gets 操作
			q.Get()
		}()
	}

	wg.Wait()
}

func BenchmarkConcurrentQueueGets(b *testing.B) {
	// 初始化 BucketQueue 和其他必要的数据结构
	q := core.NewBucketQueue[int, int](uint32(100), uint32(1000))
	for i := 0; i < 1e5; i++ {
		var a int
		a = i
		q.Put(&a, a%50)
	}
	var wg sync.WaitGroup
	b.ResetTimer()
	// 使用多个 goroutine 进行并发 Gets 操作
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 执行 Gets 操作
			q.Gets()
		}()
	}

	wg.Wait()
}
