package core

import (
	"runtime"
	"sync/atomic"
)

var (
	NORMAL = uint32(0)
	CLOSE  = uint32(1)
	FAILED = uint32(2)
)

type bucket[T any] struct {
	capacity uint32
	tail     uint32
	close    uint32
	op_num   uint32
	getPtr   uint32
	invalid  bool
	cache    []*T
}

func NewBucket[T any](capacity uint32) *bucket[T] {
	b := &bucket[T]{
		capacity: capacity,
		cache:    make([]*T, capacity),
		tail:     0,
		close:    0,
		op_num:   0,
		getPtr:   0,
		invalid:  false,
	}

	return b
}

func (b *bucket[T]) Quantity() uint32 {
	return atomic.LoadUint32(&b.tail)
}

func (b *bucket[T]) Capacity() uint32 {
	return atomic.LoadUint32(&b.capacity)
}

func (b *bucket[T]) Invalid() bool {
	return b.invalid
}

func (b *bucket[T]) Close() {
	for {
		if atomic.LoadUint32(&b.close) == 1 {
			return
		}

		if atomic.CompareAndSwapUint32(&b.close, 0, 1) {
			return
		}

		runtime.Gosched()

	}
}

func (b *bucket[T]) Put(val *T) (ok bool, info uint32) {
	if atomic.LoadUint32(&b.close) == 1 {
		return false, CLOSE
	}

	old_tail := atomic.LoadUint32(&b.tail)
	new_tail := old_tail + 1
	if old_tail >= b.capacity {
		b.Close()
		return false, CLOSE
	}
	atomic.AddUint32(&b.op_num, 1)

	if !atomic.CompareAndSwapUint32(&b.tail, old_tail, new_tail) {
		runtime.Gosched()
		atomic.AddUint32(&b.op_num, ^uint32(1)+1)
		return false, FAILED
	}
	b.cache[old_tail] = val
	atomic.AddUint32(&b.op_num, ^uint32(1)+1)
	return true, NORMAL
}

func (b *bucket[T]) Gets() (ok bool, cache *[]*T) {
	if !b.IsClose() || b.Invalid() {
		return false, nil
	}
	ptr := atomic.LoadUint32(&b.getPtr)
	if !atomic.CompareAndSwapUint32(&b.getPtr, ptr, b.Capacity()) {
		return false, nil
	}
	b.invalid = true
	s := b.cache[ptr:b.tail]
	return true, &s
}

func (b *bucket[T]) Get() (ok bool, val *T) {
	if !b.IsClose() || b.Invalid() {
		return false, nil
	}
	ptr := atomic.LoadUint32(&b.getPtr)
	next_ptr := ptr + 1

	if ptr >= b.Quantity() {
		b.invalid = true
		return false, nil
	}
	if !atomic.CompareAndSwapUint32(&b.getPtr, ptr, next_ptr) {
		return false, nil
	}
	return true, b.cache[ptr]
}

func (b *bucket[T]) IsClose() (ok bool) {
	return atomic.LoadUint32(&b.close) == 1 && atomic.LoadUint32(&b.op_num) == 0

}
