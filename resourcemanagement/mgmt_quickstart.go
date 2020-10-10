package main


import (
	"fmt"
	"net/http"
	communication "github.com/Azure/azure-sdk-for-go/services/preview/communication/mgmt/2020-08-20-preview/communication"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"context"
)

var operationsClient communication.OperationStatusesClient
var serviceClient communication.ServiceClient;
func main() {
}


//Create a CommunicationServiceManagementClient object using a Subscription ID
func GetManagementServiceClient() communication.ServiceClient {
	serviceClient = communication.NewServiceClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	serviceClient.Authorizer = a
	serviceClient.AddToUserAgent(config.UserAgent())
	return serviceClient
}

func GetOperationsStatusesClient() communication.OperationStatusesClient{
	operationsClient = communication.NewOperationStatusesClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	operationsClient.Authorizer =  a
	operationsClient.AddToUserAgent(config.UserAgent())
	return operationsClient
}

//List all ACS instances
func ListCommunicationServices(serviceClient communication.ServiceClient){
	var response, error = serviceClient.ListBySubscription(context.TODO())
	if(error!=nil){
		for _,resource := range response.Values(){
			fmt.Println("Name: ",  *resource.Name);
			fmt.Println("Provisioning state:", resource.ServiceProperties.ProvisioningState)
			fmt.Println("ImmutableResourceId", *resource.ServiceProperties.ImmutableResourceID)
		}
	}
}

func ListCommunicationServices2(ctx context.Context,serviceClient communication.ServiceClient) (resourceListPage communication.ServiceResourceListPage, err error){
		future, err := serviceClient.ListBySubscription(ctx)
		if err != nil {
			var emptyPage communication.ServiceResourceListPage
			return emptyPage, fmt.Errorf("Cannot list subscriptions")
		}
		return future, err
}
//Delete an ACS instance
func DeleteCommunicationServices(resourceGroupName string, resourceName string, serviceClient communication.ServiceClient) {
	resp, err := serviceClient.Delete(context.TODO(), resourceGroupName, resourceName)
	if(err==nil){
		fmt.Println("Resource successfully deleted");
		fmt.Println(resp)
	} else {
		fmt.Println("Resource Failed to delete")
	} 
}

func DeleteCommunicationServices2(ctx context.Context, resourceGroupName string, resourceName string, serviceClient communication.ServiceClient) (response *http.Response, err error){
	resp, err := serviceClient.Delete(ctx, resourceGroupName, resourceName)
	
	if err != nil {
		return response, fmt.Errorf("Failed to delete service")
	}

	err = resp.WaitForCompletionRef(ctx, serviceClient.Client)
	if err != nil {
		return response, fmt.Errorf("Failed to delete service")
	}

	return resp.Response(), err
}

//Get status of all operation
func GetOperationStatus(location string, operationId string, serviceClient communication.ServiceClient) {
	resp, err := operationsClient.Get(context.TODO(), location, operationId)
	if(err==nil){
		fmt.Println("Operation" + operationId + "ID is" + *resp.ID)
	} else {
		fmt.Println("Failed to get operation status")
	}
}


//Regenerate key of ACS instance
func RegenerateKeys(resourceGroupName string, communicationServiceName string, serviceClient communication.ServiceClient){
	var communicationKey = communication.RegenerateKeyParameters {"primary"}
	resp, err := serviceClient.RegenerateKey(context.TODO(), resourceGroupName, communicationServiceName, &communicationKey)
	if(err==nil){
		fmt.Println("Regenerated Key" + *resp.PrimaryKey)
	}
}

//List keys of ACS instance
func ListKeys(resourceGroupName string, communicationServiceName string, serviceClient communication.ServiceClient) {
	resp, err := serviceClient.ListKeys(context.TODO(), resourceGroupName, communicationServiceName)
	if(err==nil){
		fmt.Println("Primary Key" + *resp.PrimaryKey)
	} else {
		fmt.Println("Failed to get keys")
	}
}

//Get resources
func GetResourceAsync(resourceGroupName string, resourceName string, serviceClient communication.ServiceClient) {
	res, err := serviceClient.Get(context.TODO(), resourceGroupName, resourceName)
	if(err==nil){
		fmt.Println("Successfully got resource" + *res.Name)
	} else {
		fmt.Println("Failed to get resource")
	}
	
}

//Update ACS instance tag
func UpdateCommunicationService(resourceGroupName string, communicationServiceName string, newTag string, newValue string, serviceClient communication.ServiceClient){
	m := make(map[string]*string)
	m[newTag] = &newValue;
	var taggedResource = communication.TaggedResource{ m } 
	resp, err := serviceClient.Update(context.TODO(), resourceGroupName, communicationServiceName, &taggedResource)
	if(err==nil){
		fmt.Println("Service" + *resp.Name + "updated correctly")
	}
}

//List all communication services in resource group
func ListCommunicationServicesByResourceGroupName(resourceGroupName string, serviceClient communication.ServiceClient) {
	resp, err := serviceClient.ListByResourceGroup(context.TODO(), resourceGroupName)
	if(err==nil){
		for _,resource := range resp.Values(){
			fmt.Println("Name: ",  *resource.Name);
			fmt.Println("Provisioning state:", resource.ServiceProperties.ProvisioningState)
			fmt.Println("ImmutableResourceId", *resource.ServiceProperties.ImmutableResourceID)
		}
	}
}


//Create a ACS instance
func CreateCommunicationService(resourceGroupName string, resourceName string, serviceClient communication.ServiceClient){
	var serviceResource =  communication.ServiceResource {  }
	var serviceProperties = communication.ServiceProperties { }
	var dataLocationPtr = "UnitedStates"
	var locationPtr = "global"
	serviceProperties.DataLocation = &dataLocationPtr
	serviceResource.Location = &locationPtr
	serviceResource.ServiceProperties = &serviceProperties
	resp, err := serviceClient.CreateOrUpdate(context.TODO(), resourceGroupName, resourceName, &serviceResource)
	if(err==nil){
		fmt.Println("Resource successfully created")
		fmt.Println(resp)
	} else{
		fmt.Println("Resource failed to create")
	}
}
