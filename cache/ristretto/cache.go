package ristretto

import (
	"fmt"
	"io"
	"time"

	"github.com/catalystgo/cache-go/cache"
	"github.com/catalystgo/cache-go/cache/metrics"
	"github.com/dgraph-io/ristretto"
)

const (
	calcItemNumberInterval = 15 * time.Second
)

var (
	_ cache.NamedCache    = &Cache{}
	_ cache.WithTTLPutter = &Cache{}
	_ io.Closer           = &Cache{}
	_ cache.CapSetter     = &Cache{}
)

// Cache is a wrapper around ristretto.Cache.
type Cache struct {
	cache   *ristretto.Cache
	close   chan struct{}
	metrics *metrics.CacheMetrics
	name    string
	cap     int
	ttl     time.Duration
	cost    int64
}

// Config is a wrapper around ristretto.Config.
type Config struct {
	ristretto.Config
	TTL  time.Duration // Default TTL for cache keys
	Cost int64         // Cost parameter for calls to Set (default 1)
}

// BuildConfig creates a configuration based on https://github.com/dgraph-io/ristretto#config recommendations.
// capacity is the cache capacity, ttl is the default cache key lifetime.
func BuildConfig(capacity int, ttl time.Duration) Config {
	return Config{
		Config: ristretto.Config{
			NumCounters:        int64(capacity) * 10,
			MaxCost:            int64(capacity),
			BufferItems:        64,
			Metrics:            true,
			IgnoreInternalCost: true, // old behavior
		},
		TTL:  ttl,
		Cost: 1,
	}
}

// New returns a new Cache instance.
// name is the cache name, capacity is the cache capacity, ttl is the default cache key lifetime.
func New(name string, capacity int, ttl time.Duration) (*Cache, error) {
	return NewWithConfig(name, BuildConfig(capacity, ttl))
}

// NewWithConfig returns a new Cache instance with custom configuration.
// This is necessary if you want to customize the ristretto configuration.
//
//	config := ristretto.BuildConfig(1000, time.Minute)
//
//	config.Config.OnEvict = func(key, conflict uint64, value interface{}, cost int64) {
//	  // custom evict callback
//	}
//
//	c, err := ristretto.NewWithConfig("namespace", config)
func NewWithConfig(name string, config Config) (*Cache, error) {
	r, err := ristretto.NewCache(&config.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create ristretto: %w", err)
	}

	c := &Cache{
		cache:   r,
		close:   make(chan struct{}),
		metrics: metrics.NewCacheMetrics(name),
		name:    name,
		cap:     int(config.Config.MaxCost),
		ttl:     config.TTL,
		cost:    config.Cost,
	}

	go c.stats()

	return c, nil
}

// Cap returns the cache capacity.
func (c *Cache) Cap() int {
	return c.cap
}

// Len returns the size of the cache.
func (c *Cache) Len() int {
	return int(c.cache.Metrics.KeysAdded() - c.cache.Metrics.CostEvicted())
}

// Clear completely clears the cache.
func (c *Cache) Clear() {
	c.cache.Clear()
}

// Contains checks for the existence of a key in the cache.
func (c *Cache) Contains(key interface{}) bool {
	_, ok := c.cache.Get(key)

	return ok
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

	return c.cache.Get(key)
}

// Peek is the same as Get, but does not report metrics.
func (c *Cache) Peek(key interface{}) (value interface{}, ok bool) {
	return c.cache.Get(key)
}

// Put puts a key-value pair into the cache.
// It uses the default TTL specified in New.
func (c *Cache) Put(key interface{}, value interface{}) {
	c.PutWithTTL(key, value, c.ttl)
}

// PutWithTTL adds a key-value pair to the cache with a specified TTL.
// If ttl <= 0, no TTL is added.
func (c *Cache) PutWithTTL(key, value interface{}, ttl time.Duration) {
	start := time.Now()

	defer func() {
		c.metrics.ResponseTimeSet.Observe(metrics.SinceSeconds(start))
	}()

	if ttl == 0 {
		c.cache.Set(key, value, c.cost)
		return
	}

	c.cache.SetWithTTL(key, value, c.cost, ttl)
}

// Remove removes a value by key from the cache.
func (c *Cache) Remove(key interface{}) {
	start := time.Now()

	defer func() {
		c.metrics.ResponseTimeDelete.Observe(metrics.SinceSeconds(start))
	}()

	c.cache.Del(key)
}

// Name returns the cache name.
func (c *Cache) Name() string {
	return c.name
}

// Close completely clears the cache.
// Always call Close after using the cache to release resources.
func (c *Cache) Close() error {
	select {
	case _, ok := <-c.close:
		if !ok {
			return nil
		}
	default:
	}

	c.Clear()
	c.cache.Close()

	close(c.close)

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

// SetCap sets the MaxCost parameter, which can be interpreted as the cache capacity.
func (c *Cache) SetCap(cap int) error {
	c.cache.UpdateMaxCost(int64(cap))
	c.cap = cap

	return nil
}
