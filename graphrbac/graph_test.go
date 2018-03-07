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
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func TestMain(m *testing.M) {
	err := internal.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}

	if !internal.DeviceFlow() {
		internal.PrintAndLog("It is best to run graph examples with device auth")
	} else {
		iam.UseCLIclientID = true
		os.Exit(m.Run())
	}
}

func ExampleCreateServicePrincipal() {
	ctx := context.Background()

	app, err := CreateADApplication(ctx)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("ad app created")

	sp, err := CreateServicePrincipal(ctx, *app.AppID)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("service principal created")

	_, err = AddClientSecret(ctx, *app.ObjectID)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("added client secret")

	internal.SetResourceGroupName("CreateServicePrincipal")
	_, err = resources.CreateGroup(ctx, internal.ResourceGroupName())
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("created resource group")

	list, err := authorization.ListRoles(ctx, "roleName eq 'Contributor'")
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("list contributor role definition, with resource group scope")

	_, err = authorization.AssignRole(ctx, *sp.ObjectID, *list.Values()[0].ID)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("create role definition")

	if !internal.KeepResources() {
		_, err = resources.DeleteGroup(ctx, internal.ResourceGroupName())
		if err != nil {
			internal.PrintAndLog(err.Error())
		}

		_, err = DeleteADApplication(ctx, *app.ObjectID)
		if err != nil {
			internal.PrintAndLog(err.Error())
		}
	}

	// Output:
	// ad app created
	// service principal created
	// added client secret
	// created resource group
	// list contributor role definition, with resource group scope
	// create role definition
}
