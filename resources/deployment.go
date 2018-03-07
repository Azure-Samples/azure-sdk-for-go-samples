// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package resources

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest"
)

func getDeploymentsClient() resources.DeploymentsClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	deployClient := resources.NewDeploymentsClient(helpers.SubscriptionID())
	deployClient.Authorizer = autorest.NewBearerAuthorizer(token)
	deployClient.AddToUserAgent(helpers.UserAgent())
	return deployClient
}

// CreateDeployment creates a template deployment using the
// referenced JSON files for the template and its parameters
func CreateDeployment(ctx context.Context, deploymentName string, template, params *map[string]interface{}) (de resources.DeploymentExtended, err error) {
	deployClient := getDeploymentsClient()
	future, err := deployClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
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
		return de, fmt.Errorf("cannot create deployment: %v", err)
	}

	err = future.WaitForCompletion(ctx, deployClient.Client)
	if err != nil {
		return de, fmt.Errorf("cannot get the create deployment future respone: %v", err)
	}

	return future.Result(deployClient)
}

func ValidateDeployment(ctx context.Context, deploymentName string, template, params *map[string]interface{}) (valid resources.DeploymentValidateResult, err error) {
	deployClient := getDeploymentsClient()
	return deployClient.Validate(ctx,
		helpers.ResourceGroupName(),
		deploymentName,
		resources.Deployment{
			Properties: &resources.DeploymentProperties{
				Template:   template,
				Parameters: params,
				Mode:       resources.Incremental,
			},
		})
}
