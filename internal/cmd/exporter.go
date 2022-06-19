package cmd

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

const (
	defaultNamespace   = "pg_listener"
	defaultMetricsPort = "9938" // https://github.com/prometheus/prometheus/wiki/Default-port-allocations
)

type exporter struct {
	metrics          *Metrics
	totalMessages    *prometheus.Desc
	totalHeartbeats  *prometheus.Desc
	totalErrors      *prometheus.Desc
	bufferSize       *prometheus.Desc
	lastCommittedWal *prometheus.Desc
}

func newExporter(metrics *Metrics) *exporter {
	return &exporter{
		metrics: metrics,
		totalMessages: prometheus.NewDesc(defaultNamespace+"_processed_messages_total",
			"Total amount processed logical messages",
			nil, nil,
		),
		totalHeartbeats: prometheus.NewDesc(defaultNamespace+"_received_heartbeats_total",
			"Total amount received heartbeats",
			nil, nil,
		),
		totalErrors: prometheus.NewDesc(defaultNamespace+"_logged_errors_total",
			"Total amount logged errors",
			nil, nil,
		),
		bufferSize: prometheus.NewDesc(defaultNamespace+"_raw_buffer_bytes",
			"Current amount of raw buffer size of logical message",
			nil, nil,
		),
		lastCommittedWal: prometheus.NewDesc(defaultNamespace+"_last_committed_wal_location",
			"Last successfully commited to producer wal",
			nil, nil,
		),
	}
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.totalMessages
	ch <- e.totalHeartbeats
	ch <- e.totalErrors
	ch <- e.bufferSize
	ch <- e.lastCommittedWal
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(e.totalMessages, prometheus.CounterValue, float64(e.metrics.totalMessages))
	ch <- prometheus.MustNewConstMetric(e.totalHeartbeats, prometheus.CounterValue, float64(e.metrics.totalHeartbeats))
	ch <- prometheus.MustNewConstMetric(e.totalErrors, prometheus.CounterValue, float64(e.metrics.totalErrors))
	ch <- prometheus.MustNewConstMetric(e.bufferSize, prometheus.GaugeValue, float64(e.metrics.bufferSize))
	ch <- prometheus.MustNewConstMetric(e.lastCommittedWal, prometheus.GaugeValue, float64(e.metrics.lastCommittedWal))
}

func runMetricsExporter(metrics *Metrics) error {
	prometheus.MustRegister(newExporter(metrics))

	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":"+defaultMetricsPort, nil)
	if err != nil {
		return err
	}
	return nil
}
