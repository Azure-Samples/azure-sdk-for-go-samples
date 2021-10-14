module azuresample/sql/serverkey

go 1.16

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v0.19.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v0.10.0
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/armkeyvault v0.2.0
	github.com/Azure/azure-sdk-for-go/sdk/resources/armresources v0.3.0
	github.com/Azure/azure-sdk-for-go/sdk/sql/armsql v0.1.0
)

replace github.com/Azure/azure-sdk-for-go v57.1.0+incompatible => github.com/Azure/azure-sdk-for-go v57.3.0+incompatible
