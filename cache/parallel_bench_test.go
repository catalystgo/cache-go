package cache_test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/catalystgo/cache-go/cache"
	"github.com/catalystgo/cache-go/cache/arc"
	"github.com/catalystgo/cache-go/cache/lru"
	"github.com/catalystgo/cache-go/cache/ristretto"
)

func benchmarkParallelGet(b *testing.B, c cache.Cache, workers, iterations int) {
	keys := make([][]int, workers)
	for i := 0; i < workers; i++ {
		keys[i] = make([]int, iterations)
		for j := 0; j < iterations; j++ {
			keys[i][j] = rand.Intn(capValue) //nolint:gosec
		}
	}

	warmUp(c)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(workers)
		for i := 0; i < workers; i++ {
			workerKeys := keys[i]
			go func() {
				defer wg.Done()

				for j := 0; j < iterations; j++ {
					c.Get(workerKeys[j])
				}
			}()
		}

		wg.Wait()
	}
}

// pkg: github.com/catalystgo/cache-go/cache
// BenchmarkParallelGet/*lru.Cache_1-12                  16          65941646 ns/op          800095 B/op      99998 allocs/op
// BenchmarkParallelGet/*lru.Cache_4-12                   4         264020145 ns/op         3200220 B/op     399999 allocs/op
// BenchmarkParallelGet/*lru.Cache_8-12                   2         565829862 ns/op         6402508 B/op     800002 allocs/op
// BenchmarkParallelGet/*lru.Cache_16-12                  1        1079725471 ns/op        12806416 B/op    1600015 allocs/op
// BenchmarkParallelGet/*lru.Cache_32-12                  1        2143424758 ns/op        25620880 B/op    3200040 allocs/op
// BenchmarkParallelGet/*arc.Cache_1-12                  18          64991280 ns/op          800918 B/op     100000 allocs/op
// BenchmarkParallelGet/*arc.Cache_4-12                   4         258310478 ns/op         3200100 B/op     399999 allocs/op
// BenchmarkParallelGet/*arc.Cache_8-12                   2         554299090 ns/op         6401336 B/op     799997 allocs/op
// BenchmarkParallelGet/*arc.Cache_16-12                  1        1215714909 ns/op        28893168 B/op    1804020 allocs/op
// BenchmarkParallelGet/*arc.Cache_32-12                  1        2328896713 ns/op        41700416 B/op    3404009 allocs/op
// BenchmarkParallelGet/*ristretto.Cache_1-12            26          39008821 ns/op         1597117 B/op     101556 allocs/op
// BenchmarkParallelGet/*ristretto.Cache_4-12            16          62995929 ns/op         6251269 B/op     405958 allocs/op
// BenchmarkParallelGet/*ristretto.Cache_8-12            13          87404013 ns/op        11226728 B/op     809420 allocs/op
// BenchmarkParallelGet/*ristretto.Cache_16-12            6         190891075 ns/op        13428126 B/op    1601195 allocs/op
// BenchmarkParallelGet/*ristretto.Cache_32-12            3         384424426 ns/op        26532002 B/op    3201796 allocs/op
func BenchmarkParallelGet(b *testing.B) {
	lruFactory := func() cache.NamedCache {
		c, _ := lru.NewCache("data", capValue, time.Minute)

		return c
	}

	arcFactory := func() cache.NamedCache {
		c, _ := arc.NewCache("data", capValue, time.Minute)

		return c
	}

	ristrettoFactory := func() cache.NamedCache {
		c, _ := ristretto.New("ristretto", capValue, time.Minute)

		return c
	}

	cases := []struct {
		c          cache.NamedCache
		workers    int
		iterations int
	}{
		{c: lruFactory(), workers: 1, iterations: 100_000},
		{c: lruFactory(), workers: 4, iterations: 100_000},
		{c: lruFactory(), workers: 8, iterations: 100_000},
		{c: lruFactory(), workers: 16, iterations: 100_000},
		{c: lruFactory(), workers: 32, iterations: 100_000},

		{c: arcFactory(), workers: 1, iterations: 100_000},
		{c: arcFactory(), workers: 4, iterations: 100_000},
		{c: arcFactory(), workers: 8, iterations: 100_000},
		{c: arcFactory(), workers: 16, iterations: 100_000},
		{c: arcFactory(), workers: 32, iterations: 100_000},

		{c: ristrettoFactory(), workers: 1, iterations: 100_000},
		{c: ristrettoFactory(), workers: 4, iterations: 100_000},
		{c: ristrettoFactory(), workers: 8, iterations: 100_000},
		{c: ristrettoFactory(), workers: 16, iterations: 100_000},
		{c: ristrettoFactory(), workers: 32, iterations: 100_000},
	}

	for _, tc := range cases {
		b.Run(fmt.Sprintf("%T_%v", tc.c, tc.workers), func(b *testing.B) {
			benchmarkParallelGet(b, tc.c, tc.workers, tc.iterations)
		})
	}
}
