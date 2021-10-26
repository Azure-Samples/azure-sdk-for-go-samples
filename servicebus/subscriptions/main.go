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
	topicName         = "sample-sb-topic"
	subscriptionName  = "sample-subscription"
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

	topic, err := createTopic(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus topic:", *topic.ID)

	subscription, err := createSubscription(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus subscription:", *subscription.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createNamespace(ctx context.Context, conn *arm.Connection) (*armservicebus.SBNamespace, error) {
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
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.SBNamespace, nil
}

func createTopic(ctx context.Context, conn *arm.Connection) (*armservicebus.SBTopic, error) {
	topicsClient := armservicebus.NewTopicsClient(conn, subscriptionID)

	resp, err := topicsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		namespaceName,
		topicName,
		armservicebus.SBTopic{
			Properties: &armservicebus.SBTopicProperties{},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.SBTopic, nil
}

func createSubscription(ctx context.Context, conn *arm.Connection) (*armservicebus.SBSubscription, error) {
	subscriptionsClient := armservicebus.NewSubscriptionsClient(conn, subscriptionID)

	resp, err := subscriptionsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		namespaceName,
		topicName,
		subscriptionName,
		armservicebus.SBSubscription{},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.SBSubscription, nil
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
