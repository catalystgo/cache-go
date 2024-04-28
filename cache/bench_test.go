package cache_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/catalystgo/cache-go/cache"
	"github.com/catalystgo/cache-go/cache/arc"
	"github.com/catalystgo/cache-go/cache/lru"
	"github.com/catalystgo/cache-go/cache/ristretto"
	"github.com/catalystgo/cache-go/cache/twoqueue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	// "github.com/catalystgo/realtime-config-go/helper"
)

const (
	name       = "cache"
	capValue   = 100000
	ttlKey     = "ttlValue"
	capKey     = "capValue"
	ttlValue   = 1 * time.Minute
	maxAvgTime = time.Second
)

var (
	random   rand.Source
	testdata = make(map[interface{}]interface{}, capValue)
)

func init() {
	for i := 0; i < capValue; i++ {
		testdata[i] = i
	}
	random = rand.NewSource(time.Now().UTC().UnixNano())
}

func warmUp(cache cache.Cache) {
	for k, v := range testdata {
		cache.Put(k, v)
	}
}

func createLRU(t testing.TB) *lru.Cache {
	c, err := lru.NewCache(name, capValue, ttlValue)
	require.NoError(t, err)
	return c
}

func createARC(t testing.TB) *arc.Cache {
	c, err := arc.NewCache(name, capValue, ttlValue)
	require.NoError(t, err)
	return c
}

func create2Q(t testing.TB) *twoqueue.Cache {
	c, err := twoqueue.NewCache(name, capValue, ttlValue)
	require.NoError(t, err)
	return c
}

func createRistretto(t testing.TB) *ristretto.Cache {
	c, err := ristretto.New(name, capValue, ttlValue)
	require.NoError(t, err)
	return c
}

// func createRealtime(t testing.TB, opts ...realtime.CacheOption) *realtime.Cache {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()
// 	client := NewMockClient(ctrl)
// 	client.EXPECT().
// 		Value(gomock.Any(), capKey).Return(helper.NewValue(capValue), nil).AnyTimes()
// 	client.EXPECT().
// 		Value(gomock.Any(), ttlKey).Return(helper.NewValue(ttlValue), nil).AnyTimes()
// 	client.EXPECT().
// 		WatchVariable(gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(1).MinTimes(1)

// 	c, err := realtime.NewCache(client, name, capKey, ttlKey, opts...)
// 	require.NoError(t, err)
// 	return c
// }

func load(op func(), rps int) (stop chan struct{}) {
	stop = make(chan struct{})
	go func() {
		t, n, tc := time.Now(), 0, 0
		for {
			go func() {
				op()
			}()
			n++
			tc++
			if n == rps {
				t, n = time.Now(), 0
			}
			select {
			case <-stop:
				return
			default:

				time.Sleep(time.Until(t.Add(time.Second)) / time.Duration(rps-n))
				continue
			}
		}
	}()

	time.Sleep(1 * time.Second)
	return stop
}

func cacheOp(c cache.Cache, hitRate int) {
	v := rand.Intn(capValue)      //nolint:gosec
	if rand.Intn(100) < hitRate { //nolint:gosec
		c.Get(v)
	} else {
		c.Put(v, v)
	}
}

func bench(t testing.TB, op func()) time.Duration {
	var total, i, timeout time.Duration = 0, 0, 3 * time.Second
	start := time.Now()
	for {
		s := time.Now()
		op()
		d := time.Since(s)
		total += d
		i++
		if time.Since(start) > timeout {
			break
		}
	}
	avg := total / time.Duration(i)
	t.Logf("%s\t%d ns/op", t.Name(), avg)

	return avg
}

func benchCache(method string, cache cache.Cache, rps int, hitrate int) func(t *testing.T) {
	return func(t *testing.T) {
		warmUp(cache)
		l := load(func() { cacheOp(cache, hitrate) }, rps)
		defer close(l)

		got, exp := bench(t, func() {
			switch method {
			case "Get":
				cache.Get(rand.Intn(capValue)) //nolint:gosec
			case "Put":
				v := rand.Intn(capValue) //nolint:gosec
				cache.Put(v, v)
			default:
				t.Fatalf("unknown method %q", method)
			}
		}), maxAvgTime

		assert.Truef(t, got < exp, "expected avg op time < %s, got: %s", exp, got)
	}
}

