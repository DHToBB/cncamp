package metrics

import (
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

var funcLatency = CreateExecutionTimeMetric("hs", "time cost")

// Register 注册prometheus
func Register() {
	if err := prometheus.Register(funcLatency); err != nil {
		glog.Error(err.Error())
	}
}

func CreateExecutionTimeMetric(namespace, help string) *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "exec_latency_seconds",
			Help:      help,
			Buckets:   prometheus.ExponentialBuckets(0.001, 2, 15),
		},
		[]string{"step"},
	)
}

type ExecutionTimer struct {
	histo *prometheus.HistogramVec
	start time.Time
	last  time.Time
}

func (t *ExecutionTimer) ObserveTotal() {
	(*t.histo).WithLabelValues("total").Observe(time.Now().Sub(t.start).Seconds())
}

func NewTimer() *ExecutionTimer {
	now := time.Now()

	return &ExecutionTimer{
		funcLatency,
		now,
		now,
	}
}
