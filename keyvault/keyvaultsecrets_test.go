// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package keyvault

import (
	"context"
    "os"
    "fmt"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/keyvault"
	kvauth "github.com/Azure/azure-sdk-for-go/services/keyvault/auth"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func ExampleKVSecrets() {
	var groupName = config.GenerateGroupName("KeyVault")
	config.SetGroupName(groupName)

	ctx := context.Background()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, config.GroupName())
	if err != nil {
		util.LogAndPanic(err)
	}

    fmt.Printf("group name: %s\n", groupName)
	util.PrintAndLog("resource group created\n")

	_, err = CreateVault(ctx, kvName)
	if err != nil {
		util.LogAndPanic(err)
	}

    fmt.Printf("vault name: %s",kvName)
	util.PrintAndLog("vault created")

	_, err = SetVaultPermissions(ctx, kvName)
	if err != nil {
		util.LogAndPanic(err)
	}

	util.PrintAndLog("set vault permissions")

	_, err = CreateKey(ctx, kvName, keyName)
	if err != nil {
		util.LogAndPanic(err)
	}

    fmt.Printf("key name: %s \n", keyName)
	util.PrintAndLog("created key")

    // Client
	authorizer, err := kvauth.NewAuthorizerFromEnvironment()
	if err != nil {
		util.PrintAndLog("unable to create vault authorizer\n")
		os.Exit(1)
	}

	basicClient := keyvault.New()
	basicClient.Authorizer = authorizer

	if *setDebug {
		basicClient.RequestInspector = logRequest()
		basicClient.ResponseInspector = logResponse()
        fmt.Println("setDebug")
	}

    util.PrintAndLog("created basic client and vault authorizer\n")

	vaultName = kvName
    fmt.Sprintf("Setting vaultName to: %s\n", kvName)

    // Test
    util.PrintAndLog("list secrets\n")
	listSecrets(basicClient)

    util.PrintAndLog("create update secret\n")
	createUpdateSecret(basicClient, "TestSecret", "TestValue")

    util.PrintAndLog("get secret\n")
	getSecret(basicClient, "TestSecret")

    util.PrintAndLog("list secrets again\n")
	listSecrets(basicClient)

    util.PrintAndLog("delete secret\n")
	deleteSecret(basicClient, "TestSecret")

	// Output:
	// resource group created
	// vault created
	// set vault permissions
	// created key
    // created basic client and vault authorizer
    // list secrets
    // create update secret
    // get secret
    // list secrets again
    // delete secret
}
