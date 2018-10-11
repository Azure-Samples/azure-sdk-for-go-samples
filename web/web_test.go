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

	"github.com/marstr/randname"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

var (
	appServicePlanName = randname.GenerateWithPrefix("web-appserviceplan-go-samples", 10)
	siteName           = randname.GenerateWithPrefix("web-site-go-samples", 10)
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
		log.Fatalf("failed to parse env: %v\n", err)
	}
	flag.Parse()

	code := m.Run()
	os.Exit(code)
}

func ExampleWeb_DeployAppForContainer() {
	var groupName = config.GenerateGroupName("WebAppForContainers")
	config.SetGroupName(groupName)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	_, err := resources.CreateGroup(ctx, config.GroupName())
	if err != nil {
		fmt.Println("failed to create resource group: ", err)
		return
	}
	defer resources.Cleanup(ctx)

	_, err = CreateContainerSite(ctx, siteName, "appsvc/sample-hello-world:latest")

	if err != nil {
		fmt.Println("failed to create: ", err)
		return
	}

	configResource, err := GetAppConfiguration(ctx, siteName)
	if err != nil {
		fmt.Println("failed to get app configuration: ", err)
		return
	}
	fmt.Println(*configResource.LinuxFxVersion)

	// Output: DOCKER|appsvc/sample-hello-world:latest
}
