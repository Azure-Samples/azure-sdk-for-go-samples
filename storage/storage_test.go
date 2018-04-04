// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"flag"
	"log"
	"os"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/subosito/gotenv"
)

var (
	accountName   string
	containerName = "container1"
	blobName      = "blob1"
	messages      = []string{"Hello", "World!", "Hello", "Galaxy!"}
)

func TestMain(m *testing.M) {
	gotenv.Load()
	name := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	if len(name) > 0 {
		accountName = name
	}

	flag.StringVar(&accountName, "storageAccoutName", getAccountName(), "Provide a name for the storage account to be created")
	flag.StringVar(&containerName, "containerName", containerName, "Provide a name for the container.")
	flag.StringVar(&blobName, "blobName", blobName, "Provide a name for the blob.")

	err := helpers.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}

	os.Exit(m.Run())
}
