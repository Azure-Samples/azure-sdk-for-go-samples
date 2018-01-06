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
	err := helpers.ParseSubscriptionID()
	if err != nil {
		log.Fatalf("Error parsing subscriptionID: %v\n", err)
		os.Exit(1)
	}

	var quiet bool
	flag.BoolVar(&quiet, "quiet", false, "Run quietly")
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
	var wg sync.WaitGroup
	resources.CleanupAll(context.Background(), &wg)
	wg.Wait()
	fmt.Println("Done")
}
