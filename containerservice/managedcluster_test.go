package containerservice

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

var (
	resourceName     string
	username         = "azureuser"
	sshPublicKeyPath = os.Getenv("HOME") + "/.ssh/id_rsa.pub"
	clientID         string
	clientSecret     string
	agentPoolCount   int32
)

func TestMain(m *testing.M) {
	err := parseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}
	os.Exit(m.Run())
}

func parseArgs() error {
	err := internal.ParseArgs()
	if err != nil {
		return fmt.Errorf("cannot parse args: %v", err)
	}

	resourceName = os.Getenv("AZ_AKS_NAME")
	if !(len(resourceName) > 0) {
		resourceName = "az-samples-go-aks-" + internal.GetRandomLetterSequence(10)
	}

	clientID = os.Getenv("AZ_CLIENT_ID")
	clientSecret = os.Getenv("AZ_CLIENT_SECRET")

	apc := os.Getenv("AZ_AKS_AGENTPOOLCOUNT")
	if !(len(apc) > 0) {
		agentPoolCount = int32(2)
	} else {
		i, _ := strconv.ParseInt(apc, 10, 32)
		agentPoolCount = int32(i)
	}

	internal.OverrideCanaryLocation("eastus2euap")

	// AKS managed clusters are not yet available in many Azure locations
	internal.OverrideLocation([]string{
		"eastus",
		"westeurope",
		"centralus",
		"canadacentral",
		"canadaeast",
	})
	return nil
}

func ExampleCreateAKS() {
	internal.SetResourceGroupName("CreateAKS")
	ctx := context.Background()
	defer resources.Cleanup(ctx)
	_, err := resources.CreateGroup(ctx, internal.ResourceGroupName())
	if err != nil {
		internal.PrintAndLog(err.Error())
	}

	_, err = CreateAKS(ctx, resourceName, internal.Location(), internal.ResourceGroupName(), username, sshPublicKeyPath, clientID, clientSecret, agentPoolCount)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}

	internal.PrintAndLog("created AKS cluster")

	_, err = GetAKS(ctx, internal.ResourceGroupName(), resourceName)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}

	internal.PrintAndLog("retrieved AKS cluster")

	_, err = DeleteAKS(ctx, internal.ResourceGroupName(), resourceName)
	if err != nil {
		internal.PrintAndLog(err.Error())
	}

	internal.PrintAndLog("deleted AKS cluster")

	// Output:
	// created AKS cluster
	// retrieved AKS cluster
	// deleted AKS cluster
}
