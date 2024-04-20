package lru_test

import (
	"testing"
	"time"

	"github.com/catalystgo/kache/cache/lru"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_PutWithTTL_TryToGetExpiredData_ShouldFail(t *testing.T) {
	c, err := lru.NewCache("test", 1, 1*time.Second)
	require.NoError(t, err)
	c.PutWithTTL(1, 1, time.Nanosecond)

	time.Sleep(time.Nanosecond)
	_, ok := c.Get(1)

	assert.Equal(t, false, ok)
}

func TestCache_PutWithTTL_TryToGetActualData_ShouldOk(t *testing.T) {
	c, err := lru.NewCache("test", 1, 1*time.Second)
	require.NoError(t, err)
	c.PutWithTTL(1, 1, 1*time.Second)

	v, ok := c.Get(1)

	assert.Equal(t, true, ok)
	assert.Equal(t, 1, v)
}

func TestCache_WithoutTTL_TryToGetActualData_ShouldOk(t *testing.T) {
	c, err := lru.NewCache("test", 1, 0)
	require.NoError(t, err)
	c.Put(1, 1)

	v, ok := c.Get(1)

	assert.Equal(t, true, ok)
	assert.Equal(t, 1, v)
}

func TestCache_WithOnEvictWithUnwrapper_ShouldCallbackWithOriginalValues(t *testing.T) {
	type myVal struct {
		val int
	}
	const (
		n        = 5
		overflow = 2
	)

	var putValues []*myVal
	var expectedEvicted []*myVal
	for i := 0; i < n+overflow; i++ {
		putValues = append(putValues, &myVal{val: i})
		// once we overflow, evictions (fifo order) should occur
		if i >= n {
			expectedEvicted = append(expectedEvicted, putValues[i-n])
		}
	}

	called := 0
	cb := func(key, value interface{}) {
		require.IsType(t, &myVal{}, value, "should be of type `*myVal` but is of type `%T`", value)
		require.Equal(t, expectedEvicted[called], value.(*myVal))
		called++
	}
	c, err := lru.NewCacheWithEvictCallback("test", n, 10*time.Second, lru.WrapOnEvictWithUnwrapper(cb))
	require.NoError(t, err)

	for i, val := range putValues {
		c.Put(i, val)
	}

	require.Equal(t, overflow, called)
}

func TestCache_Keys_ShouldOK(t *testing.T) {
	c, err := lru.NewCache("test", 5, 0)
	require.NoError(t, err)

	wantKeys := make([]interface{}, 5)

	for i := 0; i < 5; i++ {
		c.Put(i, i)
		wantKeys[i] = i
	}

	// act
	gotKeys := c.Keys()

	// assert
	assert.Equal(t, wantKeys, gotKeys)
}
