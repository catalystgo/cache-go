package lru

import (
	"fmt"
	"io"
	"time"

	"github.com/catalystgo/kache/cache"
	"github.com/catalystgo/kache/cache/metrics"
	lru "github.com/hashicorp/golang-lru"
)

var (
	_ cache.NamedCache    = &Cache{}
	_ cache.WithTTLPutter = &Cache{}
	_ io.Closer           = &Cache{}
)

// Cache is a structure representing a wrapper over LRU cache (hashicorp).
type Cache struct {
	*lru.Cache

	close   chan struct{}
	metrics *metrics.CacheMetrics
	name    string
	cap     int
	ttl     time.Duration
}

type entry struct {
	value   interface{}
	expires time.Time
}

const (
	calcItemNumberInterval = 15 * time.Second
)

// NewCache creates a new LRU cache with capacity and ttl.
// If ttl == 0, the cache will be without ttl.
func NewCache(name string, cap int, ttl time.Duration) (*Cache, error) {
	return NewCacheWithEvictCallback(name, cap, ttl, nil)
}

// NewCacheWithEvictCallback is NewCache + setting a function to be called when an item is evicted from the cache.
// `onEvict` can be `nil`, in which case the function will not be called.
//
// Use WrapOnEvictWithUnwrapper to access your original item in the callback function.
// Without unwrapping using WrapOnEvictWithUnwrapper, the function will be called with typeof(value) == `entry`.
func NewCacheWithEvictCallback(name string, cap int, ttl time.Duration, onEvict func(key interface{}, value interface{})) (*Cache, error) {
	if cap <= 0 {
		return nil, fmt.Errorf("can't create cache %s: %w", name, cache.ErrWrongCapacity)
	}
	if ttl < 0 {
		return nil, fmt.Errorf("can't create cache %s: %w", name, cache.ErrWrongTTL)
	}

	lruCache, err := lru.NewWithEvict(cap, onEvict)
	if err != nil {
		return nil, err
	}
	c := &Cache{
		Cache:   lruCache,
		name:    name,
		cap:     cap,
		ttl:     ttl,
		close:   make(chan struct{}),
		metrics: metrics.NewCacheMetrics(name),
	}

	go c.stats()

	return c, nil
}

// Cap returns the cache capacity.
func (c *Cache) Cap() int {
	return c.cap
}

// Clear completely clears the cache.
func (c *Cache) Clear() {
	c.Cache.Purge()
}

// Close completely clears the cache.
// Always call Close after finishing using the cache to release resources.
func (c *Cache) Close() error {
	select {
	case _, ok := <-c.close:
		if !ok {
			return nil
		}
	default:
	}

	c.Clear()
	close(c.close)

	return nil
}

// Put puts a key-value pair into the cache.
// It uses the default TTL specified in NewCache.
func (c *Cache) Put(key, value interface{}) {
	c.PutWithTTL(key, value, c.ttl)
}

// PutWithTTL adds a key-value pair to the cache with a specified TTL.
// If ttl <= 0, no TTL is added.
func (c *Cache) PutWithTTL(key, value interface{}, ttl time.Duration) {
	start := time.Now()

	defer func() {
		c.metrics.ResponseTimeSet.Observe(metrics.SinceSeconds(start))
	}()

	var expires time.Time
	if ttl > 0 {
		expires = time.Now().Add(ttl)
	}

	c.Cache.Add(key, &entry{
		expires: expires,
		value:   value,
	})
}

// Get retrieves a value by a specific key from the cache,
// and returns a boolean indicating whether the value was found.
// If the TTL of the key has expired, it returns nil and false.
func (c *Cache) Get(key interface{}) (value interface{}, ok bool) {
	start := time.Now()
	expired := false

	defer func() {
		c.metrics.ResponseTimeGet.Observe(metrics.SinceSeconds(start))
		if ok {
			c.metrics.HitCount.Inc()
		} else if expired {
			c.metrics.ExpiredCount.Inc()
		} else {
			c.metrics.MissCount.Inc()
		}
	}()

	v, ok := c.Cache.Get(key)
	if ok {
		if v.(*entry).expires.IsZero() || time.Now().Before(v.(*entry).expires) {
			return v.(*entry).value, true
		}
		expired = true
	}
	return nil, false
}

// Peek retrieves a value by key without updating the access time,
// and returns a boolean indicating whether the value was found.
// If the TTL of the key has expired, it returns nil and false.
func (c *Cache) Peek(key interface{}) (value interface{}, ok bool) {
	v, ok := c.Cache.Peek(key)
	if ok && (v.(*entry).expires.IsZero() || time.Now().Before(v.(*entry).expires)) {
		return v.(*entry).value, true
	}
	return nil, false
}

// Remove removes a value by key from the cache.
func (c *Cache) Remove(key interface{}) {
	start := time.Now()

	defer func() {
		c.metrics.ResponseTimeDelete.Observe(metrics.SinceSeconds(start))
	}()

	c.Cache.Remove(key)
}

// Name returns the cache name.
func (c *Cache) Name() string {
	return c.name
}

// Keys returns a list of saved keys.
func (c *Cache) Keys() []interface{} {
	return c.Cache.Keys()
}

// SetCap sets the capacity of the cache to cap.
func (c *Cache) SetCap(cap int) error {
	_ = c.Resize(cap)
	c.cap = cap

	return nil
}

func (c *Cache) stats() {
	for {
		select {
		case <-time.After(calcItemNumberInterval):
		case <-c.close:
			return
		}
		c.metrics.ItemNumber.Set(float64(c.Len()))
	}
}

// WrapOnEvictWithUnwrapper unwraps private (*entry) value and passes your original value to orig callback
func WrapOnEvictWithUnwrapper(
	orig func(key interface{}, value interface{}),
) (wrapped func(key interface{}, value interface{})) {
	wrapped = func(key interface{}, value interface{}) {
		ent, ok := value.(*entry)
		if !ok {
			orig(key, value)
		} else {
			orig(key, ent.value)
		}
	}
	return wrapped
}
