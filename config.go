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

func CreateDefaultConfig() ConfigMap {
	return ConfigMap{
		Log: LogMap{
			Level: "info",
		},
		BindAddr: ":8080",
	}
}

func setupFlags() {
	cfg := CreateDefaultConfig()

	flag.String("bindAddr", cfg.BindAddr, "IP Address to bind to listen for Prometheus scrapes")
	flag.String("log.level", cfg.Log.Level, "Logging level")

	if err := viper.BindPFlags(flag.CommandLine); err != nil {
		log.Fatal(err)
	}
}

func LoadConfig() error {
	flag.Parse()

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	defaults := CreateDefaultConfig()
	ser, err := json.Marshal(defaults)
	if err == nil {
		viper.SetConfigType("ser")
		return viper.ReadConfig(bytes.NewBuffer(ser))
	} else {
		return err
	}
}

func GetConfig() ConfigMap {
	cfg := CreateDefaultConfig()
	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

type (
	ConfigMap struct {
		Log      LogMap
		BindAddr string
	}
	LogMap struct {
		Level     string
		Formatter string
	}
)

func LogrusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Keys = make(map[string]interface{})
		startTime := time.Now()

		// Processing request
		c.Next()

		logLevel := GetOrDefault(c.Keys, "log_level", log.InfoLevel).(log.Level)
		if c.Keys["log_skip"] == true {
			return
		}
		message := GetOrDefault(c.Keys, "log_message", "").(string)

		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		reqMethod := c.Request.Method
		reqUri := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		log.WithFields(filterAndCombineLoggingKeys(log.Fields{
			"status_code":  statusCode,
			"latency_time": latencyTime,
			"client_ip":    clientIP,
			"req_method":   reqMethod,
			"req_uri":      reqUri,
		}, c.Keys)).Log(logLevel, message)
	}
}

func GetOrDefault(keys map[string]interface{}, key string, defaultValue interface{}) interface{} {
	var value = keys[key]
	if value == nil {
		return defaultValue
	} else {
		return value
	}
}

func filterAndCombineLoggingKeys(fields log.Fields, keys map[string]interface{}) log.Fields {
	for key, _ := range keys {
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

func SetLogLevel(c *gin.Context, level log.Level) {
	c.Keys["log_level"] = level
}

func SetLog(c *gin.Context, level log.Level, message string) {
	c.Keys["log_level"] = level
	if message != "" {
		c.Keys["log_message"] = message
	}
}

func SetLogWithFields(c *gin.Context, level log.Level, message string, fields log.Fields) {
	c.Keys["log_level"] = level
	if message != "" {
		c.Keys["log_message"] = message
	}
	for key, _ := range fields {
		c.Keys[key] = fields[key]
	}
}