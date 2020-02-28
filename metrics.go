package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	namespace     = "znapzend"
	preSnapMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "presnap_command_started",
		Help:      "whether the command to run prior zfs snapshot was started",
	}, []string{"job"})
	postSnapMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "postsnap_command_finished",
		Help:      "whether the command to run after zfs snapshot was finished",
	}, []string{"job"})
	preSendMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "presend_command_started",
		Help:      "whether the command to run prior zfs send was started",
	}, []string{"job"})
	postSendMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "postsend_command_finished",
		Help:      "whether the command to run after zfs send was finished",
	}, []string{"job"})
	metricVector = []*prometheus.GaugeVec{preSnapMetric, postSnapMetric, preSendMetric, postSendMetric}
)

type (
	// ResetMetricTuple contains a gauge and its enable flag.
	ResetMetricTuple struct {
		resetEnabled bool
		vec          *prometheus.GaugeVec
	}
)

func (p *Parameters) setMetric(vec *prometheus.GaugeVec) {
	gauge := vec.WithLabelValues(p.JobName)
	gauge.Set(1)
	if p.SelfResetAfter > 0 {
		go func() {
			logEntry := log.WithFields(log.Fields{"job": p.JobName})
			logEntry.WithField("delay", p.SelfResetAfter).Debug("Delaying job reset.")
			time.Sleep(p.SelfResetAfter)
			gauge.Set(0)
			logEntry.Info("Reset gauge.")
		}()
	}
}

// ResetMetrics resets all given gauges to 0 if the flag is set to true.
func ResetMetrics(label string, tuples ...ResetMetricTuple) {
	for _, tuple := range tuples[:] {
		if !tuple.resetEnabled {
			continue
		}
		tuple.vec.WithLabelValues(label).Set(0)
	}
}

// RegisterMetric registers 4 new gauges with the given label (preSnap, postSnap, preSend, postSend) and initializes the
// values with 0.
func RegisterMetric(label string) error {
	logEvent := log.WithField("label", label)
	for _, vec := range metricVector {
		gauge := vec.WithLabelValues(label)
		gauge.Set(0)
		logEvent.WithField("metric", gauge.Desc().String()).Debug("Registered metric.")
	}
	return nil
}

// UnregisterMetric deletes 4 gauges with the given label, if found.
func UnregisterMetric(label string) {
	for _, vec := range metricVector {
		vec.DeleteLabelValues(label)
	}
	log.WithField("label", label).Debug("Unregistered metric.")
}
