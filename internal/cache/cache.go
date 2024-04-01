package cache

// Cache is a struct that holds the cache data
type Cache struct {
	// name to metadata
	Pods     map[string]*Pod
	Replicas int
}

// Pod is a struct that holds the pod data
type Pod struct {
	IsAvailable bool
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
