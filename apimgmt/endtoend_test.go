// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package apimgmt

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	api "github.com/Azure/azure-sdk-for-go/services/apimanagement/mgmt/2019-01-01/apimanagement"
)

const (
	demoExistingEndpoint string = "https://raw.githubusercontent.com/OAI/OpenAPI-Specification/master/examples/v3.0/api-with-examples.yaml"
)

// TestEndToEnd tests creating and delete API Mgmt svcs
func TestEndToEnd(t *testing.T) {

	// skip this test for now due to length of time constraints, comment out to execute this test
	//t.SkipNow()

	var groupName = config.GenerateGroupName("APIMSTest")
	config.SetGroupName(groupName)

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Minute*60)
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, config.GroupName())
	if err != nil {
		util.PrintAndLog(err.Error())
		t.FailNow()
	}

	// create the ServiceInfo struct with relevant info
	serviceInfo := ServiceInfo{
		Ctx:               ctx,
		ResourceGroupName: groupName,
		ServiceName:       generateName("apimsvc"),
		Email:             "test@microsoft.com",
		Name:              "test",
	}

	// Wait for the service to be created, then only proceed once activated...
	// NOTE: creating the service can take any where from 15 mins to ~ 1hr
	// and this code will block the entire time
	util.PrintAndLog("creating api management service...")
	_, err = CreateAPIMgmtSvc(serviceInfo)
	if err != nil {
		util.PrintAndLog(fmt.Sprintf("cannot create api management service: %v", err))
		t.FailNow()
	}
	for true {
		time.Sleep(time.Second)
		activated, err := IsAPIMgmtSvcActivated(serviceInfo)
		if err != nil {
			util.PrintAndLog(fmt.Sprintf("error checking for activation: %v", err))
			t.FailNow()
			break
		}
		if activated == true {
			util.PrintAndLog("api management service created")
			break
		}
	}

	// create an api endpoint
	apiName := generateName("apiname")
	existingEndpoint := demoExistingEndpoint
	protocols := []api.Protocol{api.ProtocolHTTP, api.ProtocolHTTPS}
	path := "/testpath"
	apiid := generateName("apiid")

	// initialize an APICreateOrUpdateParameter object
	apiProperties := api.APICreateOrUpdateParameter{
		APICreateOrUpdateProperties: &api.APICreateOrUpdateProperties{
			Format:      api.OpenapiLink,
			DisplayName: &apiName,
			Value:       &existingEndpoint,
			Protocols:   &protocols,
			Path:        &path,
		},
	}

	// create the api endpoint
	util.PrintAndLog("creating api endpoint...")
	_, err = CreateOrUpdateAPI(serviceInfo, apiProperties, apiid, "")
	if err != nil {
		util.PrintAndLog(fmt.Sprintf("cannot create api: %v", err))
		t.FailNow()
	} else {
		util.PrintAndLog("open api endpoint created")
	}

	// delete the api endpoint
	util.PrintAndLog("deleting the api endpoint...")
	_, err = DeleteAPI(serviceInfo, apiid, "")
	if err != nil {
		util.PrintAndLog(fmt.Sprintf("cannot delete api: %v", err))
		t.FailNow()
	} else {
		util.PrintAndLog("open api endpoint deleted")
	}

	// delete the service
	_, err = DeleteAPIMgmtSvc(serviceInfo)
	if err != nil {
		util.PrintAndLog(fmt.Sprintf("cannot delete api management service: %v", err))
		t.FailNow()
	} else {
		util.PrintAndLog("api management service deleted")
	}

	// finished with context, cancel
	cancelFunc()
}
