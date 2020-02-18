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
		Name:      "presnap_command_finished",
		Help:      "whether the command to run after zfs snapshot was started",
	}, []string{"job"})
	preSendMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "presend_command_started",
		Help:      "whether the command to run prior zfs send was started",
	}, []string{"job"})
	postSendMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "presend_command_finished",
		Help:      "whether the command to run after zfs send was started",
	}, []string{"job"})
)

func SetMetric(vec *prometheus.GaugeVec, p *Parameters, reset bool) error {
	gauge, err := vec.GetMetricWithLabelValues(p.JobName)
	if err != nil {
		return err
	}
	gauge.Set(1)
	if reset {
		go func() {
			logEntry := log.WithFields(log.Fields{"job": p.JobName,})
			logEntry.WithField("delay", p.ResetAfter).Debug("Delaying job reset.")
			time.Sleep(p.ResetAfter)
			vec.WithLabelValues(p.JobName).Set(0)
			logEntry.WithField("gauge", gauge.Desc().String()).Info("Reset gauge.")
		}()
	}
	return nil
}
