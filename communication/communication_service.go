package communication

import (
	"context"
	"fmt"
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
	serviceClient := GetManagementServiceClient()
	var serviceResource = communication.ServiceResource{
		Location: to.StringPtr("global"),
		ServiceProperties: &communication.ServiceProperties{
			DataLocation: to.StringPtr("UnitedStates"),
		},
	}
	future, err := serviceClient.CreateOrUpdate(ctx, resourceGroupName, serviceName, &serviceResource)
	if err != nil {
		return service, err
	}
	if err := future.WaitForCompletionRef(ctx, serviceClient.Client); err != nil {
		return service, err
	}
	return future.Result(serviceClient)
}

//Delete an ACS instance
func DeleteCommunicationServices(ctx context.Context, resourceGroupName string, resourceName string, serviceClient communication.ServiceClient) error {
	future, err := serviceClient.Delete(ctx, resourceGroupName, resourceName)
	if err != nil {
		return err
	}
	if err := future.WaitForCompletionRef(ctx, serviceClient.Client); err != nil {
		return err
	}
	return nil
}

//List all ACS instances
func ListCommunicationServices(ctx context.Context, serviceClient communication.ServiceClient) (resourceListPage communication.ServiceResourceListPage, err error) {
	r, err := serviceClient.ListBySubscription(ctx)
	if err != nil {
		var emptyPage communication.ServiceResourceListPage
		return emptyPage, fmt.Errorf("Cannot list subscriptions")
	}
	return r, err
}

//Get status of all operation
func GetOperationStatus(ctx context.Context, location string, operationId string, operationsClient communication.OperationStatusesClient) {
	resp, err := operationsClient.Get(ctx, location, operationId)
	if err == nil {
		fmt.Println("Operation" + operationId + "ID is" + *resp.ID)
	} else {
		fmt.Println("Failed to get operation status")
	}
}

//Regenerate key of ACS instance
func RegenerateKeys(resourceGroupName string, communicationServiceName string, serviceClient communication.ServiceClient) {
	var communicationKey = communication.RegenerateKeyParameters{"primary"}
	resp, err := serviceClient.RegenerateKey(context.TODO(), resourceGroupName, communicationServiceName, &communicationKey)
	if err == nil {
		fmt.Println("Regenerated Key" + *resp.PrimaryKey)
	}
}

//List keys of ACS instance
func ListKeys(resourceGroupName string, communicationServiceName string, serviceClient communication.ServiceClient) {
	resp, err := serviceClient.ListKeys(context.TODO(), resourceGroupName, communicationServiceName)
	if err == nil {
		fmt.Println("Primary Key" + *resp.PrimaryKey)
	} else {
		fmt.Println("Failed to get keys")
	}
}

//Get resources
func GetResourceAsync(resourceGroupName string, resourceName string, serviceClient communication.ServiceClient) {
	res, err := serviceClient.Get(context.TODO(), resourceGroupName, resourceName)
	if err == nil {
		fmt.Println("Successfully got resource" + *res.Name)
	} else {
		fmt.Println("Failed to get resource")
	}

}

//Update ACS instance tag
func UpdateCommunicationService(resourceGroupName string, communicationServiceName string, newTag string, newValue string, serviceClient communication.ServiceClient) {
	m := make(map[string]*string)
	m[newTag] = &newValue
	var taggedResource = communication.TaggedResource{m}
	resp, err := serviceClient.Update(context.TODO(), resourceGroupName, communicationServiceName, &taggedResource)
	if err == nil {
		fmt.Println("Service" + *resp.Name + "updated correctly")
	}
}

//List all communication services in resource group
func ListCommunicationServicesByResourceGroupName(resourceGroupName string, serviceClient communication.ServiceClient) {
	resp, err := serviceClient.ListByResourceGroup(context.TODO(), resourceGroupName)
	if err == nil {
		for _, resource := range resp.Values() {
			fmt.Println("Name: ", *resource.Name)
			fmt.Println("Provisioning state:", resource.ServiceProperties.ProvisioningState)
			fmt.Println("ImmutableResourceId", *resource.ServiceProperties.ImmutableResourceID)
		}
	}
}
