package ristretto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_PutWithTTL_TryToGetExpiredData_ShouldFail(t *testing.T) {
	c, err := New("test", 1, time.Minute)
	require.NoError(t, err)
	c.PutWithTTL(1, 1, time.Millisecond)

	time.Sleep(5 * time.Millisecond)
	_, ok := c.Get(1)

	assert.Equal(t, false, ok)
}

func TestCache_PutWithTTL_TryToGetActualData_ShouldOk(t *testing.T) {
	c, err := New("test", 1, time.Minute)
	require.NoError(t, err)
	c.PutWithTTL(1, 1, time.Minute)
	time.Sleep(5 * time.Millisecond)
	v, ok := c.Get(1)

	assert.Equal(t, true, ok)
	assert.Equal(t, 1, v)
}

func TestCache_WithoutTTL_TryToGetActualData_ShouldOk(t *testing.T) {
	c, err := New("test", 1, 0)
	require.NoError(t, err)
	c.Put(1, 1)

	time.Sleep(5 * time.Millisecond)

	v, ok := c.Get(1)

	assert.Equal(t, true, ok)
	assert.Equal(t, 1, v)
}
