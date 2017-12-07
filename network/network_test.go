package network

import (
	"flag"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/common"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/management"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/marstr/randname"
)

var (
	virtualNetworkName string
	subnet1Name        = "subnet" + randname.AdjNoun{}.Generate()
	subnet2Name        = "subnet" + randname.AdjNoun{}.Generate()
	nsgName            = "nsg" + randname.AdjNoun{}.Generate()
	nicName            = "nic" + randname.AdjNoun{}.Generate()
	ipName             = "ip" + randname.AdjNoun{}.Generate()
)

func init() {
	management.GetStartParams()
	flag.StringVar(&virtualNetworkName, "vNetName", "vnet"+randname.AdjNoun{}.Generate(), "Provide a name for the virtual network to be created")
	flag.Parse()
}

func ExampleCreateNIC() {
	defer resources.Cleanup()

	_, err := resources.CreateGroup()
	if err != nil {
		common.PrintAndLog(err.Error())
	}
	common.PrintAndLog("resource group created")

	_, errC := CreateVirtualNetworkAndSubnets(virtualNetworkName, subnet1Name, subnet2Name)
	err = <-errC
	if err != nil {
		common.PrintAndLog(err.Error())
	}
	common.PrintAndLog("created vnet and 2 subnets")

	_, errC = CreateNetworkSecurityGroup(nsgName)
	err = <-errC
	if err != nil {
		common.PrintAndLog(err.Error())
	}
	common.PrintAndLog("created network security group")

	_, errC = CreatePublicIp(ipName)
	err = <-errC
	if err != nil {
		common.PrintAndLog(err.Error())
	}
	common.PrintAndLog("created public IP")

	_, errC = CreateNic(virtualNetworkName, subnet1Name, nsgName, ipName, nicName)
	err = <-errC
	if err != nil {
		common.PrintAndLog(err.Error())
	}
	common.PrintAndLog("created nic")

	// Output:
	// resource group created
	// created vnet and 2 subnets
	// created network security group
	// created public IP
	// created nic
}
