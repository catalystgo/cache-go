package cache_test

import (
	"context"
	"testing"

	"github.com/catalystgo/kache/cache"
	"github.com/catalystgo/kache/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry(t *testing.T) {
	t.Parallel()

	const testName = "my-test-cache"

	t.Run("get empty", func(t *testing.T) {
		t.Parallel()

		r := cache.NewRegistry()

		// act
		_, ok := r.GetByName(testName)

		// assert
		require.False(t, ok)
	})

	t.Run("register", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		cache1 := mock.NewMockNamedCache(ctrl)
		cache1.EXPECT().Name().Return("cache-1").Times(2)

		cache2 := mock.NewMockNamedCache(ctrl)
		cache2.EXPECT().Name().Return("cache-2").Times(2)

		r := cache.NewRegistry()

		// act
		err := r.Register(cache1, cache2)

		// assert
		require.NoError(t, err)

		c1, ok := r.GetByName("cache-1")
		require.True(t, ok)
		require.Equal(t, cache1, c1)

		c2, ok := r.GetByName("cache-2")
		require.True(t, ok)
		require.Equal(t, cache2, c2)
	})

	t.Run("register nil cache", func(t *testing.T) {
		t.Parallel()

		r := cache.NewRegistry()

		// act
		err := r.Register(nil)

		// assert
		require.Error(t, err)
	})

	t.Run("register duplicate", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		cache1 := mock.NewMockNamedCache(ctrl)
		cache1.EXPECT().Name().Return("cache-1").Times(4)

		r := cache.NewRegistry()

		// act 1
		err1 := r.Register(cache1)

		// assert 1
		require.NoError(t, err1)

		// act 2
		err2 := r.Register(cache1)

		// assert 2
		require.Error(t, err2)
	})

	t.Run("must register error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		cache1 := mock.NewMockNamedCache(ctrl)

		var (
			fatalfCalls = 0
			logFatalf   = func(ctx context.Context, format string, args ...interface{}) {
				fatalfCalls++
			}
		)

		r := cache.NewRegistry(cache.WithLoggerFatalf(logFatalf))

		// act
		r.MustRegister(cache1, assert.AnError)

		// act
		require.Equal(t, 1, fatalfCalls)
	})

	t.Run("must register duplicate", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		cache1 := mock.NewMockNamedCache(ctrl)
		cache1.EXPECT().Name().Return("cache-1").Times(5)

		var (
			fatalfCalls = 0
			logFatalf   = func(ctx context.Context, format string, args ...interface{}) {
				fatalfCalls++
			}
		)

		r := cache.NewRegistry(cache.WithLoggerFatalf(logFatalf))

		// act
		r.MustRegister(cache1, nil)
		r.MustRegister(cache1, nil)

		// act
		require.Equal(t, 1, fatalfCalls)
	})

	t.Run("must register", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		cache1 := mock.NewMockNamedCache(ctrl)
		cache1.EXPECT().Name().Return("cache-1").Times(2)

		var (
			fatalfCalls = 0
			logFatalf   = func(ctx context.Context, format string, args ...interface{}) {
				fatalfCalls++
			}
		)

		r := cache.NewRegistry(cache.WithLoggerFatalf(logFatalf))

		// act
		r.MustRegister(cache1, nil)

		// act
		require.Equal(t, 0, fatalfCalls)
		_, ok := r.GetByName("cache-1")
		require.True(t, ok)
	})

	t.Run("maybe register error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		cache1 := mock.NewMockNamedCache(ctrl)

		var (
			logErrorfCalls = 0
			logErrorf      = func(ctx context.Context, format string, args ...interface{}) {
				logErrorfCalls++
			}
		)

		r := cache.NewRegistry(cache.WithLoggerErrorf(logErrorf))

		// act
		r.MaybeRegister(cache1, assert.AnError)

		// assert
		require.Equal(t, 1, logErrorfCalls)
	})

	t.Run("maybe register", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		cache1 := mock.NewMockNamedCache(ctrl)
		cache1.EXPECT().Name().Return("cache-1").Times(2)

		var (
			logErrorfCalls = 0
			logErrorf      = func(ctx context.Context, format string, args ...interface{}) {
				logErrorfCalls++
			}
		)

		r := cache.NewRegistry(cache.WithLoggerErrorf(logErrorf))

		// act
		r.MaybeRegister(cache1, nil)

		// assert
		require.Equal(t, 0, logErrorfCalls)
		_, ok := r.GetByName("cache-1")
		require.True(t, ok)
	})
}
