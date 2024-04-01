package cache

// Cache is a struct that holds the cache data
type Cache struct {
	// name to metadata
	Pods     map[string]Pod
	Replicas int
}

// Pod is a struct that holds the pod data
type Pod struct {
	IsAvailable bool
}

// NewCache is a function that creates a new cache
func NewCache() *Cache {
	return &Cache{
		Pods:     make(map[string]Pod),
		Replicas: -1,
	}
}
