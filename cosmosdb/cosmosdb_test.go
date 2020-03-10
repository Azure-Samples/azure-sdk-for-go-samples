// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cosmosdb

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mongodb"
	"github.com/globalsign/mgo/bson"
	"github.com/marstr/randname"
)

var (
	accountName = strings.ToLower(randname.GenerateWithPrefix("gosdksamplescosmos", 10))
)

// TestMain sets up the environment and initiates tests.
func TestMain(m *testing.M) {
	var err error
	err = config.ParseEnvironment()
	if err != nil {
		log.Fatalf("failed to parse env: %v\n", err)
	}
	err = config.AddFlags()
	if err != nil {
		log.Fatalf("failed to parse env: %v\n", err)
	}
	flag.Parse()

	code := m.Run()
	os.Exit(code)
}

func Example_cosmosDBOperations() {
	var groupName = config.GenerateGroupName("CosmosDB")
	config.SetGroupName(groupName)

	ctx := context.Background()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, config.GroupName())
	if err != nil {
		util.LogAndPanic(err)
	}

	_, err = CreateDatabaseAccount(ctx, accountName)
	if err != nil {
		util.LogAndPanic(fmt.Errorf("cannot create database account: %+v", err))
	}
	util.PrintAndLog("database account created")

	keys, err := ListKeys(ctx, accountName)
	if err != nil {
		util.LogAndPanic(fmt.Errorf("cannot list keys: %+v", err))
	}
	util.PrintAndLog("keys listed")

	host := fmt.Sprintf("%s.documents.azure.com", accountName)
	collection := "Packages"

	session, err := mongodb.NewMongoDBClientWithCredentials(accountName, *keys.PrimaryMasterKey, host)
	if err != nil {
		util.LogAndPanic(fmt.Errorf("cannot get mongoDB session: %+v", err))
	}
	util.PrintAndLog("got mongoDB session")

	GetCollection(session, accountName, collection)
	util.PrintAndLog("got collection")

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
		util.LogAndPanic(fmt.Errorf("cannot insert document: %v", err))
	}
	util.PrintAndLog("inserted document")

	doc, err := GetDocument(
		session,
		accountName,
		collection,
		bson.M{"fullname": "react"})
	if err != nil {
		util.LogAndPanic(fmt.Errorf("cannot get document: %v", err))
	}
	util.PrintAndLog("got document")
	util.PrintAndLog(fmt.Sprintf("document description: %s", doc["description"]))

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
		util.LogAndPanic(fmt.Errorf("cannot update document: %v", err))
	}
	util.PrintAndLog("update document")

	err = DeleteDocument(session, accountName, collection, doc["_id"].(bson.ObjectId))
	if err != nil {
		util.LogAndPanic(fmt.Errorf("cannot delete document: %v", err))
	}
	util.PrintAndLog("delete document")

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