func TestBenchmarks(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Run("LRU Get under load RPS 200k HitRate 90", benchCache("Get", createLRU(t), 200000, 90))
	t.Run("LRU Get under load RPS 400k HitRate 90", benchCache("Get", createLRU(t), 400000, 90))
	t.Run("LRU Get under load RPS 200k HitRate 10", benchCache("Get", createLRU(t), 200000, 10))
	t.Run("LRU Get under load RPS 400k HitRate 10", benchCache("Get", createLRU(t), 400000, 10))
	t.Run("LRU Put under load RPS 200k HitRate 90", benchCache("Put", createLRU(t), 200000, 90))
	t.Run("LRU Put under load RPS 400k HitRate 90", benchCache("Put", createLRU(t), 400000, 90))
	t.Run("LRU Put under load RPS 200k HitRate 10", benchCache("Put", createLRU(t), 200000, 10))
	t.Run("LRU Put under load RPS 400k HitRate 10", benchCache("Put", createLRU(t), 400000, 10))

	t.Run("ARC Get under load RPS 200k HitRate 90", benchCache("Get", createARC(t), 200000, 90))
	t.Run("ARC Get under load RPS 400k HitRate 90", benchCache("Get", createARC(t), 400000, 90))
	t.Run("ARC Get under load RPS 200k HitRate 10", benchCache("Get", createARC(t), 200000, 10))
	t.Run("ARC Get under load RPS 400k HitRate 10", benchCache("Get", createARC(t), 400000, 10))
	t.Run("ARC Put under load RPS 200k HitRate 90", benchCache("Put", createARC(t), 200000, 90))
	t.Run("ARC Put under load RPS 400k HitRate 90", benchCache("Put", createARC(t), 400000, 90))
	t.Run("ARC Put under load RPS 200k HitRate 10", benchCache("Put", createARC(t), 200000, 10))
	t.Run("ARC Put under load RPS 400k HitRate 10", benchCache("Put", createARC(t), 400000, 10))

	t.Run("2Q Get under load RPS 200k HitRate 90", benchCache("Get", create2Q(t), 200000, 90))
	t.Run("2Q Get under load RPS 400k HitRate 90", benchCache("Get", create2Q(t), 400000, 90))
	t.Run("2Q Get under load RPS 200k HitRate 10", benchCache("Get", create2Q(t), 200000, 10))
	t.Run("2Q Get under load RPS 400k HitRate 10", benchCache("Get", create2Q(t), 400000, 10))
	t.Run("2Q Put under load RPS 200k HitRate 90", benchCache("Put", create2Q(t), 200000, 90))
	t.Run("2Q Put under load RPS 400k HitRate 90", benchCache("Put", create2Q(t), 400000, 90))
	t.Run("2Q Put under load RPS 200k HitRate 10", benchCache("Put", create2Q(t), 200000, 10))
	t.Run("2Q Put under load RPS 400k HitRate 10", benchCache("Put", create2Q(t), 400000, 10))

	t.Run("Ristretto Get under load RPS 200k HitRate 90", benchCache("Get", createRistretto(t), 200000, 90))
	t.Run("Ristretto Get under load RPS 400k HitRate 90", benchCache("Get", createRistretto(t), 400000, 90))
	t.Run("Ristretto Get under load RPS 200k HitRate 10", benchCache("Get", createRistretto(t), 200000, 10))
	t.Run("Ristretto Get under load RPS 400k HitRate 10", benchCache("Get", createRistretto(t), 400000, 10))
	t.Run("Ristretto Put under load RPS 200k HitRate 90", benchCache("Put", createRistretto(t), 200000, 90))
	t.Run("Ristretto Put under load RPS 400k HitRate 90", benchCache("Put", createRistretto(t), 400000, 90))
	t.Run("Ristretto Put under load RPS 200k HitRate 10", benchCache("Put", createRistretto(t), 200000, 10))
	t.Run("Ristretto Put under load RPS 400k HitRate 10", benchCache("Put", createRistretto(t), 400000, 10))

	// t.Run("Realtime with LRU engine Get under load RPS 200k HitRate 90", benchCache("Get", createRealtime(t, realtime.WithEngine(realtime.EngineLRU)), 200000, 90))
	// t.Run("Realtime with LRU engine Get under load RPS 400k HitRate 90", benchCache("Get", createRealtime(t, realtime.WithEngine(realtime.EngineLRU)), 400000, 90))
	// t.Run("Realtime with LRU engine Get under load RPS 200k HitRate 10", benchCache("Get", createRealtime(t, realtime.WithEngine(realtime.EngineLRU)), 200000, 10))
	// t.Run("Realtime with LRU engine Get under load RPS 400k HitRate 10", benchCache("Get", createRealtime(t, realtime.WithEngine(realtime.EngineLRU)), 400000, 10))
	// t.Run("Realtime with LRU engine Put under load RPS 200k HitRate 90", benchCache("Put", createRealtime(t, realtime.WithEngine(realtime.EngineLRU)), 200000, 90))
	// t.Run("Realtime with LRU engine Put under load RPS 400k HitRate 90", benchCache("Put", createRealtime(t, realtime.WithEngine(realtime.EngineLRU)), 400000, 90))
	// t.Run("Realtime with LRU engine Put under load RPS 200k HitRate 10", benchCache("Put", createRealtime(t, realtime.WithEngine(realtime.EngineLRU)), 200000, 10))
	// t.Run("Realtime with LRU engine Put under load RPS 400k HitRate 10", benchCache("Put", createRealtime(t, realtime.WithEngine(realtime.EngineLRU)), 400000, 10))

	// t.Run("Realtime with ARC engine Get under load RPS 200k HitRate 90", benchCache("Get", createRealtime(t), 200000, 90))
	// t.Run("Realtime with ARC engine Get under load RPS 400k HitRate 90", benchCache("Get", createRealtime(t), 400000, 90))
	// t.Run("Realtime with ARC engine Get under load RPS 200k HitRate 10", benchCache("Get", createRealtime(t), 200000, 10))
	// t.Run("Realtime with ARC engine Get under load RPS 400k HitRate 10", benchCache("Get", createRealtime(t), 400000, 10))
	// t.Run("Realtime with ARC engine Put under load RPS 200k HitRate 90", benchCache("Put", createRealtime(t), 200000, 90))
	// t.Run("Realtime with ARC engine Put under load RPS 400k HitRate 90", benchCache("Put", createRealtime(t), 400000, 90))
	// t.Run("Realtime with ARC engine Put under load RPS 200k HitRate 10", benchCache("Put", createRealtime(t), 200000, 10))
	// t.Run("Realtime with ARC engine Put under load RPS 400k HitRate 10", benchCache("Put", createRealtime(t), 400000, 10))

	// t.Run("Realtime with 2Q engine Get under load RPS 200k HitRate 90", benchCache("Get", createRealtime(t, realtime.WithEngine(realtime.Engine2Q)), 200000, 90))
	// t.Run("Realtime with 2Q engine Get under load RPS 400k HitRate 90", benchCache("Get", createRealtime(t, realtime.WithEngine(realtime.Engine2Q)), 400000, 90))
	// t.Run("Realtime with 2Q engine Get under load RPS 200k HitRate 10", benchCache("Get", createRealtime(t, realtime.WithEngine(realtime.Engine2Q)), 200000, 10))
	// t.Run("Realtime with 2Q engine Get under load RPS 400k HitRate 10", benchCache("Get", createRealtime(t, realtime.WithEngine(realtime.Engine2Q)), 400000, 10))
	// t.Run("Realtime with 2Q engine Put under load RPS 200k HitRate 90", benchCache("Put", createRealtime(t, realtime.WithEngine(realtime.Engine2Q)), 200000, 90))
	// t.Run("Realtime with 2Q engine Put under load RPS 400k HitRate 90", benchCache("Put", createRealtime(t, realtime.WithEngine(realtime.Engine2Q)), 400000, 90))
	// t.Run("Realtime with 2Q engine Put under load RPS 200k HitRate 10", benchCache("Put", createRealtime(t, realtime.WithEngine(realtime.Engine2Q)), 200000, 10))
	// t.Run("Realtime with 2Q engine Put under load RPS 400k HitRate 10", benchCache("Put", createRealtime(t, realtime.WithEngine(realtime.Engine2Q)), 400000, 10))

	// t.Run("Realtime with Ristretto engine Get under load RPS 200k HitRate 90", benchCache("Get", createRealtime(t, realtime.WithEngine(realtime.EngineRistretto)), 200000, 90))
	// t.Run("Realtime with Ristretto engine Get under load RPS 400k HitRate 90", benchCache("Get", createRealtime(t, realtime.WithEngine(realtime.EngineRistretto)), 400000, 90))
	// t.Run("Realtime with Ristretto engine Get under load RPS 200k HitRate 10", benchCache("Get", createRealtime(t, realtime.WithEngine(realtime.EngineRistretto)), 200000, 10))
	// t.Run("Realtime with Ristretto engine Get under load RPS 400k HitRate 10", benchCache("Get", createRealtime(t, realtime.WithEngine(realtime.EngineRistretto)), 400000, 10))
	// t.Run("Realtime with Ristretto engine Put under load RPS 200k HitRate 90", benchCache("Put", createRealtime(t, realtime.WithEngine(realtime.EngineRistretto)), 200000, 90))
	// t.Run("Realtime with Ristretto engine Put under load RPS 400k HitRate 90", benchCache("Put", createRealtime(t, realtime.WithEngine(realtime.EngineRistretto)), 400000, 90))
	// t.Run("Realtime with Ristretto engine Put under load RPS 200k HitRate 10", benchCache("Put", createRealtime(t, realtime.WithEngine(realtime.EngineRistretto)), 200000, 10))
	// t.Run("Realtime with Ristretto engine Put under load RPS 400k HitRate 10", benchCache("Put", createRealtime(t, realtime.WithEngine(realtime.EngineRistretto)), 400000, 10))
}
