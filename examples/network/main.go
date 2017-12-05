package main

import (
	"flag"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/common"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/management"
	"github.com/Azure/go-autorest/autorest"
)

var (
	virtualNetworkName string
	subnet1Name        = "subnet1"
	subnet2Name        = "subnet2"
	nsgName            = "basic_services"
	nicName            = "nic1"
	ipName             = "ip1"
)

func init() {
	management.GetStartParams()
	flag.StringVar(&virtualNetworkName, "vNetName", "vnetname", "Provide a name for the virtual network to be created")
	flag.Parse()
}

func main() {
	var err error
	var errC <-chan error

	group, err := management.CreateGroup()
	common.OnErrorFail(err, "failed to create group")
	log.Printf("group: %+v\n", group)

	network, errC := management.CreateVirtualNetworkAndSubnets(virtualNetworkName, subnet1Name, subnet2Name)
	common.OnErrorFail(<-errC, "failed to create network")
	log.Printf("vnet: %+v\n", <-network)

	nsg, errC := management.CreateNetworkSecurityGroup(nsgName)
	common.OnErrorFail(<-errC, "failed to create network security group")
	log.Printf("network security group: %+v\n", <-nsg)

	ip, errC := management.CreatePublicIp(ipName)
	common.OnErrorFail(<-errC, "failed to create ip address")
	log.Printf("ip address: %+v\n", <-ip)

	nic, errC := management.CreateNic(virtualNetworkName, subnet1Name, nsgName, ipName, nicName)
	common.OnErrorFail(<-errC, "failed to create NIC")
	log.Printf("nic: %+v\n", <-nic)

	management.KeepResourcesAndExit()
	log.Printf("going to delete all resources\n")

	var resC <-chan autorest.Response

	resC, errC = management.DeleteNic(nicName)
	common.OnErrorFail(<-errC, "failed to delete nic")
	log.Printf("nic deleted: %+v", <-resC)

	resC, errC = management.DeleteNetworkSecurityGroup(nsgName)
	common.OnErrorFail(<-errC, "failed to delete network security group")
	log.Printf("network security group deleted: %+v\n", <-resC)

	resC, errC = management.DeleteVirtualNetwork(virtualNetworkName)
	common.OnErrorFail(<-errC, "failed to delete vnet")
	log.Printf("virtual network deleted: %+v\n", <-resC)

	resC, errC = management.DeleteGroup()
	common.OnErrorFail(<-errC, "failed to delete group")
	log.Printf("group deleted: %+v\n", <-resC)
}
