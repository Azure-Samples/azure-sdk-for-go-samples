package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/network/mgmt/network"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/resources"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

type authInfo struct {
	TenantID               string
	SubscriptionID         string
	ServicePrincipalID     string
	ServicePrincipalSecret string
}

var (
	config = authInfo{
		TenantID:               "",
		SubscriptionID:         "",
		ServicePrincipalID:     "",
		ServicePrincipalSecret: "",
	}

	ctx = context.Background()

	resourceGroupName     = "GoVMQuickstart"
	resourceGroupLocation = "eastus"

	deploymentName = "VMDeployQuickstart"
	templateFile   = "vm-quickstart-template.json"
	parametersFile = "vm-quickstart-params.json"

	token *adal.ServicePrincipalToken
)

// Authenticate with the Azure services over OAuth, using a service principal.
func init() {
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, config.TenantID)
	if err != nil {
		log.Fatalf("%s: %v\n", "failed to get OAuth config", err)
	}
	token, err = adal.NewServicePrincipalToken(
		*oauthConfig,
		config.ServicePrincipalID,
		config.ServicePrincipalSecret,
		azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		log.Fatalf("faled to get token: %v\n", err)
	}
}

func main() {
	group, err := createGroup()
	if err != nil {
		log.Fatalf("failed to create group: %v", err)
	}
	log.Printf("created group: %v\n", *group.Name)

	log.Printf("starting deployment\n")
	result, err := createDeployment()
	if err != nil {
		log.Fatalf("Failed to deploy correctly: %v", err)
	}
	log.Printf("Completed deployment: %v", *result.Name)
	getLogin()
}

// Create a resource group for the deployment.
func createGroup() (resources.Group, error) {
	groupsClient := resources.NewGroupsClient(config.SubscriptionID)
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return groupsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		resources.Group{
			Location: to.StringPtr(resourceGroupLocation)})
}

// Create the deployment
func createDeployment() (resources.DeploymentExtended, error) {
	deploymentsClient := resources.NewDeploymentsClient(config.SubscriptionID)
	deploymentsClient.Authorizer = autorest.NewBearerAuthorizer(token)

	template, err := readJSON(templateFile)
	if err != nil {
		return resources.DeploymentExtended{}, err
	}
	params, err := readJSON(parametersFile)
	if err != nil {
		return resources.DeploymentExtended{}, err
	}

	deploymentFuture, err := deploymentsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		deploymentName,
		resources.Deployment{
			Properties: &resources.DeploymentProperties{
				Template:   template,
				Parameters: params,
				Mode:       resources.Incremental,
			},
		},
	)
	if err != nil {
		log.Fatalf("Failed to create deployment: %v", err)
	}
	err = deploymentFuture.Future.WaitForCompletion(ctx, deploymentsClient.BaseClient.Client)
	if err != nil {
		log.Fatalf("Error while waiting for deployment creation: %v", err)
	}
	return deploymentFuture.Result(deploymentsClient)
}

// Get login information by querying the deployed public IP resource.
func getLogin() {
	params, err := readJSON(parametersFile)
	if err != nil {
		log.Fatalf("Unable to read parameters. Get login information with `az network public-ip list -g %s", resourceGroupName)
	}

	addressClient := network.NewPublicIPAddressesClient(config.SubscriptionID)
	addressClient.Authorizer = autorest.NewBearerAuthorizer(token)
	ipName := (*params)["publicIPAddresses_QuickstartVM_ip_name"].(map[string]interface{})
	ipAddress, err := addressClient.Get(ctx, resourceGroupName, ipName["value"].(string), "")
	if err != nil {
		log.Fatalf("Unable to get IP information. Try using `az network public-ip list -g %s", resourceGroupName)
	}

	vmUser := (*params)["vm_user"].(map[string]interface{})
	vmPass := (*params)["vm_password"].(map[string]interface{})

	log.Printf("Log in with ssh: %s@%s, password: %s",
		vmUser["value"].(string),
		*ipAddress.PublicIPAddressPropertiesFormat.IPAddress,
		vmPass["value"].(string))
}

func readJSON(path string) (*map[string]interface{}, error) {
	_json, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read template file: %v\n", err)
	}
	_map := make(map[string]interface{})
	json.Unmarshal(_json, &_map)
	return &_map, nil
}
