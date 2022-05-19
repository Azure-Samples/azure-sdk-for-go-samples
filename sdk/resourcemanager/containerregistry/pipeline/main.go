// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
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
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createRegistry(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.Registry, error) {
	registriesClient, err := armcontainerregistry.NewRegistriesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := registriesClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		armcontainerregistry.Registry{
			Location: to.Ptr(location),
			Tags: map[string]*string{
				"key": to.Ptr("value"),
			},
			SKU: &armcontainerregistry.SKU{
				Name: to.Ptr(armcontainerregistry.SKUNamePremium),
			},
			Properties: &armcontainerregistry.RegistryProperties{
				AdminUserEnabled: to.Ptr(true),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Registry, nil
}

func createImportPipeline(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.ImportPipeline, error) {
	importPipelinesClient, err := armcontainerregistry.NewImportPipelinesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := importPipelinesClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		importPipelineName,
		armcontainerregistry.ImportPipeline{
			Location: to.Ptr(location),
			Identity: &armcontainerregistry.IdentityProperties{
				Type: to.Ptr(armcontainerregistry.ResourceIdentityTypeSystemAssigned),
			},
			Properties: &armcontainerregistry.ImportPipelineProperties{
				Source: &armcontainerregistry.ImportPipelineSourceProperties{
					KeyVaultURI: to.Ptr("https://myvault.vault.azure.net/secrets/acrimportsas"),
					Type:        to.Ptr(armcontainerregistry.PipelineSourceTypeAzureStorageBlobContainer),
					URI:         to.Ptr("https://accountname.blob.core.windows.net/containername"),
				},
				Options: []*armcontainerregistry.PipelineOptions{
					to.Ptr(armcontainerregistry.PipelineOptionsContinueOnErrors),
					to.Ptr(armcontainerregistry.PipelineOptionsDeleteSourceBlobOnSuccess),
					to.Ptr(armcontainerregistry.PipelineOptionsOverwriteTags),
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ImportPipeline, nil
}

func createExportPipeline(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.ExportPipeline, error) {
	exportPipelinesClient, err := armcontainerregistry.NewExportPipelinesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := exportPipelinesClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		exportPipelineName,
		armcontainerregistry.ExportPipeline{
			Location: to.Ptr(location),
			Identity: &armcontainerregistry.IdentityProperties{
				Type: to.Ptr(armcontainerregistry.ResourceIdentityTypeSystemAssigned),
			},
			Properties: &armcontainerregistry.ExportPipelineProperties{
				Target: &armcontainerregistry.ExportPipelineTargetProperties{
					KeyVaultURI: to.Ptr("https://myvault.vault.azure.net/secrets/acrimportsas"),
					Type:        to.Ptr("AzureStorageBlobContainer"),
					URI:         to.Ptr("https://accountname.blob.core.windows.net/containername"),
				},
				Options: []*armcontainerregistry.PipelineOptions{
					to.Ptr(armcontainerregistry.PipelineOptionsOverwriteBlobs),
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ExportPipeline, nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.Ptr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, cred azcore.TokenCredential) error {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
