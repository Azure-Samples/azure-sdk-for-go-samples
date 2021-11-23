package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID      string
	location            = "westus"
	resourceGroupName   = "sample-resource-group"
	availabilitySetName = "sample-availability-sets"
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
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	availabilitySets, err := createAvailabilitySet(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("availability set:", *availabilitySets.ID)

	availabilitySetList := listAvailabilitySet(ctx, cred)
	for i, a := range availabilitySetList {
		log.Println(i, *a.ID)
	}

	availabilitySetSizes, err := listAvailabilitySizes(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("list availability size:", len(availabilitySetSizes.Value))
	for i, v := range availabilitySetSizes.Value {
		log.Println(i, *v.Name)
	}

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createAvailabilitySet(ctx context.Context, cred azcore.TokenCredential) (*armcompute.AvailabilitySet, error) {
	availabilitySetsClient := armcompute.NewAvailabilitySetsClient(subscriptionID, cred, nil)

	resp, err := availabilitySetsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		availabilitySetName,
		armcompute.AvailabilitySet{
			Resource: armcompute.Resource{
				Location: to.StringPtr(location),
			},
			Properties: &armcompute.AvailabilitySetProperties{
				PlatformFaultDomainCount:  to.Int32Ptr(1),
				PlatformUpdateDomainCount: to.Int32Ptr(1),
			},
			SKU: &armcompute.SKU{
				Name: to.StringPtr("Aligned"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.AvailabilitySet, nil
}

func listAvailabilitySet(ctx context.Context, cred azcore.TokenCredential) []*armcompute.AvailabilitySet {
	availabilitySetsClient := armcompute.NewAvailabilitySetsClient(subscriptionID, cred, nil)

	availability := availabilitySetsClient.List(resourceGroupName, nil)

	availabilitySet := make([]*armcompute.AvailabilitySet, 0)
	for availability.NextPage(ctx) {
		resp := availability.PageResponse()
		availabilitySet = append(availabilitySet, resp.AvailabilitySetListResult.Value...)
	}

	return availabilitySet
}

func listAvailabilitySizes(ctx context.Context, cred azcore.TokenCredential) (*armcompute.VirtualMachineSizeListResult, error) {
	availabilitySetsClient := armcompute.NewAvailabilitySetsClient(subscriptionID, cred, nil)

	availability, err := availabilitySetsClient.ListAvailableSizes(ctx, resourceGroupName, availabilitySetName, nil)
	if err != nil {
		return nil, err
	}

	return &availability.VirtualMachineSizeListResult, nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

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

func cleanup(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

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
