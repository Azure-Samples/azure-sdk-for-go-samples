// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/marstr/randname"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func addLocalEnvAndParse() error {
	// parse env at top-level (also controls dotenv load)
	err := config.ParseEnvironment()
	if err != nil {
		return fmt.Errorf("failed to add top-level env: %+v", err)
	}

	// add local env
	accountNameFromEnv := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	accountGroupNameFromEnv := os.Getenv("AZURE_STORAGE_ACCOUNT_GROUP_NAME")
	if len(accountNameFromEnv) > 0 {
		testAccountName = accountNameFromEnv
	}
	if len(accountGroupNameFromEnv) > 0 {
		testAccountGroupName = accountGroupNameFromEnv
	}
	return nil
}

func addLocalFlagsAndParse() error {
	// add top-level flags
	err := config.AddFlags()
	if err != nil {
		return fmt.Errorf("failed to add top-level flags: %+v", err)
	}

	// add local flags
	flag.StringVar(
		&testAccountName, "storageAccountName", testAccountName,
		"Name for test storage account.")
	flag.StringVar(
		&testAccountGroupName, "storageAccountGroupName", testAccountGroupName,
		"Name for the storage account group.")

	// parse all flags
	flag.Parse()
	return nil
}

func setup() error {
	var err error
	err = addLocalEnvAndParse()
	if err != nil {
		return err
	}
	err = addLocalFlagsAndParse()
	if err != nil {
		return err
	}

	if len(testAccountName) == 0 {
		testAccountName = generateName("gosdksamplestest")
	}
	if len(testAccountGroupName) == 0 {
		testAccountGroupName = config.GenerateGroupName("storage")
	}

	return nil
}

func teardown() error {
	if config.KeepResources() == false {
		// does not wait
		_, err := resources.DeleteGroup(context.Background(), testAccountGroupName)
		if err != nil {
			return err
		}
	}
	return nil
}

// test helpers
func generateName(prefix string) string {
	return strings.ToLower(randname.GenerateWithPrefix(prefix, 5))
}

// TestMain sets up the environment and initiates tests.
func TestMain(m *testing.M) {
	var err error
	var code int

	err = setup()
	if err != nil {
		log.Fatalf("could not set up environment: %+v", err)
	}

	code = m.Run()

	err = teardown()
	if err != nil {
		log.Fatalf(
			"could not tear down environment: %v\n; original exit code: %v\n",
			err, code)
	}

	os.Exit(code)
}
