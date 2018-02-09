// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package authorization

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func TestMain(m *testing.M) {
	err := helpers.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}

	ctx := context.Background()
	defer resources.Cleanup(ctx)

	_, err = resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog(fmt.Sprintf("resource group created on location: %s", helpers.Location()))

	os.Exit(m.Run())
}

func ExampleAssignRole() {
	ctx := context.Background()

	list, err := ListRoles(ctx, "roleName eq 'Contributor'")
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("got role definitions list")

	rgRole, err := AssignRole(ctx, helpers.ServicePrincipalObjectID(), *((*list.Value)[0].ID))
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("role assigned with resource group scope")

	subRole, err := AssignRoleWithSubscriptionScope(ctx, helpers.ServicePrincipalObjectID(), *((*list.Value)[0].ID))
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("role assigned with subscription scope")

	if !helpers.KeepResources() {
		DeleteRoleAssignment(*rgRole.ID)
		if err != nil {
			helpers.PrintAndLog(err.Error())
		}

		DeleteRoleAssignment(*subRole.ID)
		if err != nil {
			helpers.PrintAndLog(err.Error())
		}
	}

	// Output:
	// got role definitions list
	// role assigned with resource group scope
	// role assigned with subscription scope
}
