// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/subosito/gotenv"
)

var (
	tenantId              string
	subscriptionId        string
	clientId              string
	clientSecret          string
	resourceGroupName     string
	deploymentName        string
	resourceGroupLocation string
	pathToTemplateFile    string
	pathToParametersFile  string

	armToken *adal.ServicePrincipalToken
)

func main() {
	group, err := CreateGroup()
	if err != nil {
		log.Fatalf("failed to create group: %v", err)
	}
	log.Printf("created group: %v\n", group)

	log.Printf("starting deployment\n")
	result, errC := CreateDeployment()
	wait := <-errC
	if wait != nil {
		log.Fatalf("failed to deploy: %v\n", wait)
	}
	log.Printf("started deployment: %v\n", (<-result).Properties)
}

func init() {
	gotenv.Load()
	tenantId = os.Getenv("AZURE_TENANT_ID")
	subscriptionId = os.Getenv("AZURE_SUBSCRIPTION_ID")
	clientId = os.Getenv("AZURE_CLIENT_ID")
	clientSecret = os.Getenv("AZURE_CLIENT_SECRET")
	resourceGroupName = os.Getenv("AZURE_RG_NAME")
	deploymentName = "template-deployment-test"
	resourceGroupLocation = os.Getenv("AZURE_LOCATION")
	pathToTemplateFile, _ = filepath.Abs("test_data/template.json")
	pathToParametersFile, _ = filepath.Abs("test_data/parameters.json")

	// get OAuth token using Service Principal credentials
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, tenantId)
	if err != nil {
		log.Fatalf("%s: %v\n", "failed to get OAuth config", err)
	}
	token, err := adal.NewServicePrincipalToken(
		*oauthConfig,
		clientId,
		clientSecret,
		azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		log.Fatalf("failed to get token: %v\n", err)
	}
	armToken = token
}

// CreateGroup creates a new resource group named by env var
func CreateGroup() (resources.Group, error) {
	groupsClient := resources.NewGroupsClient(subscriptionId)
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(armToken)

	return groupsClient.CreateOrUpdate(
		resourceGroupName,
		resources.Group{
			Location: to.StringPtr(resourceGroupLocation)})
}

// ReadJSON reads a file and unmarshals the JSON
func ReadJSON(path string) (*map[string]interface{}, error) {
	_json, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read template file: %v\n", err)
	}
	_map := make(map[string]interface{})
	json.Unmarshal(_json, &_map)
	return &_map, nil
}

// CreateDeployment deploys a template to Azure Resource Manager
func CreateDeployment() (<-chan resources.DeploymentExtended, <-chan error) {
	deploymentsClient := resources.NewDeploymentsClient(subscriptionId)
	deploymentsClient.Authorizer = autorest.NewBearerAuthorizer(armToken)

	_template, err := ReadJSON(pathToTemplateFile)
	if err != nil {
		log.Fatalf("failed to read template.json: %v", err)
	}
	_params, err := ReadJSON(pathToParametersFile)
	if err != nil {
		log.Fatalf("failed to read parameters.json: %v", err)
	}

	return deploymentsClient.CreateOrUpdate(
		resourceGroupName,
		deploymentName,
		resources.Deployment{
			Properties: &resources.DeploymentProperties{
				Template:   _template,
				Parameters: _params,
				Mode:       resources.Incremental,
			},
		},
		nil,
	)
}
