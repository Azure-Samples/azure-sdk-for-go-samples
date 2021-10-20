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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID       string
	location             = "westus"
	resourceGroupName    = "sample-resource-group"
	partnerNamespaceName = "sample-partner-namespace"
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

	partnerNamespace, err := createPartnerNamespace(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("partner namespace:", *partnerNamespace.ID)

	partnerNamespace, err = getPartnerNamespace(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get partner namespace:", *partnerNamespace.ID)

	partnerNamespaces := listPartnerNamespace(ctx, conn)
	for _, p := range partnerNamespaces {
		log.Println(*p.Name, *p.ID)
	}

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createPartnerNamespace(ctx context.Context, conn *arm.Connection) (*armeventgrid.PartnerNamespace, error) {
	partnerNamespacesClient := armeventgrid.NewPartnerNamespacesClient(conn, subscriptionID)

	pollerResp, err := partnerNamespacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		partnerNamespaceName,
		armeventgrid.PartnerNamespace{
			TrackedResource: armeventgrid.TrackedResource{
				Location: to.StringPtr(location),
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
	return &resp.PartnerNamespace, nil
}

func getPartnerNamespace(ctx context.Context, conn *arm.Connection) (*armeventgrid.PartnerNamespace, error) {
	partnerNamespacesClient := armeventgrid.NewPartnerNamespacesClient(conn, subscriptionID)

	resp, err := partnerNamespacesClient.Get(ctx, resourceGroupName, partnerNamespaceName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.PartnerNamespace, nil
}

func listPartnerNamespace(ctx context.Context, conn *arm.Connection) []*armeventgrid.PartnerNamespace {
	partnerNamespacesClient := armeventgrid.NewPartnerNamespacesClient(conn, subscriptionID)

	pager := partnerNamespacesClient.ListBySubscription(nil)

	partnerNamespaces := make([]*armeventgrid.PartnerNamespace, 0)
	for pager.NextPage(ctx) {
		resp := pager.PageResponse()
		partnerNamespaces = append(partnerNamespaces, resp.Value...)
	}
	return partnerNamespaces
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
