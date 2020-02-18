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
	parameters, err := ParseAndValidateInput(context)
	if err != nil {
		return
	}
	SetMetric(preSnapMetric, &parameters, parameters.ResetPreSnap)

}

func handlePostSnap(context *gin.Context) {
	parameters, err := ParseAndValidateInput(context)
	if err != nil {
		return
	}
	SetMetric(preSnapMetric, &parameters, parameters.ResetPostSnap)

}

func handleMetrics(context *gin.Context) {
	SetLogLevel(context, log.DebugLevel)
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
		err := errors.New("missing Job name in URL")
		setErrorForRequest(context, err)
		return p, err
	}
	if err := context.ShouldBindQuery(&p); err != nil {
		setErrorForRequest(context, err)
	}
	SetLogWithFields(context, log.DebugLevel, "Validated Input Data", log.Fields{
		"parameters": p,
	})
	return p, nil
}

func setErrorForRequest(context *gin.Context, err error) {
	SetLog(context, log.ErrorLevel, err.Error())
	context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
