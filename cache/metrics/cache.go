package metrics

import (
	"time"
	// "github.com/catalystgo/metrics-go/metrics"
)

const (
	labelOperationSet    = "set"
	labelOperationGet    = "get"
	labelOperationDelete = "delete"

	labelSet       = "set"
	labelOperation = "operation"
)

// var (
// 	itemNumber = metrics.NewGaugeVec(
// 		"struct_cache_items_total",
// 		"Total items in struct cache.",
// 		[]string{labelSet},
// 	)
// 	hitCount = metrics.NewCounterVec(
// 		"struct_cache_hit_total",
// 		"Counter of hits to struct cache.",
// 		[]string{labelSet},
// 	)
// 	missCount = metrics.NewCounterVec(
// 		"struct_cache_miss_total",
// 		"Counter of misses to struct cache.",
// 		[]string{labelSet},
// 	)
// 	expiredCount = metrics.NewCounterVec(
// 		"struct_cache_expired_total",
// 		"Counter of expired items in struct cache.",
// 		[]string{labelSet},
// 	)
// 	responseTime = metrics.NewHistogramVec(
// 		"struct_cache_request_duration_seconds",
// 		"Histogram of RT for the request to struct cache (seconds).",
// 		metrics.TimeBucketsFast,
// 		[]string{labelOperation, labelSet},
// 	)
// )

// func init() {
// 	metrics.MustRegister(
// 		itemNumber,
// 		hitCount,
// 		missCount,
// 		expiredCount,
// 		responseTime,
// 	)
// }

// NewCacheMetrics is a constructor for the CacheMetrics structure, which contains metrics for the cache.
func NewCacheMetrics(name string) *CacheMetrics {
	ncm := noopCacheMetrics{}
	return &CacheMetrics{
		// TODO: USE REAL IMPLEMENTATION FROM METRICS_GO
		// ResponseTimeSet:    responseTime.WithLabelValues(labelOperationSet, name),
		// ResponseTimeGet:    responseTime.WithLabelValues(labelOperationGet, name),
		// ResponseTimeDelete: responseTime.WithLabelValues(labelOperationDelete, name),
		// HitCount:           hitCount.WithLabelValues(name),
		// ExpiredCount:       expiredCount.WithLabelValues(name),
		// MissCount:          missCount.WithLabelValues(name),
		// ItemNumber:         itemNumber.WithLabelValues(name),
		ResponseTimeSet:    ncm,
		ResponseTimeGet:    ncm,
		ResponseTimeDelete: ncm,
		HitCount:           ncm,
		ExpiredCount:       ncm,
		MissCount:          ncm,
		ItemNumber:         ncm,
	}
}

type noopCacheMetrics struct{}

func (n noopCacheMetrics) Inc() { /* USE REAL IMPLEMENTATION FROM METRICS_GO */ }

func (n noopCacheMetrics) Observe(float64) { /* USE REAL IMPLEMENTATION FROM METRICS_GO */ }

func (n noopCacheMetrics) Set(float64) { /* USE REAL IMPLEMENTATION FROM METRICS_GO */ }

type counter interface {
	// Inc increments the counter by 1.
	Inc()
}

type histogram interface {
	// Observe adds a single value to the histogram.
	Observe(float64)
}

type gauge interface {
	// Set sets the value of the gauge.
	Set(float64)
}

// CacheMetrics structure for cache metrics.
type CacheMetrics struct {
	ResponseTimeSet    histogram
	ResponseTimeGet    histogram
	ResponseTimeDelete histogram

	HitCount     counter
	ExpiredCount counter
	MissCount    counter

	ItemNumber gauge
}

// SinceSeconds is a wrapper for time.Since(), converting the result to seconds.
func SinceSeconds(started time.Time) float64 {
	return float64(time.Since(started)) / float64(time.Second)
}
