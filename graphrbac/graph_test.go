// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package graphrbac

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/authorization"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

// TestMain sets up the environment and initiates tests.
func TestMain(m *testing.M) {
	var err error
	err = config.ParseEnvironment()
	if err != nil {
		log.Fatalf("failed to parse env: %v\n", err)
	}
	err = config.AddFlags()
	if err != nil {
		log.Fatalf("failed to parse flags: %v\n", err)
	}
	flag.Parse()

	code := m.Run()
	os.Exit(code)
}

func ExampleCreateServicePrincipal() {
	var groupName = config.GenerateGroupName("GraphRBAC")
	config.SetGroupName(groupName)

	ctx := context.Background()

	app, err := CreateADApplication(ctx)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("ad app created")

	sp, err := CreateServicePrincipal(ctx, *app.AppID)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("service principal created")

	_, err = AddClientSecret(ctx, *app.ObjectID)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("added client secret")

	_, err = resources.CreateGroup(ctx, config.GroupName())
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created resource group")

	list, err := authorization.ListRoleDefinitions(ctx, "roleName eq 'Contributor'")
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("list contributor roledefs at group scope")

	_, err = authorization.AssignRole(ctx, *sp.ObjectID, *list.Values()[0].ID)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("assigned new principal to first contributor role")

	if !config.KeepResources() {
		_, err = resources.DeleteGroup(ctx, config.GroupName())
		if err != nil {
			util.LogAndPanic(err)
		}

		_, err = DeleteADApplication(ctx, *app.ObjectID)
		if err != nil {
			util.LogAndPanic(err)
		}
	}

	// Output:
	// ad app created
	// service principal created
	// added client secret
	// created resource group
	// list contributor roledefs at group scope
	// assigned new principal to first contributor role
}

func ExampleCreateADGroup() {
	ctx := context.Background()

	group, err := CreateADGroup(ctx)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("ad group created")

	if !config.KeepResources() {
		_, err = DeleteADGroup(ctx, *group.ObjectID)
		if err != nil {
			util.LogAndPanic(err)
		}
		util.PrintAndLog("ad group deleted")
	}

	// Output:
	// ad group created
	// ad group deleted if KeepResources=false
}
