// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/marstr/randname"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

var (
	// names used in tests
	username           = "gosdkuser"
	password           = "gosdkuserpass!1"
	vmName             = generateName("gosdk-vm1")
	vmssName           = generateName("gosdk-vmss1")
	diskName           = generateName("gosdk-disk1")
	nicName            = generateName("gosdk-nic1")
	virtualNetworkName = generateName("gosdk-vnet1")
	subnet1Name        = generateName("gosdk-subnet1")
	subnet2Name        = generateName("gosdk-subnet2")
	nsgName            = generateName("gosdk-nsg1")
	ipName             = generateName("gosdk-ip1")
	lbName             = generateName("gosdk-lb1")

	sshPublicKeyPath = os.Getenv("HOME") + "/.ssh/id_rsa.pub"

	containerGroupName        = randname.GenerateWithPrefix("gosdk-aci-", 10)
	aksClusterName            = randname.GenerateWithPrefix("gosdk-aks-", 10)
	aksUsername               = "azureuser"
	aksSSHPublicKeyPath       = os.Getenv("HOME") + "/.ssh/id_rsa.pub"
	aksAgentPoolCount   int32 = 4
)

func addLocalEnvAndParse() error {
	// parse env at top-level (also controls dotenv load)
	err := config.ParseEnvironment()
	if err != nil {
		return fmt.Errorf("failed to add top-level env: %+v", err)
	}

	// add local env
	vnetNameFromEnv := os.Getenv("AZURE_VNET_NAME")
	if len(vnetNameFromEnv) > 0 {
		virtualNetworkName = vnetNameFromEnv
	}
	return nil
}

func addLocalFlagsAndParse() error {
	// add top-level flags
	err := config.AddFlags()
	if err != nil {
		return fmt.Errorf("failed to add top-level flags: %+v", err)
	}

	// add local flags
	// flag.StringVar(
	//	&testVnetName, "testVnetName", testVnetName,
	//	"Name for test Vnet.")

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

// test helpers
func generateName(prefix string) string {
	return strings.ToLower(randname.GenerateWithPrefix(prefix, 5))
}

// TestMain sets up the environment and initiates tests.
func TestMain(m *testing.M) {
	var err error
	var code int

	err = setup()
	if err != nil {
		log.Fatalf("could not set up environment: %v\n", err)
	}

	code = m.Run()

	err = teardown()
	if err != nil {
		log.Fatalf(
			"could not tear down environment: %v\n; original exit code: %v\n",
			err, code)
	}

	os.Exit(code)
}
