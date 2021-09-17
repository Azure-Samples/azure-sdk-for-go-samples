package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/armstorage"
)

var (
	subscriptionID     string
	location           = "westus"
	resourceGroupName  = "sample-resource-group"
	storageAccountName = "sample2storage2account"
	containerName      = "blob2container"
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

	storageAccount, err := createStorageAccount(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("storage account:", *storageAccount.ID)

	blobContainer, err := createBlobContainers(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("blob container:", *blobContainer.ID)

	blobContainer, err = getBlobContainer(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("blob container ID:", *blobContainer.ID)

	containerItems := listBlobContainer(ctx, conn)
	log.Println("blob container list:")
	for _, item := range containerItems {
		log.Println("\t", *item.ID)
	}

	immutabilityPolicy, err := createContainerImmutabilityPolicy(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(*immutabilityPolicy.ID, immutabilityPolicy.Etag)

	immutabilityPolicy, err = getContainerImmutabilityPolicy(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get immutability policy:", *immutabilityPolicy.ID, *immutabilityPolicy.Etag)

	immutabilityPolicy, err = updateContainerImmutabilityPolicy(ctx, conn, *immutabilityPolicy.Etag)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("updated immutability policy:", *immutabilityPolicy.ID)

	immutabilityPolicy, err = extendContainerImmutabilityPolicy(ctx, conn, *immutabilityPolicy.Etag)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("extend immutability policy.")

	immutabilityPolicy, err = deleteContainerImmutabilityPolicy(ctx, conn, *immutabilityPolicy.Etag)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("deleted immutability policy.")

	legalHold, err := setLegalHoldBlobContainer(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	data, err := json.Marshal(legalHold)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Set Legal Hold:", string(data))

	legalHold, err = clearLegalHoldBlobContainer(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	data, err = json.Marshal(legalHold)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Clear Legal Hold:", string(data))

	leaseContainerResp, err := leaseBlobContainer(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("lease blob container ID:", *leaseContainerResp.LeaseID)

	objectLevelWorm, err := objectLevelWormBlobContainer(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(objectLevelWorm)

	blobServices(ctx, conn)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createStorageAccount(ctx context.Context, conn *arm.Connection) (*armstorage.StorageAccount, error) {
	storageAccountClient := armstorage.NewStorageAccountsClient(conn, subscriptionID)

	pollerResp, err := storageAccountClient.BeginCreate(
		ctx,
		resourceGroupName,
		storageAccountName,
		armstorage.StorageAccountCreateParameters{
			Kind: armstorage.KindStorageV2.ToPtr(),
			SKU: &armstorage.SKU{
				Name: armstorage.SKUNameStandardLRS.ToPtr(),
			},
			Location: to.StringPtr(location),
			Properties: &armstorage.StorageAccountPropertiesCreateParameters{
				AccessTier: armstorage.AccessTierCool.ToPtr(),
				Encryption: &armstorage.Encryption{
					Services: &armstorage.EncryptionServices{
						File: &armstorage.EncryptionService{
							KeyType: armstorage.KeyTypeAccount.ToPtr(),
							Enabled: to.BoolPtr(true),
						},
						Blob: &armstorage.EncryptionService{
							KeyType: armstorage.KeyTypeAccount.ToPtr(),
							Enabled: to.BoolPtr(true),
						},
					},
					KeySource: armstorage.KeySourceMicrosoftStorage.ToPtr(),
				},
			},
		}, nil)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.StorageAccount, nil
}

func createBlobContainers(ctx context.Context, conn *arm.Connection) (*armstorage.BlobContainer, error) {
	blobContainerClient := armstorage.NewBlobContainersClient(conn, subscriptionID)

	blobContainerResp, err := blobContainerClient.Create(
		ctx,
		resourceGroupName,
		storageAccountName,
		containerName,
		armstorage.BlobContainer{
			ContainerProperties: &armstorage.ContainerProperties{
				PublicAccess: armstorage.PublicAccessNone.ToPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &blobContainerResp.BlobContainer, nil
}

func getBlobContainer(ctx context.Context, conn *arm.Connection) (blobContainer *armstorage.BlobContainer, err error) {
	blobContainerClient := armstorage.NewBlobContainersClient(conn, subscriptionID)

	blobContainerResp, err := blobContainerClient.Get(ctx, resourceGroupName, storageAccountName, containerName, nil)
	if err != nil {
		return
	}

	blobContainer = &blobContainerResp.BlobContainer
	return
}

func listBlobContainer(ctx context.Context, conn *arm.Connection) (listItems []*armstorage.ListContainerItem) {
	blobContainerClient := armstorage.NewBlobContainersClient(conn, subscriptionID)

	containerItemsPager := blobContainerClient.List(resourceGroupName, storageAccountName, nil)

	listItems = make([]*armstorage.ListContainerItem, 0)
	for containerItemsPager.NextPage(ctx) {
		pageResp := containerItemsPager.PageResponse()
		listItems = append(listItems, pageResp.ListContainerItems.Value...)
	}
	return
}

func createContainerImmutabilityPolicy(ctx context.Context, conn *arm.Connection) (*armstorage.ImmutabilityPolicy, error) {
	blobContainerClient := armstorage.NewBlobContainersClient(conn, subscriptionID)

	immutabilityPolicyResp, err := blobContainerClient.CreateOrUpdateImmutabilityPolicy(
		ctx,
		resourceGroupName,
		storageAccountName,
		containerName,
		&armstorage.BlobContainersCreateOrUpdateImmutabilityPolicyOptions{
			//IfMatch: to.StringPtr("*"),
			Parameters: &armstorage.ImmutabilityPolicy{
				Properties: &armstorage.ImmutabilityPolicyProperty{
					//AllowProtectedAppendWrites: to.BoolPtr(true),
					//ImmutabilityPeriodSinceCreationInDays: to.Int32Ptr(2), //retention interval
				},
			},
		})
	if err != nil {
		return nil, err
	}

	return &immutabilityPolicyResp.ImmutabilityPolicy, nil
}

func updateContainerImmutabilityPolicy(ctx context.Context, conn *arm.Connection, ifMatch string) (*armstorage.ImmutabilityPolicy, error) {
	blobContainerClient := armstorage.NewBlobContainersClient(conn, subscriptionID)

	immutabilityPolicyResp, err := blobContainerClient.CreateOrUpdateImmutabilityPolicy(
		ctx,
		resourceGroupName,
		storageAccountName,
		containerName,
		&armstorage.BlobContainersCreateOrUpdateImmutabilityPolicyOptions{
			//IfMatch: to.StringPtr(ifMatch),
			Parameters: &armstorage.ImmutabilityPolicy{
				Properties: &armstorage.ImmutabilityPolicyProperty{
					AllowProtectedAppendWrites:            to.BoolPtr(false),
					ImmutabilityPeriodSinceCreationInDays: to.Int32Ptr(0), //retention interval
					//State:                                 armstorage.ImmutabilityPolicyStateUnlocked.ToPtr(),
				},
			},
		})
	if err != nil {
		return nil, err
	}

	return &immutabilityPolicyResp.ImmutabilityPolicy, nil
}

func getContainerImmutabilityPolicy(ctx context.Context, conn *arm.Connection) (*armstorage.ImmutabilityPolicy, error) {
	blobContainerClient := armstorage.NewBlobContainersClient(conn, subscriptionID)

	immutabilityPolicyResp, err := blobContainerClient.GetImmutabilityPolicy(
		ctx,
		resourceGroupName,
		storageAccountName,
		containerName,
		nil)
	if err != nil {
		return nil, err
	}

	return &immutabilityPolicyResp.ImmutabilityPolicy, nil
}

func extendContainerImmutabilityPolicy(ctx context.Context, conn *arm.Connection, ifMatch string) (*armstorage.ImmutabilityPolicy, error) {
	blobContainerClient := armstorage.NewBlobContainersClient(conn, subscriptionID)

	immutabilityPolicyResp, err := blobContainerClient.ExtendImmutabilityPolicy(
		ctx,
		resourceGroupName,
		storageAccountName,
		containerName,
		ifMatch,
		nil)
	if err != nil {
		return nil, err
	}

	return &immutabilityPolicyResp.ImmutabilityPolicy, nil
}

func deleteContainerImmutabilityPolicy(ctx context.Context, conn *arm.Connection, ifMatch string) (*armstorage.ImmutabilityPolicy, error) {
	blobContainerClient := armstorage.NewBlobContainersClient(conn, subscriptionID)

	immutabilityPolicyResp, err := blobContainerClient.DeleteImmutabilityPolicy(
		ctx,
		resourceGroupName,
		storageAccountName,
		containerName,
		ifMatch,
		&armstorage.BlobContainersDeleteImmutabilityPolicyOptions{})
	if err != nil {
		return nil, err
	}

	return &immutabilityPolicyResp.ImmutabilityPolicy, nil
}

func setLegalHoldBlobContainer(ctx context.Context, conn *arm.Connection) (*armstorage.LegalHold, error) {
	blobContainerClient := armstorage.NewBlobContainersClient(conn, subscriptionID)

	legalHoldResp, err := blobContainerClient.SetLegalHold(
		ctx,
		resourceGroupName,
		storageAccountName,
		containerName,
		armstorage.LegalHold{
			Tags: []*string{
				to.StringPtr("sample"),
				to.StringPtr("sample2"),
			},
		},
		nil)
	if err != nil {
		return nil, err
	}

	return &legalHoldResp.LegalHold, nil
}

func clearLegalHoldBlobContainer(ctx context.Context, conn *arm.Connection) (*armstorage.LegalHold, error) {
	blobContainerClient := armstorage.NewBlobContainersClient(conn, subscriptionID)

	legalHoldResp, err := blobContainerClient.ClearLegalHold(
		ctx,
		resourceGroupName,
		storageAccountName,
		containerName,
		armstorage.LegalHold{
			Tags: []*string{
				to.StringPtr("sample"),
			},
		},
		nil)
	if err != nil {
		return nil, err
	}

	return &legalHoldResp.LegalHold, nil
}

func objectLevelWormBlobContainer(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	blobContainerClient := armstorage.NewBlobContainersClient(conn, subscriptionID)

	pollerResp, err := blobContainerClient.BeginObjectLevelWorm(
		ctx,
		resourceGroupName,
		storageAccountName,
		containerName,
		nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return resp.RawResponse, nil
}

func leaseBlobContainer(ctx context.Context, conn *arm.Connection) (*armstorage.LeaseContainerResponse, error) {
	blobContainerClient := armstorage.NewBlobContainersClient(conn, subscriptionID)

	leaseContainerResp, err := blobContainerClient.Lease(
		ctx,
		resourceGroupName,
		storageAccountName,
		containerName,
		&armstorage.BlobContainersLeaseOptions{
			Parameters: &armstorage.LeaseContainerRequest{
				Action:        armstorage.LeaseContainerRequestActionAcquire.ToPtr(),
				LeaseDuration: to.Int32Ptr(-1),
			},
		})
	if err != nil {
		return nil, err
	}

	return &leaseContainerResp.LeaseContainerResponse, nil
}

func blobServices(ctx context.Context, conn *arm.Connection) {
	blobServicesProperties, err := setBlobServices(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	if blobServicesProperties == nil {
		log.Fatal("what")
	}
	log.Println(*blobServicesProperties.ID)

	blobServicesProperties, err = getBlobServices(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(*blobServicesProperties.ID)

	listBlob, err := listBlobServices(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	for _, properties := range listBlob {
		log.Println(*properties.ID)
	}
}

func setBlobServices(ctx context.Context, conn *arm.Connection) (*armstorage.BlobServiceProperties, error) {
	blobServicesClient := armstorage.NewBlobServicesClient(conn, subscriptionID)

	blobServicesPropertiesResp, err := blobServicesClient.SetServiceProperties(
		ctx,
		resourceGroupName,
		storageAccountName,
		armstorage.BlobServiceProperties{
			BlobServiceProperties: &armstorage.BlobServicePropertiesAutoGenerated{},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &blobServicesPropertiesResp.BlobServiceProperties, nil
}

func getBlobServices(ctx context.Context, conn *arm.Connection) (*armstorage.BlobServiceProperties, error) {
	blobServicesClient := armstorage.NewBlobServicesClient(conn, subscriptionID)

	blobServicesResp, err := blobServicesClient.GetServiceProperties(ctx, resourceGroupName, storageAccountName, nil)
	if err != nil {
		return nil, err
	}

	return &blobServicesResp.BlobServiceProperties, nil
}

func listBlobServices(ctx context.Context, conn *arm.Connection) ([]*armstorage.BlobServiceProperties, error) {
	blobServicesClient := armstorage.NewBlobServicesClient(conn, subscriptionID)

	blobServicesResp, err := blobServicesClient.List(ctx, resourceGroupName, storageAccountName, nil)
	if err != nil {
		return nil, err
	}

	return blobServicesResp.BlobServiceItems.Value, nil
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
