package compute

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/network"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/marstr/randname"
)

func generateName(prefix string) string {
	return strings.ToLower(randname.GenerateWithPrefix(prefix, 5))
}

func addLocalEnvAndParse() error {
	// parse env at top-level (also controls dotenv load)
	err := config.ParseEnvironment()
	if err != nil {
		return fmt.Errorf("failed to add top-level env: %v\n", err.Error())
	}

	return nil
}

func setup() error {
	var err error
	err = addLocalEnvAndParse()
	if err != nil {
		return err
	}
	return nil
}

func Teardown() error {
	if config.KeepResources() == false {
		// does not wait
		_, err := resources.DeleteGroup(context.Background(), config.GroupName())
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateResourceGroup_() {
	setup()
	var groupName = config.GenerateGroupName("VM")
	// TODO: remove and use local `groupName` only
	config.SetGroupName(groupName)

	ctx, _ := context.WithTimeout(context.Background(), 6000*time.Second)

	// defer cancel()
	// defer resources.Cleanup(ctx)

	group, err := resources.CreateGroup(ctx, groupName)
	if err != nil {
		log.Println(err.Error())
	}
	fmt.Printf("created a resource group with name %s", *group.Name)
}

func CreateVirtualNetworksAndSubnet_(groupName string, virtualNetworkName string, subnet1Name string, subnet2Name string) {
	setup()
	config.SetGroupName(groupName)
	ctx, _ := context.WithTimeout(context.Background(), 6000*time.Second)
	_, err := network.CreateVirtualNetworkAndSubnets(ctx, virtualNetworkName, subnet1Name, subnet2Name)
	if err != nil {
		log.Println(err.Error())
	}
	log.Printf("created vnet %s", virtualNetworkName)
	log.Printf("created subnet %s", subnet1Name)
	log.Printf("created subnet %s", subnet2Name)
}

func CreateNetworkSecurityGroup_(groupName, networkSecurityGroupName string) {
	setup()
	config.SetGroupName(groupName)
	ctx, _ := context.WithTimeout(context.Background(), 6000*time.Second)
	_, err := network.CreateNetworkSecurityGroup(ctx, networkSecurityGroupName)
	if err != nil {
		log.Println(err.Error())
	}
	log.Printf("created network security group %s in %s", networkSecurityGroupName, groupName)
}

func CreatePublicIP_(groupName, ipName string) {
	setup()
	config.SetGroupName(groupName)
	ctx, _ := context.WithTimeout(context.Background(), 6000*time.Second)
	_, err := network.CreatePublicIP(ctx, ipName)
	if err != nil {
		log.Println(err.Error())
	}
	log.Printf("created public IP %s in %s", ipName, groupName)
}

func CreateNIC_(groupName, virtualNetworkName, subnet1Name, nsgName, ipName, nicName string) {
	setup()
	config.SetGroupName(groupName)
	ctx, _ := context.WithTimeout(context.Background(), 6000*time.Second)
	_, err := network.CreateNIC(ctx, virtualNetworkName, subnet1Name, nsgName, ipName, nicName)

	if err != nil {
		log.Println(err.Error())
	}
	log.Printf("created nic %s in %s", nicName, groupName)
}

func CreateVM_(groupName, vmName, nicName, username, password string) {
	setup()
	config.SetGroupName(groupName)
	ctx, _ := context.WithTimeout(context.Background(), 6000*time.Second)
	err := CreateVM(ctx, vmName, nicName, username, password)
	if err != nil {
		log.Println(err.Error())
	}
	log.Printf("created VM %s in %s", vmName, groupName)
}

func GetVM_(groupName, vmName string) {
	setup()
	config.SetGroupName(groupName)
	ctx, _ := context.WithTimeout(context.Background(), 6000*time.Second)
	vm, _ := GetVM(ctx, vmName)
	statuses := *vm.InstanceView.Statuses
	networkInterfaces := *vm.VirtualMachineProperties.NetworkProfile.NetworkInterfaces
	fmt.Println(*vm.VirtualMachineProperties.ProvisioningState)
	fmt.Println(*statuses[len(statuses)-1].DisplayStatus)
	fmt.Println(*networkInterfaces[0].ID)
	nic, _ := network.GetNIC(*networkInterfaces[0].ID)
}
