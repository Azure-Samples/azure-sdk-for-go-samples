package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

var (
	tenantId              = "your tenant id"
	subscriptionId        = "your subscription id"
	clientId              = "your AAD application id"
	clientSecret          = "your AAD application secret"
	resourceGroupName     = "a-resource-group"
	deploymentName        = "a-deployment"
	resourceGroupLocation = "an Azure region"
	pathToTemplateFile    = "template.json"
	pathToParametersFile  = "parameters.json"

	armToken *adal.ServicePrincipalToken
)

func main() {
	// create the resource group
	group, err := CreateGroup()
	if err != nil {
		log.Fatalf("failed to create group: %v", err)
	}
	log.Printf("created group: %v\n", group)

	// deploy template into resource group
	log.Printf("starting deployment\n")
	result, errC := CreateDeployment()
	wait := <-errC
	if wait != nil {
		log.Fatalf("failed to deploy: %v\n", wait)
	}
	log.Printf("started deployment: %v\n", (<-result).Properties)
}

func init() {
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
			Location: to.StringPtr(location),
		},
	)
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
