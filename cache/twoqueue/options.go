package twoqueue

type options struct {
	ghostEntriesRation float64
	recentEntriesRatio float64
}

type Option func(*options)

func WithRatios(ghostRatio, recentRatio float64) Option {
	return func(o *options) {
		o.ghostEntriesRation = ghostRatio
		o.recentEntriesRatio = recentRatio
	}
}
