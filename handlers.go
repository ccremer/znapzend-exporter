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
	Parameters struct {
		JobName       string
		ResetPreSnap  bool          `binding:"-"`
		ResetPostSnap bool          `binding:"-"`
		ResetPreSend  bool          `binding:"-"`
		ResetPostSend bool          `binding:"-"`
		ResetAfter    time.Duration `binding:"-"`
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
	job.SetMetric(preSnapMetric, job.Parameters.ResetPreSnap)

}

func handlePostSnap(context *gin.Context) {
	job, err := NewJobContext(context)
	if err != nil {
		return
	}
	job.SetMetric(preSnapMetric, job.Parameters.ResetPostSnap)

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

func ParseAndValidateInput(context *gin.Context) (Parameters, error) {
	p := Parameters{}
	if p.JobName = strings.TrimPrefix(context.Param("job"), "/"); p.JobName == "" {
		return p, errors.New("missing Job name in URL")
	}
	if err := context.ShouldBindQuery(&p); err != nil {
		return p, err
	}
	SetLogWithFields(context, log.DebugLevel, "Validated Input Data", log.Fields{
		"parameters": p,
	})
	return p, nil
}
