// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/marstr/randname"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

var (
	containerGroupName  string = randname.GenerateWithPrefix("gosdk-aci-", 10)
	aksClusterName      string = randname.GenerateWithPrefix("gosdk-aks-", 10)
	aksUsername         string = "azureuser"
	aksSSHPublicKeyPath string = os.Getenv("HOME") + "/.ssh/id_rsa.pub"
	aksAgentPoolCount   int32  = 4
)

func addLocalEnvAndParse() error {
	// parse env at top-level (also controls dotenv load)
	err := config.ParseEnvironment()
	if err != nil {
		return fmt.Errorf("failed to add top-level env: %v\n", err.Error())
	}
}

func addLocalFlagsAndParse() error {
	// add top-level flags
	err := config.AddFlags()
	if err != nil {
		return fmt.Errorf("failed to add top-level flags: %v\n", err.Error())
	}
	// parse all flags
	flag.Parse()
	return nil
}

func setup() error {
	var err error
	err = addLocalEnvAndParse()
	if err != nil {
		return err
	}
	err = addLocalFlagsAndParse()
	if err != nil {
		return err
	}

	return nil
}

func teardown() error {
	if config.KeepResources() == false {
		// does not wait
		_, err := resources.DeleteGroup(context.Background(), config.GroupName())
		if err != nil {
			return err
		}
	}
	return nil
}
