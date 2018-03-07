// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package resources

import (
	"context"
	"go/build"
	"log"
	"path/filepath"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
)

func ExampleCreateTemplateDeployment() {
	helpers.SetResourceGroupName("CreateTemplateDeploy")
	ctx := context.Background()
	defer Cleanup(ctx)
	_, err := CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	gopath := build.Default.GOPATH
	repo := filepath.Join("github.com", "Azure-Samples", "azure-sdk-for-go-samples")
	templateFile := filepath.Join(gopath, "src", repo, "testdata", "template.json")
	parametersFile := filepath.Join(gopath, "src", repo, "testdata", "parameters.json")
	deployName := "VMdeploy"

	template, err := helpers.ReadJSON(templateFile)
	if err != nil {
		return
	}
	params, err := helpers.ReadJSON(parametersFile)
	if err != nil {
		return
	}

	_, err = ValidateDeployment(ctx, deployName, template, params)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("validated VM template deployment")

	_, err = CreateDeployment(ctx, deployName, template, params)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created VM template deployment")

	ipName := (*params)["publicIPAddresses_QuickstartVM_ip_name"].(map[string]interface{})["value"].(string)
	vmUser := (*params)["vm_user"].(map[string]interface{})["value"].(string)
	vmPass := (*params)["vm_password"].(map[string]interface{})["value"].(string)

	resource, err := GetResource(ctx,
		"Microsoft.Network",
		"publicIPAddresses",
		ipName,
		"2018-01-01")
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("got public IP info via get generic resource")

	log.Printf("Log in with ssh: %s@%s, password: %s",
		vmUser,
		resource.Properties.(map[string]interface{})["ipAddress"].(string),
		vmPass)

	// Output:
	// validated VM template deployment
	// created VM template deployment
	// got public IP info via get generic resource
}
