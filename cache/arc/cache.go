package arc

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

// Cache is a structure representing a wrapper over ARC cache (hashicorp).
type Cache struct {
	*lru.ARCCache

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

// NewCache creates a new ARC cache with capacity and ttl.
// If ttl == 0, the cache will be without ttl.
func NewCache(name string, cap int, ttl time.Duration) (*Cache, error) {
	if cap <= 0 {
		return nil, fmt.Errorf("can't create cache %s: %w", name, cache.ErrWrongCapacity)
	}
	if ttl < 0 {
		return nil, fmt.Errorf("can't create cache %s: %w", name, cache.ErrWrongTTL)
	}

	arc, err := lru.NewARC(cap)
	if err != nil {
		return nil, err
	}
	c := &Cache{
		ARCCache: arc,
		name:     name,
		cap:      cap,
		ttl:      ttl,
		close:    make(chan struct{}),
		metrics:  metrics.NewCacheMetrics(name),
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
	c.ARCCache.Purge()
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

	c.ARCCache.Add(key, &entry{
		expires: expires,
		value:   value,
	})
}

// Get retrieves a value by a specific key from the cache.
// It also returns a boolean indicating whether the value was found.
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

	v, ok := c.ARCCache.Get(key)
	if ok {
		if v.(*entry).expires.IsZero() || time.Now().Before(v.(*entry).expires) {
			return v.(*entry).value, true
		}
		expired = true
	}
	return nil, false
}

// Peek retrieves a value by key without updating the access time or access frequency.
// It also returns a boolean indicating whether the value was found.
// If the TTL of the key has expired, it returns nil and false.
func (c *Cache) Peek(key interface{}) (value interface{}, ok bool) {
	v, ok := c.ARCCache.Peek(key)
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

	c.ARCCache.Remove(key)
}

// Name returns the cache name.
func (c *Cache) Name() string {
	return c.name
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

// Keys returns a list of saved keys.
func (c *Cache) Keys() []interface{} {
	return c.ARCCache.Keys()
}
