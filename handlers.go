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
	// Job contains query parameters that modify the behaviour of the exporter
	Job struct {
		JobName        string
		ResetPreSnap   bool          `binding:"-"`
		ResetPostSnap  bool          `binding:"-"`
		ResetPreSend   bool          `binding:"-"`
		ResetPostSend  bool          `binding:"-"`
		SelfResetAfter time.Duration `binding:"-"`
		TargetHost     string        `binding:"-"`
	}
)

var (
	promHandler = promhttp.Handler()
)

const (
	parameterKey = "parameters"
)

func handlePreSnap(context *gin.Context) {
	job := context.MustGet(parameterKey).(Job)
	job.setMetric(preSnapMetric)
	job.ResetMetrics(
		ResetMetricTuple{job.ResetPostSnap, "", postSnapMetric},
		ResetMetricTuple{job.ResetPreSend, "", preSendMetric},
	)
}

func handlePostSnap(context *gin.Context) {
	job := context.MustGet(parameterKey).(Job)
	job.setMetric(postSnapMetric)
	job.ResetMetrics(
		ResetMetricTuple{job.ResetPreSnap, "", preSnapMetric},
		ResetMetricTuple{job.ResetPreSend, "", preSendMetric},
	)
}

func handlePreSend(context *gin.Context) {
	job := context.MustGet(parameterKey).(Job)
	job.setMetricWithHost(preSendMetric)
	job.ResetMetrics(
		ResetMetricTuple{job.ResetPreSnap, "", preSnapMetric},
		ResetMetricTuple{job.ResetPostSnap, "", postSnapMetric},
		ResetMetricTuple{job.ResetPostSend, job.TargetHost, postSendMetric},
	)
}

func handlePostSend(context *gin.Context) {
	job := context.MustGet(parameterKey).(Job)
	job.setMetricWithHost(postSendMetric)
	job.ResetMetrics(
		ResetMetricTuple{job.ResetPreSnap, "", preSnapMetric},
		ResetMetricTuple{job.ResetPostSnap, "", postSnapMetric},
		ResetMetricTuple{job.ResetPreSend, job.TargetHost, preSendMetric},
	)
}

func handleRegister(context *gin.Context) {
	job := context.MustGet(parameterKey).(Job)
	if err := job.RegisterMetric(); err != nil {
		SetLogWithFields(context, log.WarnLevel, "Could not register metric.", log.Fields{
			"error": err,
		})
		context.JSON(http.StatusBadRequest, gin.H{
			"status": "failed",
			"job":    job.JobName,
			"error":  err.Error(),
		})
	}
	context.JSON(http.StatusOK, gin.H{"status": "registered", "job": job.JobName})
}

func handleUnregister(context *gin.Context) {
	job := context.MustGet(parameterKey).(Job)
	job.UnregisterMetric()
	context.JSON(http.StatusOK, gin.H{"status": "unregistered", "job": job.JobName})
}

func handleMetrics(context *gin.Context) {
	SetLog(context, log.DebugLevel, "Accessing metrics.")
	promHandler.ServeHTTP(context.Writer, context.Request)
}

// For now, only one endpoint is required.
func handleHealthcheck(context *gin.Context) {
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
func ParseAndValidateInput(c *gin.Context) (Job, error) {
	p := Job{
		ResetPreSnap:  true,
		ResetPostSnap: true,
		ResetPreSend:  true,
		ResetPostSend: true,
	}
	if p.JobName = strings.TrimPrefix(c.Param("job"), "/"); p.JobName == "" {
		return p, errors.New("missing Job name in URL")
	}
	if err := c.ShouldBindQuery(&p); err != nil {
		return p, err
	}
	if strings.HasPrefix(c.Request.URL.Path, "/postsend") || strings.HasPrefix(c.Request.URL.Path, "/presend") {
		if p.TargetHost == "" {
			return p, errors.New("missing TargetHost parameter in query")
		}
	}
	log.WithFields(log.Fields{
		"parameters": p,
	}).Debug("Validated Input Data.")
	return p, nil
}

// InputValidationHandle returns a Gin handler that parses the input of the request and puts the parsed content into
// the Gin context keys for later retrieval.
func InputValidationHandle(paths ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		toValidate := false
		for _, path := range paths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				toValidate = true
			}
		}
		if !toValidate {
			return
		}
		parameters, err := ParseAndValidateInput(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			SetError(c, "Validation failed.", err, log.Fields{})
			return
		}

		c.Set("parameters", parameters)
	}
}

// ErrorHandle returns a Gin handler that prints the errors from c.Errors in JSON. Does nothing if the response was
// already written or if no errors were added to the context.
func ErrorHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		err := c.Errors.Last()
		if c.Writer.Written() {
			return
		}
		if err != nil {
			c.JSON(c.Writer.Status(), gin.H{
				"errors": c.Errors.Errors(),
			})
		}
	}
}
