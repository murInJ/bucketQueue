package bucketQueue

import (
	"fmt"
	"math"
	"runtime"
	"sync/atomic"
)

type Cache struct {
	putNo uint32
	getNo uint32
	value any
}

// Queue lock free queue
type Queue struct {
	queueCapacity uint32
	splitCapacity uint32
	splitBits     uint32
	capMod        uint32
	putPos        uint32
	getPos        uint32
	cache         []Cache
}

func NewQueue(capacity, splitCapacity uint32) *Queue {
	q := new(Queue)
	q.queueCapacity = minQuantity(capacity)
	q.splitCapacity = minQuantity(splitCapacity)
	q.capMod = q.queueCapacity - 1
	q.splitBits = bits(q.splitCapacity) - 1
	q.putPos = 0
	q.getPos = 0
	q.cache = make([]Cache, q.queueCapacity)
	for i := range q.cache {
		cache := &q.cache[i]
		cache.getNo = uint32(i)
		cache.putNo = uint32(i)
	}
	cache := &q.cache[0]
	cache.getNo = q.queueCapacity
	cache.putNo = q.queueCapacity
	return q
}

func (q *Queue) String() string {
	getPos := atomic.LoadUint32(&q.getPos)
	putPos := atomic.LoadUint32(&q.putPos)
	return fmt.Sprintf("Queue{queueCapacity: %v, capMod: %v, putPos: %v, getPos: %v}",
		q.queueCapacity, q.capMod, putPos, getPos)
}

func (q *Queue) Capaciity() uint32 {
	return q.queueCapacity
}

func (q *Queue) Quantity() uint32 {
	var putPos, getPos uint32
	var quantity uint32
	getPos = atomic.LoadUint32(&q.getPos)
	putPos = atomic.LoadUint32(&q.putPos)

	if putPos >= getPos {
		quantity = putPos - getPos
	} else {
		quantity = q.capMod + (putPos - getPos)
	}

	return quantity
}

// put queue functions
func (q *Queue) Put(val any) (ok bool, splitNum uint32, quantity uint32) {
	var putPos, putPosNew, getPos, posCnt uint32
	var cache *Cache
	capMod := q.capMod

	getPos = atomic.LoadUint32(&q.getPos)
	putPos = atomic.LoadUint32(&q.putPos)

	if putPos >= getPos {
		posCnt = putPos - getPos
	} else {
		posCnt = capMod + (putPos - getPos)
	}

	if posCnt >= capMod-1 {
		runtime.Gosched()
		return false, 0, posCnt
	}

	putPosNew = putPos + 1
	if !atomic.CompareAndSwapUint32(&q.putPos, putPos, putPosNew) {
		runtime.Gosched()
		return false, 0, posCnt
	}

	splitNum = countMultiples(putPos, putPosNew, q.splitBits)

	cache = &q.cache[putPosNew&capMod]

	for {
		getNo := atomic.LoadUint32(&cache.getNo)
		putNo := atomic.LoadUint32(&cache.putNo)
		if putPosNew == putNo && getNo == putNo {
			cache.value = val
			atomic.AddUint32(&cache.putNo, q.queueCapacity)
			return true, splitNum, posCnt + 1
		} else {
			runtime.Gosched()
		}
	}
}

