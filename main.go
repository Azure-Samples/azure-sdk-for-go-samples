package main

import (
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/compute"
)

func main1() {
	fmt.Println("running vm test logic")
	compute.CreateResourceGroup_()
}

func main2() {
	fmt.Println("creating virtual network")
	compute.CreateVirtualNetworksAndSubnet_("az-samples-go-VM-ps240", "az-samples-go-VM-ps240-vnet", "az-samples-go-VM-ps240-subnet1", "az-samples-go-VM-ps240-subnet2")
}

func main3() {
	fmt.Println("creating network security group")
	compute.CreateNetworkSecurityGroup_("az-samples-go-VM-ps240", "az-samples-go-VM-ps240-nsg1")
}

func main4() {
	fmt.Println("creating public ip")
	compute.CreatePublicIP_("az-samples-go-VM-ps240", "az-samples-go-VM-ps240-ip1")
}

func main5() {
	fmt.Println("creating nic")
	compute.CreateNIC_("az-samples-go-VM-ps240", "az-samples-go-VM-ps240-vnet", "az-samples-go-VM-ps240-subnet1", "az-samples-go-VM-ps240-nsg1", "az-samples-go-VM-ps240-ip1", "az-samples-go-VM-ps240-nic1")
}

func main6() {
	fmt.Println("creating vm")
	compute.CreateVM_("az-samples-go-VM-ps240", "az-samples-go-VM-ps240-vm1", "az-samples-go-VM-ps240-nic1", "umarmuneer-admin", "admin#@1")
}

func main() {
	fmt.Println("getting vm")
	compute.GetVM_("az-samples-go-VM-x980I", "az-samples-go-VM-x980I-vm1")
}
