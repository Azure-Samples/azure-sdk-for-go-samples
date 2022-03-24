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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID     string
	location           = "eastus"
	resourceGroupName  = "sample-resource-group"
	registryName       = "sample2registry"
	importPipelineName = "sample2import2pipeline"
	exportPipelineName = "sample2export2pipeline"
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

	registry, err := createRegistry(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("registry:", *registry.ID)

	importPipeline, err := createImportPipeline(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("import pipeline:", *importPipeline.ID)

	exportPipeline, err := createExportPipeline(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("export pipeline:", *exportPipeline.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createRegistry(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.Registry, error) {
	registriesClient := armcontainerregistry.NewRegistriesClient(subscriptionID, cred, nil)

	pollerResp, err := registriesClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		armcontainerregistry.Registry{
			Resource: armcontainerregistry.Resource{
				Location: to.StringPtr(location),
				Tags: map[string]*string{
					"key": to.StringPtr("value"),
				},
			},
			SKU: &armcontainerregistry.SKU{
				Name: armcontainerregistry.SKUNamePremium.ToPtr(),
			},
			Properties: &armcontainerregistry.RegistryProperties{
				AdminUserEnabled: to.BoolPtr(true),
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
	return &resp.Registry, nil
}

func createImportPipeline(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.ImportPipeline, error) {
	importPipelinesClient := armcontainerregistry.NewImportPipelinesClient(subscriptionID, cred, nil)

	pollerResp, err := importPipelinesClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		importPipelineName,
		armcontainerregistry.ImportPipeline{
			Location: to.StringPtr(location),
			Identity: &armcontainerregistry.IdentityProperties{
				Type: armcontainerregistry.ResourceIdentityTypeSystemAssigned.ToPtr(),
			},
			Properties: &armcontainerregistry.ImportPipelineProperties{
				Source: &armcontainerregistry.ImportPipelineSourceProperties{
					KeyVaultURI: to.StringPtr("https://myvault.vault.azure.net/secrets/acrimportsas"),
					Type:        armcontainerregistry.PipelineSourceTypeAzureStorageBlobContainer.ToPtr(),
					URI:         to.StringPtr("https://accountname.blob.core.windows.net/containername"),
				},
				Options: []*armcontainerregistry.PipelineOptions{
					armcontainerregistry.PipelineOptionsContinueOnErrors.ToPtr(),
					armcontainerregistry.PipelineOptionsDeleteSourceBlobOnSuccess.ToPtr(),
					armcontainerregistry.PipelineOptionsOverwriteTags.ToPtr(),
				},
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
	return &resp.ImportPipeline, nil
}

func createExportPipeline(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.ExportPipeline, error) {
	exportPipelinesClient := armcontainerregistry.NewExportPipelinesClient(subscriptionID, cred, nil)

	pollerResp, err := exportPipelinesClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		exportPipelineName,
		armcontainerregistry.ExportPipeline{
			Location: to.StringPtr(location),
			Identity: &armcontainerregistry.IdentityProperties{
				Type: armcontainerregistry.ResourceIdentityTypeSystemAssigned.ToPtr(),
			},
			Properties: &armcontainerregistry.ExportPipelineProperties{
				Target: &armcontainerregistry.ExportPipelineTargetProperties{
					KeyVaultURI: to.StringPtr("https://myvault.vault.azure.net/secrets/acrimportsas"),
					Type:        to.StringPtr("AzureStorageBlobContainer"),
					URI:         to.StringPtr("https://accountname.blob.core.windows.net/containername"),
				},
				Options: []*armcontainerregistry.PipelineOptions{
					armcontainerregistry.PipelineOptionsOverwriteBlobs.ToPtr(),
				},
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
	return &resp.ExportPipeline, nil
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
