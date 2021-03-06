package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"os"
	"strings"
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

	if log.GetLevel() != log.DebugLevel {
		gin.SetMode(gin.ReleaseMode)
	}

	for _, job := range cfg.Jobs.Register {
		arr := strings.Split(job, "@")
		j := Job{JobName: arr[0]}
		if len(arr) >= 2 {
			j.TargetHost = arr[1]
		}
		if err := j.RegisterMetric(); err != nil {
			log.WithField("job", job).WithError(err).Warn("Failed to register job.")
		} else {
			log.WithField("job", job).Info("Registered job.")
		}
	}

	log.WithField("port", cfg.BindAddr).Info("Starting webserver.")
	r := SetupRouter()
	err := r.Run(cfg.BindAddr)
	log.WithError(err).Fatal("Shutting down.")
}

// SetupRouter initializes Gin with the handlers.
func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(
		LogrusHandler(),
		ErrorHandle(),
		InputValidationHandle("/pre", "/post", "/register", "/unregister"),
		gin.Recovery(),
	)
	r.GET("/", handleRoot)
	r.GET("/presnap/*job", handlePreSnap)
	r.GET("/postsnap/*job", handlePostSnap)
	r.GET("/presend/*job", handlePreSend)
	r.GET("/postsend/*job", handlePostSend)
	r.GET("/register/*job", handleRegister)
	r.GET("/unregister/*job", handleUnregister)
	r.GET("/health/ready", handleHealthcheck)
	r.GET("/health/alive", handleHealthcheck)
	r.GET("/metrics", handleMetrics)
	return r
}
