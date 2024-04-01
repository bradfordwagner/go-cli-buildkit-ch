package cache

import (
	"fmt"

	jump "github.com/lithammer/go-jump-consistent-hash"
)

// Cache is a struct that holds the cache data
type Cache struct {
	// name to metadata
	Pods      map[string]*Pod
	Replicas  int
	DnsFormat string
}

// Pod is a struct that holds the pod data
type Pod struct {
	IsAvailable bool
	Index       int
}

// NewCache is a function that creates a new cache
func NewCache() *Cache {
	return &Cache{
		Pods:     make(map[string]*Pod),
		Replicas: -1,
	}
}

// GetPod is a function that returns a pod from the cache
func (c *Cache) GetPod(name string) (v *Pod) {
	var ok bool
	if v, ok = c.Pods[name]; !ok {
		v = &Pod{}
		c.Pods[name] = v
	}
	return v
}

// TODO: handle case where there are no available pods
func (c *Cache) ConsistentHash(i string) string {
	// calculate the number of buckets based on available pods
	var numAvailable int32
	var bucketToIndex []int
	for _, pod := range c.Pods {
		if pod.IsAvailable {
			numAvailable++
			bucketToIndex = append(bucketToIndex, pod.Index)
		}
	}

	// compute hash
	h := jump.HashString(i, numAvailable, jump.NewCRC64())
	// get index
	index := bucketToIndex[h]
	//format string
	return fmt.Sprintf(c.DnsFormat, index)
}
