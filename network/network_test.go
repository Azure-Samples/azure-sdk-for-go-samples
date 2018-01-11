package network

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/subosito/gotenv"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

var (
	virtualNetworkName = "vnet1"
	subnet1Name        = "subnet1"
	subnet2Name        = "subnet2"
	nsgName            = "nsg1"
	nicName            = "nic1"
	ipName             = "ip1"
)

func TestMain(m *testing.M) {
	err := parseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}

	ctx := context.Background()
	defer resources.Cleanup(ctx)

	_, err = resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog(fmt.Sprintf("resource group created on location: %s", helpers.Location()))

	os.Exit(m.Run())
}

func parseArgs() error {
	gotenv.Load()

	virtualNetworkName = os.Getenv("AZ_VNET_NAME")
	flag.StringVar(&virtualNetworkName, "vnetName", virtualNetworkName, "Specify a name for the vnet.")

	err := helpers.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}

	if !(len(virtualNetworkName) > 0) {
		virtualNetworkName = "vnet1"
	}

	return nil
}

func ExampleCreateNIC() {
	ctx := context.Background()

	_, err := CreateVirtualNetworkAndSubnets(ctx, virtualNetworkName, subnet1Name, subnet2Name)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created vnet and 2 subnets")

	_, err = CreateNetworkSecurityGroup(ctx, nsgName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created network security group")

	_, err = CreatePublicIP(ctx, ipName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created public IP")

	_, err = CreateNIC(ctx, virtualNetworkName, subnet1Name, nsgName, ipName, nicName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("created nic")

	// Output:
	// created vnet and 2 subnets
	// created network security group
	// created public IP
	// created nic
}
