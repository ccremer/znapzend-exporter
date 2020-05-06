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
	}, []string{"job", "target_host"})
	postSendMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "postsend_command_finished",
		Help:      "whether the command to run after zfs send was finished",
	}, []string{"job", "target_host"})
	metricVector = []*prometheus.GaugeVec{preSnapMetric, postSnapMetric, preSendMetric, postSendMetric}
)

type (
	// ResetMetricTuple contains a gauge and its enable flag.
	ResetMetricTuple struct {
		resetEnabled bool
		targetHost   string
		vec          *prometheus.GaugeVec
	}
)

func (p *Job) setMetric(vec *prometheus.GaugeVec) {
	gauge := vec.WithLabelValues(p.JobName)
	p.setValue(gauge)
}

func (p *Job) setMetricWithHost(vec *prometheus.GaugeVec) {
	gauge := vec.With(prometheus.Labels{
		"job":         p.JobName,
		"target_host": p.TargetHost,
	})
	p.setValue(gauge)
}

func (p *Job) setValue(gauge prometheus.Gauge) {
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
func (p *Job) ResetMetrics(tuples ...ResetMetricTuple) {
	for _, tuple := range tuples[:] {
		if !tuple.resetEnabled {
			continue
		}
		if tuple.targetHost == "" {
			tuple.vec.WithLabelValues(p.JobName).Set(0)
		} else {
			tuple.vec.WithLabelValues(p.JobName, tuple.targetHost).Set(0)
		}
	}
}

// RegisterMetric registers 4 new gauges with the given label (preSnap, postSnap, preSend, postSend) and initializes the
// values with 0.
func (p *Job) RegisterMetric() error {
	logEvent := log.WithField("jobName", p.JobName)
	preSnapMetric.WithLabelValues(p.JobName).Set(1)
	postSnapMetric.WithLabelValues(p.JobName).Set(1)
	if p.TargetHost != "" {
		preSendMetric.WithLabelValues(p.JobName, p.TargetHost).Set(1)
		postSendMetric.WithLabelValues(p.JobName, p.TargetHost).Set(1)
	}
	logEvent.Debug("Registered metric.")
	return nil
}

// UnregisterMetric deletes 4 gauges with the given label, if found.
func (p *Job) UnregisterMetric() {
	for _, vec := range metricVector {
		if p.TargetHost == "" {
			vec.DeleteLabelValues(p.JobName)
		} else {
			vec.DeleteLabelValues(p.JobName, p.TargetHost)
		}
	}
	log.WithField("job", p.JobName).Debug("Unregistered metric.")
}
