package main

import (
	"log"
	"os"
	"github.com/joshgav/az-go/common"
	"github.com/joshgav/az-go/management"
  "github.com/Azure/go-autorest/autorest"
)

func main() {
	var err error
	var errC <-chan error

	group, err := management.CreateGroup()
	common.OnErrorFail(err, "failed to create group")
	log.Printf("group: %+v\n", group)

  network, errC := management.CreateVirtualNetwork()
  common.OnErrorFail(<-errC, "failed to create network")
  log.Printf("vnet: %+v\n", <-network)

  nsg, errC := management.CreateNetworkSecurityGroup()
  common.OnErrorFail(<-errC, "failed to create network security group")
  log.Printf("network security group: %+v\n", <-nsg)

  ip, errC := management.CreatePublicIp()
  common.OnErrorFail(<-errC, "failed to create ip address")
  log.Printf("ip address: %+v\n", <-ip)

  nic, errC := management.CreateNic()
  common.OnErrorFail(<-errC, "failed to create NIC")
  log.Printf("nic: %+v\n", <-nic)

	if os.Getenv("AZURE_KEEP_SAMPLE_RESOURCES") == "1" {
    log.Printf("retaining resources because env var is set\n")
		os.Exit(0)
	}

	log.Printf("going to delete all resources\n")

  // var res autorest.Response
  var resC <-chan autorest.Response

  resC, errC = management.DeleteNic()
  common.OnErrorFail(<-errC, "failed to delete nic")
  log.Printf("nic deleted: %+v", <-resC)

	resC, errC = management.DeleteNetworkSecurityGroup()
	common.OnErrorFail(<-errC, "failed to delete network security group")
	log.Printf("network security group deleted: %+v\n", <-resC) 

  resC, errC = management.DeleteVirtualNetwork()
  common.OnErrorFail(<-errC, "failed to delete vnet")
  log.Printf("virtual network deleted: %+v\n", <-resC)

	resC, errC = management.DeleteGroup()
	common.OnErrorFail(<-errC, "failed to delete group")
	log.Printf("group deleted: %+v\n", <-resC)
}
