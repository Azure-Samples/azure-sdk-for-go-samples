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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID             string
	location                   = "westus"
	resourceGroupName          = "sample-resource-group"
	namespacesName             = "sample1namespace"
	secondNamespacesName       = "sample1second1namespace"
	disasterRecoveryConfigName = "sample-disaster-recovery-config"
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
	log.Println("eventhub namespace:", *namespace.ID)

	secondNamespace, err := createSecondNamespace(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("eventhub second namespace:", *secondNamespace.ID)

	ava, err := checkNameAva(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("check name availability:", *ava.NameAvailable)

	disasterRecoveryConfig, err := createDisasterRecoveryConfig(ctx, conn, *secondNamespace.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("disaster recovery config:", *disasterRecoveryConfig.ID)

	disasterRecoveryConfig, err = getDisasterRecoveryConfig(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get disaster recovery config:", *disasterRecoveryConfig.ID)

	// Only after breakPairing or failOVer can clean resource
	breakPairing, err := breakPairingDisasterRecoveryConfig(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("break pairing:", *breakPairing)

	//failOver, err := failOverDisasterRecoveryConfig(ctx, conn)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Println("fail over:", *failOver)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createNamespace(ctx context.Context, conn *arm.Connection) (*armeventhub.EHNamespace, error) {
	namespacesClient := armeventhub.NewNamespacesClient(conn, subscriptionID)

	pollerResp, err := namespacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		namespacesName,
		armeventhub.EHNamespace{
			TrackedResource: armeventhub.TrackedResource{
				Location: to.StringPtr(location),
				Tags: map[string]*string{
					"tag1": to.StringPtr("value1"),
					"tag2": to.StringPtr("value2"),
				},
			},
			SKU: &armeventhub.SKU{
				Name: armeventhub.SKUNameStandard.ToPtr(),
				Tier: armeventhub.SKUTierStandard.ToPtr(),
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
	return &resp.EHNamespace, nil
}

func createSecondNamespace(ctx context.Context, conn *arm.Connection) (*armeventhub.EHNamespace, error) {
	namespacesClient := armeventhub.NewNamespacesClient(conn, subscriptionID)

	pollerResp, err := namespacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		secondNamespacesName,
		armeventhub.EHNamespace{
			TrackedResource: armeventhub.TrackedResource{
				Location: to.StringPtr("eastus"),
				Tags: map[string]*string{
					"tag1": to.StringPtr("value1"),
					"tag2": to.StringPtr("value2"),
				},
			},
			SKU: &armeventhub.SKU{
				Name: armeventhub.SKUNameStandard.ToPtr(),
				Tier: armeventhub.SKUTierStandard.ToPtr(),
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
	return &resp.EHNamespace, nil
}

func checkNameAva(ctx context.Context, conn *arm.Connection) (*armeventhub.CheckNameAvailabilityResult, error) {
	disasterRecoveryConfigsClient := armeventhub.NewDisasterRecoveryConfigsClient(conn, subscriptionID)

	resp, err := disasterRecoveryConfigsClient.CheckNameAvailability(
		ctx,
		resourceGroupName,
		namespacesName,
		armeventhub.CheckNameAvailabilityParameter{
			Name: to.StringPtr(secondNamespacesName),
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.CheckNameAvailabilityResult, nil
}

func createDisasterRecoveryConfig(ctx context.Context, conn *arm.Connection, secondNamespaceID string) (*armeventhub.ArmDisasterRecovery, error) {
	disasterRecoveryConfigsClient := armeventhub.NewDisasterRecoveryConfigsClient(conn, subscriptionID)

	resp, err := disasterRecoveryConfigsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		namespacesName,
		disasterRecoveryConfigName,
		armeventhub.ArmDisasterRecovery{
			Properties: &armeventhub.ArmDisasterRecoveryProperties{
				PartnerNamespace: to.StringPtr(secondNamespaceID),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.ArmDisasterRecovery, nil
}

func getDisasterRecoveryConfig(ctx context.Context, conn *arm.Connection) (*armeventhub.ArmDisasterRecovery, error) {
	disasterRecoveryConfigsClient := armeventhub.NewDisasterRecoveryConfigsClient(conn, subscriptionID)

	resp, err := disasterRecoveryConfigsClient.Get(ctx, resourceGroupName, namespacesName, disasterRecoveryConfigName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ArmDisasterRecovery, nil
}

func breakPairingDisasterRecoveryConfig(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	disasterRecoveryConfigsClient := armeventhub.NewDisasterRecoveryConfigsClient(conn, subscriptionID)

	resp, err := disasterRecoveryConfigsClient.BreakPairing(ctx, resourceGroupName, namespacesName, disasterRecoveryConfigName, nil)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}

func failOverDisasterRecoveryConfig(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	disasterRecoveryConfigsClient := armeventhub.NewDisasterRecoveryConfigsClient(conn, subscriptionID)

	resp, err := disasterRecoveryConfigsClient.FailOver(ctx, resourceGroupName, secondNamespacesName, disasterRecoveryConfigName, nil)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
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
