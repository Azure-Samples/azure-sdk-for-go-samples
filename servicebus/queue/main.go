package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	namespaceName     = "sample-sb-namespace"
	queueName         = "sample-sb-queue"
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}

	conn := arm.NewDefaultConnection(cred, &arm.ConnectionOptions{
		Logging: policy.LogOptions{
			IncludeBody: true,
		},
	})
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	namespace, err := createNamespace(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus namespace:", *namespace.ID)

	queue, err := createQueue(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus queue:", *queue.ID)

	queueGet, err := getQueue(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get service bus queue:", *queueGet.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createNamespace(ctx context.Context, conn *arm.Connection) (np armservicebus.NamespacesCreateOrUpdateResponse, err error) {
	namespacesClient := armservicebus.NewNamespacesClient(conn, subscriptionID)

	pollerResp, err := namespacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		namespaceName,
		armservicebus.SBNamespace{
			TrackedResource: armservicebus.TrackedResource{
				Location: to.StringPtr(location),
			},
			SKU: &armservicebus.SBSKU{
				Name: armservicebus.SKUNamePremium.ToPtr(),
				Tier: armservicebus.SKUTierPremium.ToPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return np, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return np, err
	}
	return resp, nil
}

func createQueue(ctx context.Context, conn *arm.Connection) (q armservicebus.QueuesCreateOrUpdateResponse, err error) {
	queuesClient := armservicebus.NewQueuesClient(conn, subscriptionID)

	resp, err := queuesClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		namespaceName,
		queueName,
		armservicebus.SBQueue{
			Properties: &armservicebus.SBQueueProperties{
				EnablePartitioning: to.BoolPtr(true),
			},
		},
		nil,
	)
	if err != nil {
		return q, nil
	}

	return resp, nil
}

func getQueue(ctx context.Context, conn *arm.Connection) (q armservicebus.QueuesGetResponse, err error) {
	queuesClient := armservicebus.NewQueuesClient(conn, subscriptionID)

	resp, err := queuesClient.Get(ctx, resourceGroupName, namespaceName, queueName, nil)
	if err != nil {
		return q, err
	}
	return resp, nil
}

func createResourceGroup(ctx context.Context, conn *arm.Connection) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.StringPtr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}
