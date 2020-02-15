package main

import (
	"github.com/sirupsen/logrus"
)

var (
	// These will be populated by Goreleaser
	version string
	commit  string
	date    string
)

func main() {
	logrus.WithFields(logrus.Fields{
		"version": version,
		"commit": commit,
		"date": date,
	}).Info("Starting Znapzend exporter")
}
