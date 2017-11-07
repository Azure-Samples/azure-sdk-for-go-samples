package management

import (
	"github.com/joshgav/az-go/common"
	"github.com/subosito/gotenv"
)

var (
	subscriptionId    string
	tenantId          string
	resourceGroupName string
	location          string
)

func init() {
	gotenv.Load() // read from .env file

	subscriptionId = common.GetEnvVarOrFail("AZURE_SUBSCRIPTION_ID")
	tenantId = common.GetEnvVarOrFail("AZURE_TENANT_ID")
	resourceGroupName = common.GetEnvVarOrFail("AZURE_RG_NAME")
	location = common.GetEnvVarOrFail("AZURE_LOCATION")
}
