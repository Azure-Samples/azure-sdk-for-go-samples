package communication

import (
	"context"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/preview/communication/mgmt/2020-08-20-preview/communication"
	"github.com/Azure/go-autorest/autorest/to"
)

//Create a CommunicationServiceManagementClient object using a Subscription ID
func GetManagementServiceClient() communication.ServiceClient {
	serviceClient := communication.NewServiceClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	serviceClient.Authorizer = a
	serviceClient.AddToUserAgent(config.UserAgent())
	return serviceClient
}

func GetOperationsStatusesClient() communication.OperationStatusesClient {
	operationsClient := communication.NewOperationStatusesClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	operationsClient.Authorizer = a
	operationsClient.AddToUserAgent(config.UserAgent())
	return operationsClient
}

//Create a ACS instance
func CreateCommunicationService(ctx context.Context, resourceGroupName string, serviceName string) (service communication.ServiceResource, err error) {
	client := GetManagementServiceClient()
	var serviceResource = communication.ServiceResource{
		Location: to.StringPtr("global"),
		ServiceProperties: &communication.ServiceProperties{
			DataLocation: to.StringPtr("UnitedStates"),
		},
	}
	future, err := client.CreateOrUpdate(ctx, resourceGroupName, serviceName, &serviceResource)
	if err != nil {
		return service, err
	}
	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return service, err
	}
	return future.Result(client)
}

//Delete an ACS instance
func DeleteCommunicationServices(ctx context.Context, resourceGroupName string, resourceName string) error {
	client := GetManagementServiceClient()
	future, err := client.Delete(ctx, resourceGroupName, resourceName)
	if err != nil {
		return err
	}
	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return err
	}
	return nil
}

//List all ACS instances
func ListCommunicationServices(ctx context.Context) (communication.ServiceResourceListIterator, error) {
	client := GetManagementServiceClient()
	return client.ListBySubscriptionComplete(ctx)
}

//Get status of all operation
func GetOperationStatus(ctx context.Context, location string, operationID string) (communication.OperationStatus, error) {
	operationsClient := GetOperationsStatusesClient()
	return operationsClient.Get(ctx, location, operationID)
}

//Regenerate key of ACS instance
func RegenerateKeys(ctx context.Context, resourceGroupName string, communicationServiceName string) (communication.ServiceKeys, error) {
	client := GetManagementServiceClient()
	communicationKey := communication.RegenerateKeyParameters{
		KeyType: communication.Primary,
	}
	return client.RegenerateKey(ctx, resourceGroupName, communicationServiceName, &communicationKey)
}

//List keys of ACS instance
func ListKeys(ctx context.Context, resourceGroupName string, communicationServiceName string) (communication.ServiceKeys, error) {
	client := GetManagementServiceClient()
	return client.ListKeys(ctx, resourceGroupName, communicationServiceName)
}

//Get resources
func GetCommunicationService(ctx context.Context, resourceGroupName string, resourceName string) (communication.ServiceResource, error) {
	client := GetManagementServiceClient()
	return client.Get(ctx, resourceGroupName, resourceName)
}

//Update ACS instance tag
func UpdateCommunicationService(ctx context.Context, resourceGroupName string, communicationServiceName string, tags map[string]*string) (communication.ServiceResource, error) {
	client := GetManagementServiceClient()
	taggedResource := communication.TaggedResource{
		Tags: tags,
	}
	return client.Update(ctx, resourceGroupName, communicationServiceName, &taggedResource)
}

//List all communication services in resource group
func ListCommunicationServicesByResourceGroupName(ctx context.Context, resourceGroupName string) (communication.ServiceResourceListIterator, error) {
	serviceClient := GetManagementServiceClient()
	return serviceClient.ListByResourceGroupComplete(ctx, resourceGroupName)
}
