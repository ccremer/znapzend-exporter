package main

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	"net/http"
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
	// JobContext wraps the parsed query parameters and the HTTP request context.
	JobContext struct {
		Parameters Parameters
		Context    *gin.Context
	}
)

// NewJobContext parses the input data. On errors, the response is already set and logged.
func NewJobContext(c *gin.Context) (*JobContext, error) {
	p, err := ParseAndValidateInput(c)
	if err != nil {
		SetLog(c, log.ErrorLevel, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, err
	}
	return &JobContext{
		Parameters: p,
		Context:    c,
	}, nil
}

func (j *JobContext) setMetric(vec *prometheus.GaugeVec) error {
	p := j.Parameters
	gauge, err := vec.GetMetricWithLabelValues(p.JobName)
	if err != nil {
		SetError(j.Context, "Invalid job label", err, log.Fields{
			"job": p.JobName,
		})
		return err
	}
	gauge.Set(1)
	if p.AutoReset {
		go func() {
			logEntry := log.WithFields(log.Fields{"job": p.JobName})
			logEntry.WithField("delay", p.ResetAfter).Debug("Delaying job reset.")
			time.Sleep(p.ResetAfter)
			vec.WithLabelValues(p.JobName).Set(0)
			logEntry.Info("Reset gauge.")
		}()
	}
	return nil
}

func (j *JobContext) resetMetricIf(condition bool, vec *prometheus.GaugeVec) error {
	if !condition {
		return nil
	}
	p := j.Parameters
	gauge, err := vec.GetMetricWithLabelValues(p.JobName)
	if err != nil {
		SetError(j.Context, "Invalid job label", err, log.Fields{
			"job": p.JobName,
		})
		return err
	}
	gauge.Set(0)
	return nil
}

// RegisterMetric registers 4 new gauges with the given label (preSnap, postSnap, preSend, postSend) and initializes the
// values with 0.
func RegisterMetric(label string) error {
	logEvent := log.WithField("label", label)
	for _, vec := range metricVector {
		gauge, err := vec.GetMetricWithLabelValues(label)
		if err != nil {
			return err
		}
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
