package cache

import (
	"errors"
	"fmt"
	"net/http"

	jump "github.com/lithammer/go-jump-consistent-hash"
)

// Cache is a struct that holds the cache data
type Cache struct {
	// name to metadata
	Pods                map[string]*Pod
	Replicas            int
	DnsFormatApiGateway string
	DnsFormatInCluster  string
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

type HashMode int

const (
	// HashMode is the mode used for hashing
	HashModeAPIGateway HashMode = iota
	HashModeInCluster
)

// TODO: handle case where there are no available pods
func (c *Cache) ConsistentHash(mode HashMode, i string, w http.ResponseWriter) (host string, err error) {
	// calculate the number of buckets based on available pods
	var numAvailable int32
	var bucketToIndex []int
	for _, pod := range c.Pods {
		if pod.IsAvailable {
			numAvailable++
			bucketToIndex = append(bucketToIndex, pod.Index)
		}
	}

	// if unavailable throw an error
	if numAvailable == 0 {
		err = errors.New("no available pods")
		http.Error(w, "no available pods", http.StatusServiceUnavailable)
		return
	}

	// compute format string
	var dnsFormat string
	switch mode {
	case HashModeAPIGateway:
		dnsFormat = c.DnsFormatApiGateway
	case HashModeInCluster:
		dnsFormat = c.DnsFormatInCluster
	}

	// compute hash
	h := jump.HashString(i, numAvailable, jump.NewCRC64())
	index := bucketToIndex[h]
	host = fmt.Sprintf(dnsFormat, index)
	return
}
