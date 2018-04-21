package containerservice

import (
	"context"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
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
		log.Fatalf("failed to parse container service args: %v\n", err)
	}

	err = iam.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse IAM args")
	}

	// AKS managed clusters are not yet available in many Azure locations
	helpers.OverrideCanaryLocation("eastus2euap")
	helpers.OverrideLocation([]string{
		"eastus",
		"westeurope",
		"centralus",
		"canadacentral",
		"canadaeast",
	})

	os.Exit(m.Run())
}

func parseArgs() error {
	resourceName = os.Getenv("AZURE_AKS_NAME")
	if !(len(resourceName) > 0) {
		resourceName = "az-samples-go-aks-" + helpers.GetRandomLetterSequence(10)
	}

	clientID = os.Getenv("AZURE_CLIENT_ID")
	clientSecret = os.Getenv("AZURE_CLIENT_SECRET")

	apc := os.Getenv("AZURE_AKS_AGENTPOOLCOUNT")
	if !(len(apc) > 0) {
		agentPoolCount = int32(2)
	} else {
		i, _ := strconv.ParseInt(apc, 10, 32)
		agentPoolCount = int32(i)
	}
	return nil
}

func ExampleCreateAKS() {
	helpers.SetResourceGroupName("CreateAKS")
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*30))
	defer cancel()
	defer resources.Cleanup(ctx)
	_, err := resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	_, err = CreateAKS(ctx, resourceName, helpers.Location(), helpers.ResourceGroupName(), username, sshPublicKeyPath, clientID, clientSecret, agentPoolCount)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	helpers.PrintAndLog("created AKS cluster")

	_, err = GetAKS(ctx, helpers.ResourceGroupName(), resourceName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	helpers.PrintAndLog("retrieved AKS cluster")

	_, err = DeleteAKS(ctx, helpers.ResourceGroupName(), resourceName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	helpers.PrintAndLog("deleted AKS cluster")

	// Output:
	// created AKS cluster
	// retrieved AKS cluster
	// deleted AKS cluster
}
