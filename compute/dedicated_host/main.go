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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	TenantID          string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	hostGroupName     = "sample-host-group"
	hostName          = "sample-host"
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	TenantID = os.Getenv("AZURE_TENANT_ID")
	if len(TenantID) == 0 {
		log.Fatal("AZURE_TENANT_ID is not set.")
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

	dedicatedHostGroup, err := createDedicatedHostGroups(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("dedicated host group:", *dedicatedHostGroup.ID)

	dedicatedHost, err := createDedicatedHost(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("dedicated host:", *dedicatedHost.ID)

	dedicatedHostGroup, err = getDedicatedHostGroups(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get dedicated host:", *dedicatedHost.ID)

	dedicatedHost, err = getDedicatedHost(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get dedicated host:", *dedicatedHost.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createDedicatedHostGroups(ctx context.Context, conn *arm.Connection) (*armcompute.DedicatedHostGroup, error) {
	dedicatedHostGroupsClient := armcompute.NewDedicatedHostGroupsClient(conn, subscriptionID)

	resp, err := dedicatedHostGroupsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		hostGroupName,
		armcompute.DedicatedHostGroup{
			Resource: armcompute.Resource{
				Location: to.StringPtr("eastus"),
			},
			Properties: &armcompute.DedicatedHostGroupProperties{
				PlatformFaultDomainCount: to.Int32Ptr(3),
			},
			Zones: []*string{to.StringPtr("1")},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.DedicatedHostGroup, nil
}

func getDedicatedHostGroups(ctx context.Context, conn *arm.Connection) (*armcompute.DedicatedHostGroup, error) {
	dedicatedHostGroupsClient := armcompute.NewDedicatedHostGroupsClient(conn, subscriptionID)

	resp, err := dedicatedHostGroupsClient.Get(ctx, resourceGroupName, hostGroupName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.DedicatedHostGroup, nil
}

func createDedicatedHost(ctx context.Context, conn *arm.Connection) (*armcompute.DedicatedHost, error) {
	dedicatedHostClient := armcompute.NewDedicatedHostsClient(conn, subscriptionID)

	pollerResp, err := dedicatedHostClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		hostGroupName,
		hostName,
		armcompute.DedicatedHost{
			Resource: armcompute.Resource{
				Location: to.StringPtr("eastus"),
			},
			Properties: &armcompute.DedicatedHostProperties{
				PlatformFaultDomain: to.Int32Ptr(1),
			},
			SKU: &armcompute.SKU{
				Name: to.StringPtr("DSv3-Type1"),
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
	return &resp.DedicatedHost, nil
}

func getDedicatedHost(ctx context.Context, conn *arm.Connection) (*armcompute.DedicatedHost, error) {
	dedicatedHostClient := armcompute.NewDedicatedHostsClient(conn, subscriptionID)

	resp, err := dedicatedHostClient.Get(ctx, resourceGroupName, hostGroupName, hostName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.DedicatedHost, nil
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
