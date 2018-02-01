// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package graphrbac

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/authorization"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func TestMain(m *testing.M) {
	err := helpers.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}

	if !helpers.DeviceFlow() {
		helpers.PrintAndLog("It is best to run graph examples with device auth")
	} else {
		os.Exit(m.Run())
	}
}

func ExampleCreateServicePrincipal() {
	ctx := context.Background()

	app, err := CreateADApplication(ctx)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("ad app created")

	sp, err := CreateServicePrincipal(ctx, *app.AppID)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("service principal created")

	_, err = resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created resource group")

	list, err := authorization.ListRoles(ctx, "roleName eq 'Contributor'")
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("list contributor role definition, with resource group scope")

	_, err = authorization.AssignRole(ctx, *sp.ObjectID, *((*list.Value)[0].ID))
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("create role definition")

	_, err = resources.DeleteGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("deleted resource group")

	_, err = DeleteADApplication(ctx, *app.ObjectID)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("ad app deleted")

	// Output:
	// ad app created
	// service principal created
	// created resource group
	// list contributor role definition, with resource group scope
	// create role definition
	// deleted resource group
	// ad app deleted
}
