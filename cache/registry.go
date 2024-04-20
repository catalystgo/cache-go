package cache

import (
	"context"
	"fmt"
	"sync"

	"github.com/catalystgo/tracerok/logger"
)

// Registry is a registry with cache instances.
// Used by the gRPC interceptor (file interceptor.go).
type Registry interface {
	GetByName(name string) (NamedCache, bool)
	Register(caches ...NamedCache) error
	MaybeRegister(cache NamedCache, err error)
	MustRegister(cache NamedCache, err error)
}

type registryOpt func(*cacheRegistry)

func WithLoggerErrorf(f func(ctx context.Context, format string, args ...interface{})) registryOpt {
	return func(r *cacheRegistry) {
		r.logErrorf = f
	}
}

func WithLoggerFatalf(f func(ctx context.Context, format string, args ...interface{})) registryOpt {
	return func(r *cacheRegistry) {
		r.logFatalf = f
	}
}

type cacheRegistry struct {
	mu     sync.RWMutex
	caches map[string]NamedCache

	logErrorf func(ctx context.Context, format string, args ...interface{})
	logFatalf func(ctx context.Context, format string, args ...interface{})
}

// NewRegistry creates a registry with cache instances.
func NewRegistry(opts ...registryOpt) Registry {
	cr := &cacheRegistry{
		caches:    make(map[string]NamedCache),
		logErrorf: logger.Errorf,
		logFatalf: logger.Fatalf,
	}
	for _, opt := range opts {
		opt(cr)
	}
	return cr
}

// Register registers a named cache in the registry.
// Do not use this method concurrently with the ByName method, as it will provoke a race condition, including panic.
func (r *cacheRegistry) Register(cc ...NamedCache) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range cc {
		if s == nil {
			return fmt.Errorf("registry.Register: can't register <nil> cache")
		}
		if _, ok := r.caches[s.Name()]; ok {
			return fmt.Errorf("registry.Register: cache with the same Name %#q has already registered", s.Name())
		}
		r.caches[s.Name()] = s
	}
	return nil
}

// MaybeRegister is a helper for creating and registering a cache in one line.
// It logs an error in case of unsuccessful registration.
//
// Example:
//
//	cache.MaybeRegister(arc.NewCache("/Service/Method", size, ttl))
func (r *cacheRegistry) MaybeRegister(cache NamedCache, err error) {
	if err != nil {
		r.logErrorf(context.Background(), "cache.MaybeRegister: failed to create cache, got err: %s", err)
		return
	}
	if err = r.Register(cache); err != nil {
		r.logErrorf(context.Background(), "cache.MaybeRegister: failed to register cache %#q, got err: %s", cache.Name(), err)
	}
}

// MustRegister is a helper for creating and registering a cache in one line.
// It triggers a fatal error upon registration failure.
//
// Example:
//
//	cache.MustRegister(arc.NewCache("/Service/Method", size, ttl))
func (r *cacheRegistry) MustRegister(cache NamedCache, err error) {
	if err != nil {
		r.logFatalf(context.Background(), "cache.MaybeRegister: failed to create cache, got err: %s", err)
		return
	}
	if err = r.Register(cache); err != nil {
		r.logFatalf(context.Background(), "cache.MaybeRegister: failed to register cache %#q, got err: %s", cache.Name(), err)
	}
}

// GetByName returns a cache instance by name from the registry.
func (r *cacheRegistry) GetByName(name string) (NamedCache, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.caches[name]
	return s, ok
}
