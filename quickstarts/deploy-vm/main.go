package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"

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

const (
	resourceGroupName     = "GoVMQuickstart"
	resourceGroupLocation = "eastus"

	deploymentName = "VMDeployQuickstart"
	templateFile   = "vm-quickstart-template.json"
	parametersFile = "vm-quickstart-params.json"
)

var (
	config = authInfo{ // Your application credentials
		TenantID:               "", // Azure account tenantID
		SubscriptionID:         "", // Azure subscription subscriptionID
		ServicePrincipalID:     "", // Service principal appId
		ServicePrincipalSecret: "", // Service principal password/secret
	}

	ctx = context.Background()

	token *adal.ServicePrincipalToken
)

// Authenticate with the Azure services over OAuth, using a service principal.
func init() {
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, config.TenantID)
	if err != nil {
		log.Fatalf("Failed to get OAuth config: %v\n", err)
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

	log.Println("starting deployment")
	result, err := createDeployment()
	if err != nil {
		log.Fatalf("Failed to deploy correctly: %v", err)
	}
	log.Printf("Completed deployment: %v", *result.Name)
	getLogin()
}

// Create a resource group for the deployment.
func createGroup() (group resources.Group, err error) {
	groupsClient := resources.NewGroupsClient(config.SubscriptionID)
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return groupsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		resources.Group{
			Location: to.StringPtr(resourceGroupLocation)})
}

// Create the deployment
func createDeployment() (deployment resources.DeploymentExtended, err error) {
	template, err := readJSON(templateFile)
	if err != nil {
		return
	}
	params, err := readJSON(parametersFile)
	if err != nil {
		return
	}

	deploymentsClient := resources.NewDeploymentsClient(config.SubscriptionID)
	deploymentsClient.Authorizer = autorest.NewBearerAuthorizer(token)

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

