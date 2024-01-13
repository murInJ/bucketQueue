package main

import (
	"runtime"
	"sync"
)

type BucketQueue[K comparable] struct {
	queueCapacity  uint32
	bucketCapacity uint32
	split          uint32
	queue          *Queue
	buckets        sync.Map
}

func NewBucketQueue[K comparable](QueueCap uint32, BucketCap uint32, split uint32) *BucketQueue[K] {
	q := new(BucketQueue[K])
	q.queueCapacity = minQuantity(QueueCap)
	q.bucketCapacity = minQuantity(BucketCap)
	q.split = minQuantity(split)
	q.queue = NewQueue(q.queueCapacity, 1)

	return q
}

func (q *BucketQueue[K]) Put(key K, val any) (ok bool) {
	newBucket := NewQueue(q.bucketCapacity, q.split)
	ok, splitNum, _ := newBucket.Put(val)
	if !ok {
		return false
	}

	actual, loaded := q.buckets.LoadOrStore(key, newBucket)
	bucket := actual.(*Queue)

	if loaded {
		ok, splitNum, _ = bucket.Put(val)
		if !ok {
			return false
		}
	} else {
		splitNum = 1
	}
	keys := make([]any, splitNum)
	for i := uint32(0); i < splitNum; i++ {
		keys[i] = key
	}

	for splitNum > 0 {
		puts, _, _ := q.queue.Puts(keys)
		if puts == splitNum {
			break
		}
		runtime.Gosched()
	}

	return true
}

func (q *BucketQueue[K]) Puts(key K, vals []any) (ok bool) {
	newBucket := NewQueue(q.bucketCapacity, q.split)
	puts, splitNum, _ := newBucket.Puts(vals)
	if puts == 0 {
		return false
	}

	actual, loaded := q.buckets.LoadOrStore(key, newBucket)
	bucket := *actual.(*Queue)
	if loaded {
		puts, splitNum, _ = bucket.Puts(vals)
		if puts == 0 {
			return false
		}
	} else {
		splitNum += 1
	}
	keys := make([]any, splitNum)
	for i := uint32(0); i < splitNum; i++ {
		keys[i] = key
	}

	for {
		puts, _, _ = q.queue.Puts(keys)
		if puts != 0 {
			break
		}
		runtime.Gosched()
	}

	return true
}

func (q *BucketQueue[K]) Get() (values []any, ok bool) {

	val, ok, _ := q.queue.Get()
	if !ok {
		return nil, false
	}
	key := val.(K)

	val, ok = q.buckets.Load(key)
	if !ok {
		return nil, false
	}

	bucket := val.(*Queue)
	values = make([]any, q.split)
	gets, _ := bucket.Gets(values)

	if gets == 0 {
		return nil, false
	}
	return values[:gets], true
}

func (q *BucketQueue[K]) Quantity() uint32 {
	return q.queue.Quantity()
}
