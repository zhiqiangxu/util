package metrics

import (
	"sync"

	kitmetrics "github.com/go-kit/kit/metrics"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

var (

	// metric maps

	counterMap = make(map[string]kitmetrics.Counter)
	cl         sync.Mutex

	gaugeMap = make(map[string]kitmetrics.Gauge)
	gl       sync.Mutex

	histMap = make(map[string]kitmetrics.Histogram)
	hl      sync.Mutex
)

// RegisterCounter registers a counter
func RegisterCounter(name string, labels []string) kitmetrics.Counter {
	cl.Lock()
	defer cl.Unlock()

	counter, ok := counterMap[name]
	if ok {
		panic("registered the same counter twice")
	}

	counter = kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{Name: name}, labels)
	counterMap[name] = counter
	return counter
}

// GetCounter get a Counter by name
func GetCounter(name string) kitmetrics.Counter {
	cl.Lock()
	defer cl.Unlock()

	counter, ok := counterMap[name]
	if ok {
		return counter
	}

	panic("unregistered counter")

}

// RegisterGauge registers a gauge
func RegisterGauge(name string, labels []string) kitmetrics.Gauge {
	gl.Lock()
	defer gl.Unlock()

	gauge, ok := gaugeMap[name]
	if ok {
		panic("registered the same gauge twice")
	}

	gauge = kitprometheus.NewGaugeFrom(stdprometheus.GaugeOpts{Name: name}, labels)
	gaugeMap[name] = gauge
	return gauge
}

// GetGauge get a Gauge by name
func GetGauge(name string) kitmetrics.Gauge {
	gl.Lock()
	defer gl.Unlock()

	gauge, ok := gaugeMap[name]
	if ok {
		return gauge
	}

	panic("unregistered gauge")

}

// RegisterHist registers a hist
func RegisterHist(name string, labels []string) kitmetrics.Histogram {
	hl.Lock()
	defer hl.Unlock()

	hist, ok := histMap[name]
	if ok {
		panic("registered the same gauge twice")
	}

	hist = kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{Name: name, Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}}, labels)
	histMap[name] = hist
	return hist
}

// GetHist get a Hist by name
func GetHist(name string) kitmetrics.Histogram {
	hl.Lock()
	defer hl.Unlock()

	hist, ok := histMap[name]
	if ok {
		return hist
	}

	panic("unregistered gauge")

}
