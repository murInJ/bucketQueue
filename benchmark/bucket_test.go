package benchmark

import (
	"bucketQueue/core"
	"sync"
	"testing"
)

func BenchmarkBucket_Put(b *testing.B) {
	// 初始化 Bucket 和其他必要的数据结构
	bucket := core.NewBucket[int](uint32(1e5))
	// 使用 b.N 次 Put 操作进行性能测试
	for i := 0; i < b.N; i++ {
		// 执行 Put 操作
		bucket.Put(&i)
	}
}

func BenchmarkBucket_Get(b *testing.B) {
	// 初始化 Bucket 和其他必要的数据结构
	bucket := core.NewBucket[int](uint32(1e5))
	for i := 0; i < 1e5; i++ {
		bucket.Put(&i)
	}
	// 使用 b.N 次 Gets 操作进行性能测试
	for i := 0; i < b.N; i++ {
		// 执行 Gets 操作
		bucket.Get()
	}
}

func BenchmarkBucket_Gets(b *testing.B) {
	// 初始化 Bucket 和其他必要的数据结构
	bucket := core.NewBucket[int](uint32(1e5))
	for i := 0; i < 1e5; i++ {
		bucket.Put(&i)
	}
	// 使用 b.N 次 Gets 操作进行性能测试
	for i := 0; i < b.N; i++ {
		// 执行 Gets 操作
		bucket.Gets()
	}
}

// 可以添加其他 Benchmark 函数来测试其他操作

func BenchmarkConcurrentBucketPut(b *testing.B) {
	// 初始化 Bucket 和其他必要的数据结构
	bucket := core.NewBucket[int](uint32(1e5))
	var wg sync.WaitGroup
	b.ResetTimer()

	// 使用多个 goroutine 进行并发 Put 操作
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 执行 Put 操作
			bucket.Put(&i)
		}()
	}

	wg.Wait()
}

func BenchmarkConcurrentBucketGet(b *testing.B) {
	// 初始化 Bucket 和其他必要的数据结构
	bucket := core.NewBucket[int](uint32(1e5))
	for i := 0; i < 1e5; i++ {
		bucket.Put(&i)
	}
	var wg sync.WaitGroup
	b.ResetTimer()

	// 使用多个 goroutine 进行并发 Gets 操作
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 执行 Gets 操作
			bucket.Get()
		}()
	}

	wg.Wait()
}

func BenchmarkConcurrentBucketGets(b *testing.B) {
	// 初始化 Bucket 和其他必要的数据结构
	bucket := core.NewBucket[int](uint32(1e5))
	for i := 0; i < 1e5; i++ {
		bucket.Put(&i)
	}
	var wg sync.WaitGroup
	b.ResetTimer()

	// 使用多个 goroutine 进行并发 Gets 操作
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 执行 Gets 操作
			bucket.Gets()
		}()
	}

	wg.Wait()
}
