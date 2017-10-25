package util

import (
  "log"
  "os"
)

func OnErrorFail(err error, message string) {
  if err != nil { log.Fatalf("%s: %v", message, err) }
}

func GetEnvVarOrFail(envVar string) string {
	value := os.Getenv(envVar)
	if value == "" {
		log.Fatalf("envVar %s must be specified", envVar)
	}

	return value
}

