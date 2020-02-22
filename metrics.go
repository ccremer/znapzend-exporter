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
	JobContext struct {
		Parameters Parameters
		Context    *gin.Context
	}
)

func NewJobContext(c *gin.Context) (*JobContext, error) {
	if p, err := ParseAndValidateInput(c); err != nil {
		SetLog(c, log.ErrorLevel, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, err
	} else {
		return &JobContext{
			Parameters: p,
			Context:    nil,
		}, nil
	}
}

func (j *JobContext) SetMetric(vec *prometheus.GaugeVec) error {
	p := j.Parameters
	gauge, err := vec.GetMetricWithLabelValues(p.JobName)
	if err != nil {
		return err
	}
	gauge.Set(1)
	if p.AutoReset {
		go func() {
			logEntry := log.WithFields(log.Fields{"job": p.JobName,})
			logEntry.WithField("delay", p.ResetAfter).Debug("Delaying job reset.")
			time.Sleep(p.ResetAfter)
			vec.WithLabelValues(p.JobName).Set(0)
			logEntry.Info("Reset gauge.")
		}()
	}
	return nil
}

func (j *JobContext) ResetMetricIf(condition bool, vec *prometheus.GaugeVec) error {
	return nil