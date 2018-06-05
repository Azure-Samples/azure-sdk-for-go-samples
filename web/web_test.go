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
	siteName           = "web-site-go-samples" + helpers.GetRandomLetterSequence(10)
)

func TestMain(m *testing.M) {
	flag.StringVar(&appServicePlanName, "appServicePlanName", appServicePlanName, "Optionally provide a name for the App Service Plan to be created.")
	flag.StringVar(&siteName, "siteName", siteName, "Optionally provided a name for the Site to be created.")

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

	configResource, err := CreateContainerSite(ctx, siteName, "appsvc/sample-hello-world:latest")

	if err != nil {
		fmt.Println("failed to create: ", err)
		return
	}

	fmt.Println(*configResource.LinuxFxVersion)

	// Output: DOCKER|appsvc/sample-hello-world:latest
}
