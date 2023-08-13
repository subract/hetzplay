package main

import (
	"os"
	"strings"

	"github.com/op/go-logging"
)

func initializeLogging(level string) (err error) {
	// Parse log level
	logLevel, err := logging.LogLevel(strings.ToUpper(level))
	if err != nil {
		return
	}

	// Define format
	logFormat := logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{level}%{color:reset} %{message}`,
	)
	// Initialize logger backend
	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	// Make logs pretty
	logBackendFormatter := logging.NewBackendFormatter(logBackend, logFormat)
	// Set log level
	logBackendLeveled := logging.AddModuleLevel(logBackendFormatter)
	logBackendLeveled.SetLevel(logLevel, "")
	// Apply settings to logger
	logging.SetBackend(logBackendLeveled)
	return
}
