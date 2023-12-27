// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	deploymentName    = "sample-deployment"
)

var (
	resourcesClientFactory *armresources.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	deploymentsClient   *armresources.DeploymentsClient
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

	resourcesClientFactory, err = armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	resourceGroupClient = resourcesClientFactory.NewResourceGroupsClient()
	deploymentsClient = resourcesClientFactory.NewDeploymentsClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	exist, err := checkExistDeployment(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("deployment is exist:", exist)

	template, err := readJson("testdata/template.json")
	if err != nil {
		log.Fatal(err)
	}
	params, err := readJson("testdata/parameters.json")
	if err != nil {
		log.Fatal(err)
	}
	deploymentExtended, err := createDeployment(ctx, template, params)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("created deployment:", *deploymentExtended.ID)

	validateResult, err := validateDeployment(ctx, template, params)
	if err != nil {
		log.Fatal(err)
	}
	data, _ := json.Marshal(validateResult)
	log.Println("validate deployment:", string(data))

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func readJson(path string) (map[string]interface{}, error) {
	templateFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	template := make(map[string]interface{})
	if err := json.Unmarshal(templateFile, &template); err != nil {
		return nil, err
	}

	return template, nil
}

func checkExistDeployment(ctx context.Context) (bool, error) {

	boolResp, err := deploymentsClient.CheckExistence(ctx, resourceGroupName, deploymentName, nil)
	if err != nil {
		return false, err
	}

	return boolResp.Success, nil
}

func createDeployment(ctx context.Context, template, params map[string]interface{}) (*armresources.DeploymentExtended, error) {

	deploymentPollerResp, err := deploymentsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		deploymentName,
		armresources.Deployment{
			Properties: &armresources.DeploymentProperties{
				Template:   template,
				Parameters: params,
				Mode:       to.Ptr(armresources.DeploymentModeIncremental),
			},
		},
		nil)

	if err != nil {
		return nil, fmt.Errorf("cannot create deployment: %v", err)
	}

	resp, err := deploymentPollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot get the create deployment future respone: %v", err)
	}

	return &resp.DeploymentExtended, nil
}

func validateDeployment(ctx context.Context, template, params map[string]interface{}) (*armresources.DeploymentValidateResult, error) {

	pollerResp, err := deploymentsClient.BeginValidate(
		ctx,
		resourceGroupName,
		deploymentName,
		armresources.Deployment{
			Properties: &armresources.DeploymentProperties{
				Template:   template,
				Parameters: params,
				Mode:       to.Ptr(armresources.DeploymentModeIncremental),
			},
		},
		nil)

	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &resp.DeploymentValidateResult, nil
}

func createResourceGroup(ctx context.Context) (*armresources.ResourceGroup, error) {

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

func cleanup(ctx context.Context) error {

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
