package network

import (
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func Example() {
	var err error
	var errC <-chan error

	defer resources.Cleanup()

	group, err := resources.CreateGroup()
	helpers.OnErrorFail(err, "failed to create group")
	log.Printf("group: %+v\n", group)

	network, errC := CreateVirtualNetwork()
	helpers.OnErrorFail(<-errC, "failed to create network")
	log.Printf("vnet: %+v\n", <-network)

	nsg, errC := CreateNetworkSecurityGroup()
	helpers.OnErrorFail(<-errC, "failed to create network security group")
	log.Printf("network security group: %+v\n", <-nsg)

	ip, errC := CreatePublicIp()
	helpers.OnErrorFail(<-errC, "failed to create ip address")
	log.Printf("ip address: %+v\n", <-ip)

	nic, errC := CreateNic()
	helpers.OnErrorFail(<-errC, "failed to create NIC")
	log.Printf("nic: %+v\n", <-nic)

	fmt.Println("Success")
	// Output: Success
}
