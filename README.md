
# Bucket Queue
<div id="top" align="center">
  <img src="assert\image\struct.png" width="500px"/>
  <div>&nbsp;</div>
  <div align="center">
    <font size="5"><b>Bucket Queue</font>
   </div>
</div>

bucket queue is a data structure designed to meet the needs of a specific application scenario.

Bucket queue is an improved queue based on lock-free queues. It is characterized by dividing elements of the same label into the same bucket, and creating a new bucket when the bucket capacity is exceeded. 
This allows elements to be efficiently queued out in buckets, ensuring that each bucket element has the same label. 

The design objective is to facilitate one-time processing of elements with the same label in scenarios with localized optimization needs, while avoiding potential starvation problems caused by direct storage.


For an intuitive understanding of bucket queues, consider the following input/output example (with a bucket capacity of 4).

> Input element label:
> 
> `1, 1, 2, 3, 3, 2, 2, 3, 1, 1, 2, 3, 1, 2, 1`
> 
> Output element labels (where [] denotes a bucket):
> 
>`[1, 1, 1, 1], [2, 2, 2, 2], [3, 3, 3, 3], [1, 1], [2]`




## Install
`go get github.com/murinj/bucketQueue`


## Usage
```go
input_value := []int{1, 1, 2, 3, 3, 2, 2, 3, 1, 1, 2, 3, 1, 2, 1}

q := NewBucketQueue[int](1024,1024,4)

for _, val := range input_value {
    q.Put(val, val)
}


for {
    bucket_i, ok := q.Get()
    vals := make([]int, len(bucket_i))
    for i, v := range bucket_i {
        vals[i] = v.(int)
    }

    if !ok {
        break
    } 
	
	printf("%v ",vals)
}
//[1, 1, 1, 1], [2, 2, 2, 2], [3, 3, 3, 3], [1, 1], [2]



```
