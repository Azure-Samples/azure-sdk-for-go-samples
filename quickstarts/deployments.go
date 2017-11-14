package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/common"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/management"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/satori/go.uuid"
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
	group, err := management.CreateGroup()
	if err != nil {
		log.Fatalf("failed to create group: %v", err)
	}
	log.Printf("created group: %v\n", group)

	log.Printf("starting deployment\n")
	result, errC := CreateDeployment()
	if errC != nil {
		log.Fatalf("failed to deploy: %v\n", <-errC)
	}
	log.Printf("started deployment: %v\n", (<-result).Properties)
}

func init() {
	gotenv.Load()
	tenantId = common.GetEnvVarOrFail("AZURE_TENANT_ID")
	subscriptionId = common.GetEnvVarOrFail("AZURE_SUBSCRIPTION_ID")
	clientId = common.GetEnvVarOrFail("AZURE_CLIENT_ID")
	clientSecret = common.GetEnvVarOrFail("AZURE_CLIENT_SECRET")
	resourceGroupName = common.GetEnvVarOrFail("AZURE_RG_NAME")
	deploymentName = fmt.Sprintf("template-deployment-%s", uuid.NewV4())
	resourceGroupLocation = common.GetEnvVarOrFail("AZURE_LOCATION")
	pathToTemplateFile, _ = filepath.Abs("test_data/template.json")
	pathToParametersFile, _ = filepath.Abs("test_data/parameters.json")

	// set up OAuth token using service principal credentials
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

// ReadJSON reads a file and unmarshals the JSON
func ReadJSON(path string) (*map[string]interface{}, error) {
	_json, err := ioutil.ReadFile(path)
	log.Printf("%s", string(_json[:]))
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

	_template, _ := ReadJSON(pathToTemplateFile)
	_params, _ := ReadJSON(pathToParametersFile)

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
