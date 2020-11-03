// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package resources

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
)

func Example_createTemplateDeployment() {
	groupName := config.GenerateGroupName("groups-template")
	config.SetGroupName(groupName) // TODO: don't rely on globals
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	defer Cleanup(ctx)

	_, err := CreateGroup(ctx, config.GroupName())
	if err != nil {
		util.LogAndPanic(err)
	}

	wd, _ := os.Getwd()
	templateFile := filepath.Join(wd, "testdata", "template.json")
	parametersFile := filepath.Join(wd, "testdata", "parameters.json")
	deployName := "VMdeploy"

	template, err := util.ReadJSON(templateFile)
	if err != nil {
		return
	}
	params, err := util.ReadJSON(parametersFile)
	if err != nil {
		return
	}

	_, err = ValidateDeployment(ctx, deployName, template, params)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("validated VM template deployment")

	_, err = CreateDeployment(ctx, deployName, template, params)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created VM template deployment")

	ipName := (*params)["publicIPAddresses_QuickstartVM_ip_name"].(map[string]interface{})["value"].(string)
	vmUser := (*params)["vm_user"].(map[string]interface{})["value"].(string)
	vmPass := (*params)["vm_password"].(map[string]interface{})["value"].(string)

	r, err := GetResource(ctx,
		"Microsoft.Network",
		"publicIPAddresses",
		ipName,
		"2018-01-01")
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("got public IP info via get generic resource")

	log.Printf("Log in with ssh: %s@%s, password: %s",
		vmUser,
		r.Properties.(map[string]interface{})["ipAddress"].(string),
		vmPass)

	// Output:
	// validated VM template deployment
	// created VM template deployment
	// got public IP info via get generic resource
}
