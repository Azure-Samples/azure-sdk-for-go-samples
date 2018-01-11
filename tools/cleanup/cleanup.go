package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func main() {
	var quiet bool
	flag.BoolVar(&quiet, "quiet", false, "Run quietly")

	err := helpers.ParseSubscriptionID()
	if err != nil {
		log.Fatalf("Error parsing subscriptionID: %v\n", err)
		os.Exit(1)
	}
	err = helpers.ParseDeviceFlow()
	if err != nil {
		log.Fatalf("Error parsing device flow: %v\n", err)
		log.Fatalf("Using device flow: %v", helpers.DeviceFlow())
	}
	flag.Parse()

	if !quiet {
		fmt.Println("Are you sure you want to delete all resource groups in the subscription? (yes | no)")
		var input string
		fmt.Scanln(&input)
		if input != "yes" {
			fmt.Println("Keeping resource groups")
			os.Exit(0)
		}
	}

	futures, groups := resources.DeleteAllGroupsWithPrefix(context.Background(), helpers.GroupPrefix())

	var wg sync.WaitGroup
	resources.WaitForDeleteCompletion(context.Background(), &wg, futures, groups)
	wg.Wait()

	fmt.Println("Done")
}
