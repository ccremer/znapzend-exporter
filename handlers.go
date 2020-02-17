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

func handlePreSnap(context *gin.Context) {
	parameters, err := ParseAndValidateInput(context)
	if err != nil {
		SetLog(context, log.WarnLevel, err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	SetLogWithFields(context, log.DebugLevel, "Input Data", log.Fields{
		"parameters": parameters,
	})

}

func handlePostSnap(context *gin.Context) {
	parameters, err := ParseAndValidateInput(context)
	if err != nil {
		SetLog(context, log.ErrorLevel, err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	SetLogWithFields(context, log.DebugLevel, "Input Data", log.Fields{
		"parameters": parameters,
	})

}

func handleMetrics(context *gin.Context) {
	SetLogLevel(context, log.DebugLevel)
	promhttp.Handler().ServeHTTP(context.Writer, context.Request)
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
	return p, nil
}
