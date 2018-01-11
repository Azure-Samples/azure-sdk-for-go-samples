package containerservice

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
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

func init() {
	err := parseArgs()
	if err != nil {
		log.Fatalf("cannot parse arguments: %v", err)
	}
}

func parseArgs() error {
	err := helpers.ParseArgs()
	if err != nil {
		return fmt.Errorf("cannot parse args: %v", err)
	}

	resourceName = os.Getenv("AZ_RESOURCE_GROUP_NAME")
	if !(len(resourceName) > 0) {
		resourceName = "az-samples-go-aks-" + helpers.GetRandomLetterSequence(10)
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

	return nil
}

func ExampleCreateAKS() {
	_, err := resources.CreateGroup(context.Background(), helpers.ResourceGroupName())
	if err != nil {
		log.Printf("cannot create resource group: %v", err)
	}
	helpers.PrintAndLog("created resource group")

	_, err = CreateAKS(context.Background(), resourceName, helpers.Location(), helpers.ResourceGroupName(), username, sshPublicKeyPath, clientID, clientSecret, agentPoolCount)
	if err != nil {
		log.Fatalf("cannot create AKS cluster: %v", err)
	}

	helpers.PrintAndLog("created AKS cluster")

	// Output:
	// created resource group
	// created AKS cluster
}

func ExampleGetAKS() {
	c, err := GetAKS(context.Background(), helpers.ResourceGroupName(), resourceName)
	if err != nil {
		log.Fatalf("cannot get AKS cluster %v from resource group %v", resourceName, helpers.ResourceGroupName())
	}

	if *c.Name != resourceName {
		log.Fatalf("incorrect name of AKS cluster: expected %v, got %v", resourceName, *c.Name)
	}

	helpers.PrintAndLog("retrieved AKS cluster")

	// Output:
	// retrieved AKS cluster
}

func ExampleDeleteAKS() {
	defer resources.Cleanup(context.Background())

	_, err := DeleteAKS(context.Background(), helpers.ResourceGroupName(), resourceName)
	if err != nil {
		log.Fatalf("cannot delete AKS cluster %v from resource group %v: %v", resourceName, helpers.ResourceGroupName(), err)
	}

	helpers.PrintAndLog("deleted AKS cluster")

	// Output:
	// deleted AKS cluster
}
