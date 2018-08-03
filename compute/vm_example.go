package compute

import (
	"context"
	"fmt"
	"time"
	"log"
	"os"
	"strings"
	"github.com/marstr/randname"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

var (
	// names used in tests
	username           = "gosdkuser"
	password           = "gosdkuserpass!1"
	vmName             = generateName("gosdk-vm1")
	diskName           = generateName("gosdk-disk1")
	nicName            = generateName("gosdk-nic1")
	virtualNetworkName = generateName("gosdk-vnet1")
	subnet1Name        = generateName("gosdk-subnet1")
	subnet2Name        = generateName("gosdk-subnet2")
	nsgName            = generateName("gosdk-nsg1")
	ipName             = generateName("gosdk-ip1")
	lbName             = generateName("gosdk-lb1")

	sshPublicKeyPath = os.Getenv("HOME") + "/.ssh/id_rsa.pub"

	containerGroupName  string = randname.GenerateWithPrefix("gosdk-aci-", 10)
	aksClusterName      string = randname.GenerateWithPrefix("gosdk-aks-", 10)
	aksUsername         string = "azureuser"
	aksSSHPublicKeyPath string = os.Getenv("HOME") + "/.ssh/id_rsa.pub"
	aksAgentPoolCount   int32  = 4
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

	// add local env
	vnetNameFromEnv := os.Getenv("AZURE_VNET_NAME")
	if len(vnetNameFromEnv) > 0 {
		virtualNetworkName = vnetNameFromEnv
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

func CreateResourceGroup() {
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