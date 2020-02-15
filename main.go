package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"net/http"
	"os"
)

var (
	// These will be populated by Goreleaser
	version string
	commit  string
	date    string

	helpText = `%s (version %s, %s, %s)

All flags can be read from Environment variables as well (replace . with _ , e.g. LOG_LEVEL).
However, CLI flags take precedence.

`
)

const (
	delayResetParam = "delayResetBy"
)

func main() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, helpText, os.Args[0], version, commit, date)
		flag.PrintDefaults()
	}
	if err := LoadConfig(); err != nil {
		log.WithError(err).Error("Could not load config.")
	}
	SetupLogging()

	cfg := GetConfig()

	log.WithFields(log.Fields{
		"version": version,
		"commit":  commit,
		"date":    date,
	}).Info("Starting Znapzend exporter")

	r := gin.New()
	r.Use(
		LogrusHandler(),
		gin.Recovery(),
	)
	r.GET("/health/ready", handleReadiness)
	r.GET("/health/alive", handleLiveness)
	r.GET("/metrics", handleMetrics)

	err := r.Run(cfg.BindAddr)
	log.WithError(err).Fatal("Shutting down.")
}

func handleMetrics(context *gin.Context) {
	SetLogLevel(context, log.DebugLevel)
	promhttp.Handler().ServeHTTP(context.Writer, context.Request)
}

func handleReadiness(context *gin.Context) {
	SetLogLevel(context, log.DebugLevel)
	context.JSON(http.StatusOK, struct{ Message string }{
		Message: "Webserver up and ready.",
	})
}

func handleLiveness(context *gin.Context) {
	SetLogLevel(context, log.DebugLevel)
	context.JSON(http.StatusOK, struct{ Message string }{
		Message: "If you can reach this, I'm alive!",
	})
}
