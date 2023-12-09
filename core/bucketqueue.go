package core

import (
	"runtime"
	"sync"
	"sync/atomic"
)

var (
	INVALIDATE = -1
)

type BucketQueue[T any, K any] struct {
	queueCapacity  uint32
	bucketCapacity uint32
	buckets        []*bucket[T]
	tagController  tagController
	tagsMap        sync.Map
	head           uint32
	tail           uint32
}

func NewBucketQueue[T any, K any](bucketCapacity uint32, queueCapacity uint32) *BucketQueue[T, K] {
	q := &BucketQueue[T, K]{
		queueCapacity:  queueCapacity,
		bucketCapacity: bucketCapacity,
		buckets:        make([]*bucket[T], queueCapacity),
		head:           0,
		tail:           0,
	}
	return q
}

func (q *BucketQueue[T, K]) Quantity() uint32 {
	tail := atomic.LoadUint32(&q.tail)
	head := atomic.LoadUint32(&q.head)
	return (tail - head + q.queueCapacity) % q.queueCapacity
}

func (q *BucketQueue[T, K]) Empty() bool {
	tail := atomic.LoadUint32(&q.tail)
	head := atomic.LoadUint32(&q.head)
	return head == tail
}

func (q *BucketQueue[T, K]) Full() bool {
	tail := atomic.LoadUint32(&q.tail)
	head := atomic.LoadUint32(&q.tail)
	return head == (tail+1)%q.queueCapacity
}

func (q *BucketQueue[T, K]) Capacity() uint32 {
	return atomic.LoadUint32(&q.queueCapacity)
}

func (q *BucketQueue[T, K]) Put(val *T, tag K) {
	for {
		var ptr int
		i, ok := q.tagsMap.Load(tag)
		if !ok || i.(int) == INVALIDATE {

			if q.tagController.ClaimTag(tag) {
				for !q.addBucket(tag) {
				}
				q.tagController.ReleaseTag(tag)
			}

			runtime.Gosched()

			continue

		}

		ptr = i.(int)

		ok, info := q.buckets[ptr].Put(val)
		if ok {
			return
		}

		if info == FAILED || !q.tagController.ClaimTag(tag) {
			runtime.Gosched()
			continue
		}

		i, ok = q.tagsMap.Load(tag)
		ptr2 := i.(int)

		if ptr2 != INVALIDATE && ptr == ptr2 {
			q.tagsMap.Store(tag, INVALIDATE)
		}
		q.tagController.ReleaseTag(tag)

	}
}

func (q *BucketQueue[T, K]) Gets() (vals *[]*T, ok bool) {
	for {
		if q.Empty() {
			return nil, false
		}
		old_head := atomic.LoadUint32(&q.head)
		new_head := (old_head + 1) % q.Capacity()
		q.buckets[old_head].Close()

		if q.buckets[old_head].IsClose() {
			ok, vals = q.buckets[old_head].Gets()
			if ok {
				if atomic.CompareAndSwapUint32(&q.head, old_head, new_head) {
					return vals, true
				}
			}
		}
	}
}
func (q *BucketQueue[T, K]) Get() (val *T, ok bool) {
	for {
		if q.Empty() {
			return nil, false
		}
		old_head := atomic.LoadUint32(&q.head)
		new_head := (old_head + 1) % q.Capacity()
		q.buckets[old_head].Close()
		if q.buckets[old_head].IsClose() {
			if q.buckets[old_head].Invalid() {
				if !atomic.CompareAndSwapUint32(&q.head, old_head, new_head) {
					return nil, false
				}
			} else {
				ok, val := q.buckets[old_head].Get()
				if ok {
					return val, true
				}
			}

		}
	}
}
func (q *BucketQueue[T, K]) addBucket(tag K) (ok bool) {

	if q.Full() {
		runtime.Gosched()
		return false
	}

	old_tail := atomic.LoadUint32(&q.tail)
	new_tail := (old_tail + 1) % q.queueCapacity
	if !atomic.CompareAndSwapUint32(&q.tail, old_tail, new_tail) {
		runtime.Gosched()
		return false
	}
	q.buckets[old_tail] = NewBucket[T](q.bucketCapacity)
	q.tagsMap.Store(tag, int(old_tail))
	return true
}
