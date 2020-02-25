package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

type (
	// Parameters contains query parameters that modify the behaviour of the exporter
	Parameters struct {
		JobName       string
		ResetPreSnap  bool          `binding:"-"`
		ResetPostSnap bool          `binding:"-"`
		ResetPreSend  bool          `binding:"-"`
		ResetPostSend bool          `binding:"-"`
		ResetAfter    time.Duration `binding:"-"`
		AutoReset     bool          `binding:"-"`
	}
)

var (
	promHandler = promhttp.Handler()
)

func handlePreSnap(context *gin.Context) {
	job, err := NewJobContext(context)
	if err != nil {
		return
	}
	if err := job.setMetric(preSnapMetric); err != nil {
		return
	}
}

func handlePostSnap(context *gin.Context) {
	job, err := NewJobContext(context)
	if err != nil {
		return
	}
	if err := job.setMetric(postSnapMetric); err != nil {
		return
	}
	if err := job.resetMetricIf(job.Parameters.ResetPreSnap, preSnapMetric); err != nil {
		return
	}
}

func handlePreSend(context *gin.Context) {
	job, err := NewJobContext(context)
	if err != nil {
		return
	}
	if err := job.setMetric(preSendMetric); err != nil {
		return
	}
	if err := job.resetMetricIf(job.Parameters.ResetPreSnap, preSnapMetric); err != nil {
		return
	}
	if err := job.resetMetricIf(job.Parameters.ResetPostSnap, postSnapMetric); err != nil {
		return
	}
}

func handlePostSend(context *gin.Context) {
	job, err := NewJobContext(context)
	if err != nil {
		return
	}
	if err := job.setMetric(postSendMetric); err != nil {
		return
	}
	if err := job.resetMetricIf(job.Parameters.ResetPreSnap, preSnapMetric); err != nil {
		return
	}
	if err := job.resetMetricIf(job.Parameters.ResetPostSnap, postSnapMetric); err != nil {
		return
	}
	if err := job.resetMetricIf(job.Parameters.ResetPreSend, preSendMetric); err != nil {
		return
	}
}

func handleRegister(context *gin.Context) {
	job, err := NewJobContext(context)
	if err != nil {
		return
	}
	if err := RegisterMetric(job.Parameters.JobName); err != nil {
		SetLogWithFields(context, log.WarnLevel, "Could not register metric.", log.Fields{
			"error": err,
		})
		context.JSON(http.StatusBadRequest, gin.H{
			"status": "failed",
			"job":    job.Parameters.JobName,
			"error":  err.Error(),
		})
	}
	context.JSON(http.StatusOK, gin.H{"status": "registered", "job": job.Parameters.JobName})
}

func handleUnregister(context *gin.Context) {
	job, err := NewJobContext(context)
	if err != nil {
		return
	}
	UnregisterMetric(job.Parameters.JobName)
	context.JSON(http.StatusOK, gin.H{"status": "unregistered", "job": job.Parameters.JobName})
}

func handleMetrics(context *gin.Context) {
	SetLog(context, log.DebugLevel, "Accessing metrics.")
	promHandler.ServeHTTP(context.Writer, context.Request)
}

// For now, only one endpoint is required.
func handleLiveness(context *gin.Context) {
	SetLogLevel(context, log.DebugLevel)
	context.JSON(http.StatusOK, gin.H{
		"message": "If you can reach this, I'm alive!",
	})
}

func handleRoot(context *gin.Context) {
	SetLogLevel(context, log.DebugLevel)
	context.JSON(http.StatusOK, gin.H{
		"message": "exporter reachable. You might want to check /metrics",
		"version": version,
	})
}

// ParseAndValidateInput parses the query parameters from a given Gin HTTP request. Returns an error upon constraint violations.
func ParseAndValidateInput(context *gin.Context) (Parameters, error) {
	p := Parameters{}
	if p.JobName = strings.TrimPrefix(context.Param("job"), "/"); p.JobName == "" {
		return p, errors.New("missing Job name in URL")
	}
	if err := context.ShouldBindQuery(&p); err != nil {
		return p, err
	}
	log.WithFields(log.Fields{
		"parameters": p,
	}).Debug("Validated Input Data.")
	return p, nil
}
