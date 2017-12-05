package helpers

import (
	"log"
	"os"

	"github.com/subosito/gotenv"
)

var (
	SubscriptionID    string
	TenantID          string
	ResourceGroupName string
	Location          string
)

func init() {
	gotenv.Load() // read from .env file

	SubscriptionID = GetEnvVarOrFail("AZURE_SUBSCRIPTION_ID")
	TenantID = GetEnvVarOrFail("AZURE_TENANT_ID")
	ResourceGroupName = GetEnvVarOrFail("AZURE_RG_NAME")
	Location = GetEnvVarOrFail("AZURE_LOCATION")
}

func OnErrorFail(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}

func GetEnvVarOrFail(envVar string) string {
	value := os.Getenv(envVar)
	if value == "" {
		log.Fatalf("envVar %s must be specified", envVar)
	}

	return value
}
