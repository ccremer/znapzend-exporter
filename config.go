package main

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"strings"

	"github.com/spf13/viper"
	"os"
	"time"
)

func init() {
	setupFlags()
}

// SetupLogging initializes logging framework
func SetupLogging() {

	cfg := GetConfig()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	level, err := log.ParseLevel(cfg.Log.Level)
	if err != nil {
		log.WithField("error", err).Warn("Using info level.")
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
	}
}

// CreateDefaultConfig creates a config map with the hardcoded internal defaults.
func CreateDefaultConfig() ConfigMap {
	return ConfigMap{
		Log: LogMap{
			Level: "info",
		},
		BindAddr: ":8080",
		Jobs:     JobMap{},
	}
}

func setupFlags() {
	cfg := CreateDefaultConfig()

	flag.String("bindAddr", cfg.BindAddr, "IP Address to bind to listen for Prometheus scrapes")
	flag.String("log.level", cfg.Log.Level, "Logging level")
	flag.StringSlice("jobs.register", []string{}, "A list of job labels to register at startup. Can be specified multiple times")

	if err := viper.BindPFlags(flag.CommandLine); err != nil {
		log.Fatal(err)
	}
}

// LoadConfig loads the configuration from Environment Variables and CLI flags and merge with hardcoded internal defaults.
func LoadConfig() error {
	flag.Parse()

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	defaults := CreateDefaultConfig()
	ser, err := json.Marshal(defaults)
	if err == nil {
		viper.SetConfigType("ser")
		return viper.ReadConfig(bytes.NewBuffer(ser))
	}
	return err
}

// GetConfig gets the parsed, final configuration.
func GetConfig() ConfigMap {
	cfg := CreateDefaultConfig()
	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

type (
	// ConfigMap is the root config map
	ConfigMap struct {
		Log      LogMap
		BindAddr string
		Jobs     JobMap
	}
	// LogMap contains config for logging
	LogMap struct {
		Level     string
		Formatter string
	}
	// JobMap contains values for prometheus "jobs"
	JobMap struct {
		Register []string
	}
)

// LogrusHandler implements a Gin HandlerFunc that logs the request with logrus instead of Gin builtin logger.
func LogrusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Processing request
		c.Next()

		logLevel := GetOrDefault(c.Keys, "log_level", log.InfoLevel).(log.Level)
		if c.GetBool("log_skip") {
			return
		}
		message := GetOrDefault(c.Keys, "log_message", "").(string)

		log.WithFields(filterAndCombineLoggingKeys(log.Fields{
			"status_code":  c.Writer.Status(),
			"latency_time": time.Now().Sub(startTime),
			"client_ip":    c.ClientIP(),
			"req_method":   c.Request.Method,
			"req_uri":      c.Request.RequestURI,
		}, c.Keys)).Log(logLevel, message)
	}
}

// GetOrDefault gets the value of the given map by key. If the value is nil or not found, defaultValue is returned.
func GetOrDefault(keys map[string]interface{}, key string, defaultValue interface{}) interface{} {
	var value = keys[key]
	if value == nil {
		return defaultValue
	}
	return value
}

func filterAndCombineLoggingKeys(fields log.Fields, keys map[string]interface{}) log.Fields {
	for key := range keys {
		switch key {
		case "log_message":
		case "log_level":
		case "log_skip":
			break
		default:
			fields[key] = keys[key]
		}
	}
	return fields
}

// SetLogLevel sets the log level of the Gin HTTP context (takes effect when being logged)
func SetLogLevel(c *gin.Context, level log.Level) {
	SetLogWithFields(c, level, "", log.Fields{})
}

// SetLog sets the log level and message of the Gin HTTP context (takes effect when being logged)
func SetLog(c *gin.Context, level log.Level, message string) {
	SetLogWithFields(c, level, message, log.Fields{})
}

// SetLogWithFields sets the log level, message and fields of the Gin HTTP context (takes effect when being logged)
func SetLogWithFields(c *gin.Context, level log.Level, message string, fields log.Fields) {
	c.Set("log_level", level)
	if message != "" {
		c.Set("log_message", message)
	}
	for key := range fields {
		c.Set(key, fields[key])
	}
}

// SetError sets the error log level, message and fields of the Gin HTTP context (takes effect when being logged)
func SetError(c *gin.Context, message string, err error, fields log.Fields) {
	SetLogWithFields(c, log.ErrorLevel, message, fields)
	c.Set("error", err)
}
