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

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
)

func TestMain(m *testing.M) {
	err := helpers.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}

	if !helpers.DeviceFlow() {
		helpers.PrintAndLog("It is best to run graph examples with device auth")
	} else {
		os.Exit(m.Run())
	}
}

func ExampleCreateServicePrincipal() {
	ctx := context.Background()

	app, err := CreateADApplication(ctx)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("ad app created")

	_, err = CreateServicePrincipal(ctx, *app.AppID)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("service principal created")

	_, err = DeleteADApplication(ctx, *app.ObjectID)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("ad app deleted")

	// Output:
	// ad app created
	// service principal created
	// list contributor role definition, sub scope
	// create role definition with subscription scope
	// ad app deleted
}
