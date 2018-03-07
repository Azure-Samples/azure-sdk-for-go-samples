// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package sql

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/sql/mgmt/2015-05-01-preview/sql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

// Servers

func getServersClient() sql.ServersClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	serversClient := sql.NewServersClient(internal.SubscriptionID())
	serversClient.Authorizer = autorest.NewBearerAuthorizer(token)
	serversClient.AddToUserAgent(internal.UserAgent())
	return serversClient
}

// CreateServer creates a new SQL Server
func CreateServer(ctx context.Context, serverName, dbLogin, dbPassword string) (server sql.Server, err error) {
	serversClient := getServersClient()
	future, err := serversClient.CreateOrUpdate(
		ctx,
		internal.ResourceGroupName(),
		serverName,
		sql.Server{
			Location: to.StringPtr(internal.Location()),
			ServerProperties: &sql.ServerProperties{
				AdministratorLogin:         to.StringPtr(dbLogin),
				AdministratorLoginPassword: to.StringPtr(dbPassword),
			},
		})

	if err != nil {
		return server, fmt.Errorf("cannot create sql server: %v", err)
	}

	err = future.WaitForCompletion(ctx, serversClient.Client)
	if err != nil {
		return server, fmt.Errorf("cannot get the sql server create or update future response: %v", err)
	}

	return future.Result(serversClient)
}

// Databases

func getDbClient() sql.DatabasesClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	dbClient := sql.NewDatabasesClient(internal.SubscriptionID())
	dbClient.Authorizer = autorest.NewBearerAuthorizer(token)
	dbClient.AddToUserAgent(internal.UserAgent())
	return dbClient
}

// CreateDB creates a new SQL Database on a given server
func CreateDB(ctx context.Context, serverName, dbName string) (db sql.Database, err error) {
	dbClient := getDbClient()
	future, err := dbClient.CreateOrUpdate(
		ctx,
		internal.ResourceGroupName(),
		serverName,
		dbName,
		sql.Database{
			Location: to.StringPtr(internal.Location()),
		})
	if err != nil {
		return db, fmt.Errorf("cannot create sql database: %v", err)
	}

	err = future.WaitForCompletion(ctx, dbClient.Client)
	if err != nil {
		return db, fmt.Errorf("cannot get the sql database create or update future response: %v", err)
	}

	return future.Result(dbClient)
}

// DeleteDB deletes an existing database from a server
func DeleteDB(ctx context.Context, serverName, dbName string) (autorest.Response, error) {
	dbClient := getDbClient()
	return dbClient.Delete(
		ctx,
		internal.ResourceGroupName(),
		serverName,
		dbName,
	)
}

// Firewall rukes

func getFwRulesClient() sql.FirewallRulesClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	fwrClient := sql.NewFirewallRulesClient(internal.SubscriptionID())
	fwrClient.Authorizer = autorest.NewBearerAuthorizer(token)
	fwrClient.AddToUserAgent(internal.UserAgent())
	return fwrClient
}

// CreateFirewallRules creates new firewall rules for a given server
func CreateFirewallRules(ctx context.Context, serverName string) error {
	fwrClient := getFwRulesClient()

	_, err := fwrClient.CreateOrUpdate(
		ctx,
		internal.ResourceGroupName(),
		serverName,
		"unsafe open to world",
		sql.FirewallRule{
			FirewallRuleProperties: &sql.FirewallRuleProperties{
				StartIPAddress: to.StringPtr("0.0.0.0"),
				EndIPAddress:   to.StringPtr("255.255.255.255"),
			},
		},
	)
	if err != nil {
		return err
	}

	_, err = fwrClient.CreateOrUpdate(
		ctx,
		internal.ResourceGroupName(),
		serverName,
		"open to Azure internal",
		sql.FirewallRule{
			FirewallRuleProperties: &sql.FirewallRuleProperties{
				StartIPAddress: to.StringPtr("0.0.0.0"),
				EndIPAddress:   to.StringPtr("0.0.0.0"),
			},
		},
	)

	return err
}

// PrintInfo logs information on SQL user agent and ARM client
func PrintInfo() {
	log.Printf("user agent string: %s\n", sql.UserAgent())
	log.Printf("SQL ARM Client version: %s\n", sql.Version())
}
