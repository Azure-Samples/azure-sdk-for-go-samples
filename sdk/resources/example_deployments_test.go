// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package resources

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/arm/resources/2020-06-01/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/armcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// func ExampleDeploymentsClient_BeginCreateOrUpdate() {
func CreateOrUpdateDeployment(template string, parameters string) (*armresources.DeploymentExtended, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	client := armresources.NewDeploymentsClient(armcore.NewDefaultConnection(cred, nil), "<subscription ID>")
	data, err := ioutil.ReadFile(template)
	if err != nil {
		return nil, err
	}
	contents := make(map[string]interface{})
	if err := json.Unmarshal(data, &contents); err != nil {
		return nil, err
	}
	data, err = ioutil.ReadFile(parameters)
	if err != nil {
		return nil, err
	}
	params := make(map[string]interface{})
	if err := json.Unmarshal(data, &params); err != nil {
		return nil, err
	}
	poller, err := client.BeginCreateOrUpdate(
		context.Background(),
		"<resource group>",
		"<deployment name>",
		armresources.Deployment{
			Properties: &armresources.DeploymentProperties{
				Template:   contents,
				Parameters: params,
				Mode:       armresources.DeploymentModeIncremental.ToPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := poller.PollUntilDone(context.Background(), 5*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.DeploymentExtended, nil
}

// func ExampleDeploymentsClient_BeginValidate() {
func ValidateDeployment(template string, parameters string) (*armresources.DeploymentValidateResult, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	client := armresources.NewDeploymentsClient(armcore.NewDefaultConnection(cred, nil), "<subscription ID>")
	data, err := ioutil.ReadFile(template)
	if err != nil {
		return nil, err
	}
	contents := make(map[string]interface{})
	if err := json.Unmarshal(data, &contents); err != nil {
		return nil, err
	}
	data, err = ioutil.ReadFile(parameters)
	if err != nil {
		return nil, err
	}
	params := make(map[string]interface{})
	if err := json.Unmarshal(data, &params); err != nil {
		return nil, err
	}
	poller, err := client.BeginValidate(
		context.Background(),
		"<resource group>",
		"<deployment name>",
		armresources.Deployment{
			Properties: &armresources.DeploymentProperties{
				Mode:       armresources.DeploymentModeIncremental.ToPtr(),
				Parameters: params,
				Template:   contents,
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := poller.PollUntilDone(context.Background(), 5*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.DeploymentValidateResult, nil
}

// func ExampleDeploymentsClient_Get() {
func GetDeployment(id string) (*armresources.DeploymentExtended, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	client := armresources.NewDeploymentsClient(armcore.NewDefaultConnection(cred, nil), "<subscription ID>")
	dep, err := client.Get(context.Background(), "<resource group>", id, nil)
	if err != nil {
		return nil, err
	}
	return dep.DeploymentExtended, nil
}

func ExampleDeploymentsClient() {
	depValid, err := ValidateDeployment("template.json", "parameters.json")
	if err != nil {
		log.Fatalf("failed to validate deployment: %v", err)
	}
	log.Print("validated deployment")
	if depValid.Error != nil {
		log.Fatalf("deployment validation failed: %v", depValid.Error.Message)
	}
	dep, err := CreateOrUpdateDeployment("template.json", "parameters.json")
	if err != nil {
		log.Fatalf("failed to create deployment: %v", err)
	}
	log.Print("created deployment")
	depCheck, err := GetDeployment(*dep.Name)
	if err != nil {
		log.Fatalf("failed to get deployment: %v", err)
	}
	log.Printf("retreived deployment. Deployment ID: %v", *depCheck.ID)
}
