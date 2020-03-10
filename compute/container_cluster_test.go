package compute

import (
	"context"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func ExampleCreateAKS() {
	var groupName = config.GenerateGroupName("CreateAKS")
	config.SetGroupName(groupName)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Hour*1))
	defer cancel()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, config.GroupName())
	if err != nil {
		util.LogAndPanic(err)
	}

	_, err = CreateAKS(ctx, aksClusterName, config.Location(), config.GroupName(), aksUsername, aksSSHPublicKeyPath, config.ClientID(), config.ClientSecret(), aksAgentPoolCount)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created AKS cluster")

	_, err = GetAKS(ctx, config.GroupName(), aksClusterName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved AKS cluster")

	_, err = DeleteAKS(ctx, config.GroupName(), aksClusterName)
	if err != nil {
		util.LogAndPanic(err)
	}

	util.PrintAndLog("deleted AKS cluster")

	// Output:
	// created AKS cluster
	// retrieved AKS cluster
	// deleted AKS cluster
}
