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
	siteName = randname.GenerateWithPrefix("web-site-go-samples", 10)
)

// TestMain sets up the environment and initiates tests.
func TestMain(m *testing.M) {
	if err := config.ParseEnvironment(); err != nil {
		log.Fatalf("failed to parse env: %v\n", err)
	}
	if err := config.AddFlags(); err != nil {
		log.Fatalf("failed to add flags: %+v", err)
	}
	flag.Parse()

	code := m.Run()
	os.Exit(code)
}

func TestCreateApp(t *testing.T) {
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

	_, err = CreateWebApp(ctx, siteName)

	if err != nil {
		fmt.Println("failed to create: ", err)
		return
	}

	configResource, err := GetAppConfiguration(ctx, siteName)
	if err != nil {
		fmt.Println("failed to get app configuration: ", err)
		return
	}
	fmt.Println(*configResource.Name)
}
