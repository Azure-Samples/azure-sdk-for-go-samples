// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package keyvault

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/graphrbac"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

var (
	vaultName string
)

func TestMain(m *testing.M) {
	flag.StringVar(&vaultName, "vaultName", getVaultName(), "Specify name of vault to create.")
	err := iam.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse IAM args")
	}

	os.Exit(m.Run())
}

func ExampleSetVaultPermissions() {
	helpers.SetResourceGroupName("SetVaultPermissions")
	ctx := context.Background()
	defer resources.Cleanup(ctx)
	_, err := resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	_, err = CreateVault(ctx, vaultName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("vault created")

	_, err = SetVaultPermissions(ctx, vaultName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("set vault permissions")

	// Output:
	// vault created
	// set vault permissions
}

func ExampleCreateKeyBundle() {
	helpers.SetResourceGroupName("CreateKeyBundle")
	ctx := context.Background()
	defer resources.Cleanup(ctx)
	_, err := resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	// If authenticating as a user, also add the user to the keyvault access policies
	userID := ""
	if iam.AuthGrantType() == iam.OAuthGrantTypeDeviceFlow {
		cu, err := graphrbac.GetCurrentUser(ctx)
		if err != nil {
			helpers.PrintAndLog(err.Error())
		}
		userID = *cu.ObjectID
	}

	vaultName := getVaultName()

	_, err = CreateComplexKeyVault(ctx, vaultName, userID)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created keyvault")

	_, err = CreateKeyBundle(ctx, vaultName)
	if err != nil {
		helpers.PrintAndLog(err.Error())

	}
	helpers.PrintAndLog("created key bundle")

	// Output:
	// created keyvault
	// created key bundle
}
