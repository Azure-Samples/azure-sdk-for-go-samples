// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package web

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

var (
	appServicePlanName = "web-appserviceplan-go-samples" + helpers.GetRandomLetterSequence(10)
)

func TestMain(m *testing.M) {
	flag.StringVar(&appServicePlanName, "appServicePlanName", appServicePlanName, "Optionally provide a name for the App Service Plan to be Created.")

	err := iam.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse IAM args")
	}

	os.Exit(m.Run())
}

func ExampleWeb_DeployAppForContainer() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	helpers.SetResourceGroupName("WebAppForContainers")
	_, err := resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		fmt.Println("failed to create resource group: ", err)
		return
	}
	defer resources.Cleanup(ctx)

	err = CreateSite(ctx, "myThing")

	if err != nil {
		fmt.Println("failed to create: ", err)
		return
	}

	// Output: lol, this won't be the output
}
