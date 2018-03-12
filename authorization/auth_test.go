// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package authorization

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func TestMain(m *testing.M) {
	err := internal.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}

	os.Exit(m.Run())
}

func ExampleAssignRole() {
	internal.SetResourceGroupName("AssignRole")
	ctx := context.Background()
	defer resources.Cleanup(ctx)
	_, err := resources.CreateGroup(ctx, internal.ResourceGroupName())
	if err != nil {
		internal.PrintAndLog(err.Error())
	}

	list, err := ListRoles(ctx, "roleName eq 'Contributor'")
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("got role definitions list")

	rgRole, err := AssignRole(ctx, internal.ServicePrincipalObjectID(), *list.Values()[0].ID)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("role assigned with resource group scope")

	subRole, err := AssignRoleWithSubscriptionScope(ctx, internal.ServicePrincipalObjectID(), *list.Values()[0].ID)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}
	internal.PrintAndLog("role assigned with subscription scope")

	if !internal.KeepResources() {
		DeleteRoleAssignment(ctx, *rgRole.ID)
		if err != nil {
			internal.PrintAndLog(err.Error())
		}

		DeleteRoleAssignment(ctx, *subRole.ID)
		if err != nil {
			internal.PrintAndLog(err.Error())
		}
	}

	// Output:
	// got role definitions list
	// role assigned with resource group scope
	// role assigned with subscription scope
}
