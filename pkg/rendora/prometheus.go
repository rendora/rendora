/*
Copyright 2018 George Badawi.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
