`go get github.com/murinj/bucketQueue`

An improved queue based on a lock-free queue, which divides incoming elements into buckets based on their labels, which enables queueing out in buckets or in elements, so that outgoing elements are locally as homolabeled as possible

Currently, a thread-safe version of the direct implementation of lock-free queues is supported, with thread-unsafe and compromise solutions and some performance-optimizing improvements to be added over time.

