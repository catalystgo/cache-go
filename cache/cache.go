package cache

import (
	"errors"
	"time"
)

var (
	// ErrWrongTTL is the TTL error if it is less than 0.
	ErrWrongTTL = errors.New("wrong TTL, it should be >= 0")
	// ErrWrongCapacity is the capacity error if it is less than 0.
	ErrWrongCapacity = errors.New("wrong capacity, it should be positive")
)

// Cache is the common interface for all types of caches.
type Cache interface {
	// Cap returns the cache capacity.
	Cap() int
	// Len returns the size of the cache.
	Len() int
	// Clear completely clears the cache.
	Clear()
	// Contains checks for the presence of a key in the cache.
	Contains(key interface{}) bool
	// Get returns the value for the given key and marks this key as the most recently used.
	Get(key interface{}) (value interface{}, ok bool)
	// Peek returns the value for the given key without any changes to the cache.
	Peek(key interface{}) (value interface{}, ok bool)
	// Put stores the value in the cache with the specified key.
	Put(key, value interface{})
	// Remove removes the value from the cache by key.
	Remove(key interface{})
}

// NamedCache is the interface for a named cache.
type NamedCache interface {
	Cache
	Named
}

// Named represents an object with a name.
type Named interface {
	Name() string
}

// CapSetter is an interface for setting the capacity of a cache.
type CapSetter interface {
	SetCap(cap int) error
}

// TTLSetter is an interface for setting the TTL of a cache.
type TTLSetter interface {
	SetTTL(ttl time.Duration)
}

// WithTTLPutter is an interface for putting a value into the cache with a specified TTL.
type WithTTLPutter interface {
	PutWithTTL(key, value interface{}, ttl time.Duration)
}

// KeysGetter is an interface for cache implementations that support returning saved keys.
type KeysGetter interface {
	Keys() []interface{}
}
