package twoqueue_test

import (
	"testing"
	"time"

	"github.com/catalystgo/cache-go/cache/twoqueue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_PutWithTTL_TryToGetExpiredData_ShouldFail(t *testing.T) {
	c, err := twoqueue.NewCache("test", 2, 1*time.Second)
	require.NoError(t, err)
	c.PutWithTTL(1, 1, time.Nanosecond)

	time.Sleep(time.Nanosecond)
	_, ok := c.Get(1)

	assert.Equal(t, false, ok)
}

func TestCache_PutWithTTL_TryToGetActualData_ShouldOk(t *testing.T) {
	c, err := twoqueue.NewCache("test", 2, 1*time.Second)
	require.NoError(t, err)
	c.PutWithTTL(1, 1, 1*time.Second)

	v, ok := c.Get(1)

	assert.Equal(t, true, ok)
	assert.Equal(t, 1, v)
}

func TestCache_WithoutTTL_TryToGetActualData_ShouldOk(t *testing.T) {
	c, err := twoqueue.NewCache("test", 2, 0)
	require.NoError(t, err)
	c.Put(1, 1)

	v, ok := c.Get(1)

	assert.Equal(t, true, ok)
	assert.Equal(t, 1, v)
}

func TestCache_Keys_ShouldOK(t *testing.T) {
	c, err := twoqueue.NewCache("test", 5, 0)
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