// puts queue functions
func (q *Queue) Puts(values []any) (puts uint32, splitNum uint32, quantity uint32) {
	var putPos, putPosNew, getPos, posCnt, putCnt uint32
	capMod := q.capMod

	getPos = atomic.LoadUint32(&q.getPos)
	putPos = atomic.LoadUint32(&q.putPos)

	if putPos >= getPos {
		posCnt = putPos - getPos
	} else {
		posCnt = capMod + (putPos - getPos)
	}

	if posCnt >= capMod-1 {
		runtime.Gosched()
		return 0, 0, posCnt
	}

	if capPuts, size := q.queueCapacity-posCnt, uint32(len(values)); capPuts >= size {
		putCnt = size
	} else {
		putCnt = capPuts
	}
	putPosNew = putPos + putCnt

	if !atomic.CompareAndSwapUint32(&q.putPos, putPos, putPosNew) {
		runtime.Gosched()
		return 0, 0, posCnt
	}
	splitNum = countMultiples(putPos, putPosNew, q.splitBits)
	for posNew, v := putPos+1, uint32(0); v < putCnt; posNew, v = posNew+1, v+1 {
		var cache *Cache = &q.cache[posNew&capMod]
		for {
			getNo := atomic.LoadUint32(&cache.getNo)
			putNo := atomic.LoadUint32(&cache.putNo)
			if posNew == putNo && getNo == putNo {
				cache.value = values[v]
				atomic.AddUint32(&cache.putNo, q.queueCapacity)
				break
			} else {
				runtime.Gosched()
			}
		}
	}
	return putCnt, splitNum, posCnt + putCnt
}

// get queue functions
func (q *Queue) Get() (val any, ok bool, quantity uint32) {
	var putPos, getPos, getPosNew, posCnt uint32
	var cache *Cache
	capMod := q.capMod

	putPos = atomic.LoadUint32(&q.putPos)
	getPos = atomic.LoadUint32(&q.getPos)

	if putPos >= getPos {
		posCnt = putPos - getPos
	} else {
		posCnt = capMod + (putPos - getPos)
	}

	if posCnt < 1 {
		runtime.Gosched()
		return nil, false, posCnt
	}

	getPosNew = getPos + 1
	if !atomic.CompareAndSwapUint32(&q.getPos, getPos, getPosNew) {
		runtime.Gosched()
		return nil, false, posCnt
	}

	cache = &q.cache[getPosNew&capMod]

	for {
		getNo := atomic.LoadUint32(&cache.getNo)
		putNo := atomic.LoadUint32(&cache.putNo)
		if getPosNew == getNo && getNo == putNo-q.queueCapacity {
			val = cache.value
			cache.value = nil
			atomic.AddUint32(&cache.getNo, q.queueCapacity)
			return val, true, posCnt - 1
		} else {
			runtime.Gosched()
		}
	}
}

// gets queue functions
func (q *Queue) Gets(values []any) (gets, quantity uint32) {
	var putPos, getPos, getPosNew, posCnt, getCnt uint32
	capMod := q.capMod

	putPos = atomic.LoadUint32(&q.putPos)
	getPos = atomic.LoadUint32(&q.getPos)

	if putPos >= getPos {
		posCnt = putPos - getPos
	} else {
		posCnt = capMod + (putPos - getPos)
	}

	if posCnt < 1 {
		runtime.Gosched()
		return 0, posCnt
	}

	if size := uint32(len(values)); posCnt >= size {
		getCnt = size
	} else {
		getCnt = posCnt
	}
	getPosNew = getPos + getCnt

	if !atomic.CompareAndSwapUint32(&q.getPos, getPos, getPosNew) {
		runtime.Gosched()
		return 0, posCnt
	}

	for posNew, v := getPos+1, uint32(0); v < getCnt; posNew, v = posNew+1, v+1 {
		var cache *Cache = &q.cache[posNew&capMod]
		for {
			getNo := atomic.LoadUint32(&cache.getNo)
			putNo := atomic.LoadUint32(&cache.putNo)
			if posNew == getNo && getNo == putNo-q.queueCapacity {
				values[v] = cache.value
				cache.value = nil
				getNo = atomic.AddUint32(&cache.getNo, q.queueCapacity)
				break
			} else {
				runtime.Gosched()
			}
		}
	}

	return getCnt, posCnt - getCnt
}

// round 到最近的2的倍数
func minQuantity(v uint32) uint32 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}

func bits(num uint32) uint32 {
	return uint32(math.Log2(float64(num))) + 1
}

func countMultiples(n, m, bit uint32) uint32 {
	num := (m >> bit) - (n >> bit)
	return num
}
