// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cosmos

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mongodb"
	"github.com/globalsign/mgo/bson"
)

var (
	accountName = "cosmos-db-account-samples-" + helpers.GetRandomLetterSequence(10)
)

func TestMain(m *testing.M) {
	flag.StringVar(&accountName, "cosmosDBAccountName", accountName, "Provide a name for the CosmosDB account to be created")

	err := iam.ParseArgs()
	if err != nil {
		log.Fatalf("failed to parse IAM args: %v\n", err)
	}
	os.Exit(m.Run())
}

func ExampleCosmosDBOperations() {
	accountName := strings.ToLower(accountName)
	helpers.SetResourceGroupName("CosmosDBOperations")
	ctx := context.Background()
	defer resources.Cleanup(ctx)
	_, err := resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	_, err = CreateDatabaseAccount(ctx, accountName)
	if err != nil {
		helpers.PrintAndLog(fmt.Sprintf("cannot create database account: %v", err))
	}
	helpers.PrintAndLog("database account created")

	keys, err := ListKeys(ctx, accountName)
	if err != nil {
		helpers.PrintAndLog(fmt.Sprintf("cannot list keys: %v", err))
	}
	helpers.PrintAndLog("keys listed")

	host := fmt.Sprintf("%s.documents.azure.com", accountName)
	collection := "Packages"

	session, err := mongodb.NewMongoDBClientWithCredentials(accountName, *keys.PrimaryMasterKey, host)
	if err != nil {
		helpers.PrintAndLog(fmt.Sprintf("cannot get mongoDB session: %v", err))
	}
	helpers.PrintAndLog("got mongoDB session")

	GetCollection(session, accountName, collection)
	helpers.PrintAndLog("got collection")

	err = InsertDocument(
		session,
		accountName,
		collection,
		map[string]interface{}{
			"fullname":      "react",
			"description":   "A framework for building native apps with React.",
			"forksCount":    11392,
			"StarsCount":    48794,
			"LastUpdatedBy": "shergin",
		})
	if err != nil {
		helpers.PrintAndLog(fmt.Sprintf("cannot insert document: %v", err))
	}
	helpers.PrintAndLog("inserted document")

	doc, err := GetDocument(
		session,
		accountName,
		collection,
		bson.M{"fullname": "react"})
	if err != nil {
		helpers.PrintAndLog(fmt.Sprintf("cannot get document: %v", err))
	}
	helpers.PrintAndLog("got document")
	helpers.PrintAndLog(fmt.Sprintf("document description: %s", doc["description"]))

	err = UpdateDocument(
		session,
		accountName,
		collection,
		doc["_id"].(bson.ObjectId),
		bson.M{
			"$set": bson.M{
				"fullname": "react-native",
			},
		})
	if err != nil {
		helpers.PrintAndLog(fmt.Sprintf("cannot update document: %v", err))
	}
	helpers.PrintAndLog("update document")

	err = DeleteDcoument(session, accountName, collection, doc["_id"].(bson.ObjectId))
	if err != nil {
		helpers.PrintAndLog(fmt.Sprintf("cannot delete document: %v", err))
	}
	helpers.PrintAndLog("delete document")

	// Output:
	// database account created
	// keys listed
	// got mongoDB session
	// got collection
	// inserted document
	// got document
	// document description: A framework for building native apps with React.
	// update document
	// delete document
}
