package rendora

import "github.com/prometheus/client_golang/prometheus"

//metrics provides various Prometheus metrics
type metrics struct {
	Duration       prometheus.Histogram
	CountTotal     prometheus.Counter
	CountSSR       prometheus.Counter
	CountSSRCached prometheus.Counter
}

func (R *Rendora) initPrometheus() {
	ret := &metrics{}
	ret.CountTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "rendora_requests_total",
		Help: "Total Requests",
	})

	ret.CountSSR = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "rendora_requests_ssr",
		Help: "SSR Requests",
	})

	ret.CountSSRCached = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "rendora_requests_ssr_cached",
		Help: "Cached SSR Requests",
	})

	ret.Duration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "rendora_latency_ssr",
		Help:    "SSR Latency",
		Buckets: []float64{50, 100, 150, 200, 250, 300, 350, 400, 500},
	})

	prometheus.MustRegister(ret.CountTotal)
	prometheus.MustRegister(ret.CountSSR)
	prometheus.MustRegister(ret.Duration)
	R.metrics = ret
}
