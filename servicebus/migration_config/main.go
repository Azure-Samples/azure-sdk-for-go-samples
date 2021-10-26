package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
)

var (
	subscriptionID        string
	location              = "westus"
	resourceGroupName     = "sample-resource-group"
	namespaceName         = "sample-sb-namespace"
	authorizationRuleName = "sample-sb-authorization-rule"
	postMigrationName     = "sample-sb-post-migration"
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}

	conn := arm.NewDefaultConnection(cred, &arm.ConnectionOptions{
		Logging: policy.LogOptions{
			IncludeBody: true,
		},
	})
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	namespace, err := createNamespace(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus namespace:", *namespace.ID)

	namespacePrimary, err := createNamespacePrimary(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus namespace primary:", *namespacePrimary.ID)

	namespaceAuthorizationRule, err := createNamespaceAuthorizationRule(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus namespace authorization rule:", *namespaceAuthorizationRule.ID)

	migrationConfig, err := createMigrationConfig(ctx, conn, *namespacePrimary.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus migration config:", *migrationConfig.ID)

	complete, err := completeMigrationConfig(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus migration complete:", complete)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createNamespace(ctx context.Context, conn *arm.Connection) (np armservicebus.NamespacesCreateOrUpdateResponse, err error) {
	namespacesClient := armservicebus.NewNamespacesClient(conn, subscriptionID)

	pollerResp, err := namespacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		namespaceName,
		armservicebus.SBNamespace{
			TrackedResource: armservicebus.TrackedResource{
				Location: to.StringPtr(location),
			},
			SKU: &armservicebus.SBSKU{
				Name: armservicebus.SKUNamePremium.ToPtr(),
				Tier: armservicebus.SKUTierPremium.ToPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return np, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return np, err
	}
	return resp, nil
}

func createNamespacePrimary(ctx context.Context, conn *arm.Connection) (np armservicebus.NamespacesCreateOrUpdateResponse, err error) {
	namespacesClient := armservicebus.NewNamespacesClient(conn, subscriptionID)

	pollerResp, err := namespacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		namespaceName,
		armservicebus.SBNamespace{
			TrackedResource: armservicebus.TrackedResource{
				Location: to.StringPtr("eastus"),
			},
			SKU: &armservicebus.SBSKU{
				Name: armservicebus.SKUNamePremium.ToPtr(),
				Tier: armservicebus.SKUTierPremium.ToPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return np, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return np, err
	}
	return resp, nil
}

func createNamespaceAuthorizationRule(ctx context.Context, conn *arm.Connection) (ar armservicebus.NamespacesCreateOrUpdateAuthorizationRuleResponse, err error) {
	namespacesClient := armservicebus.NewNamespacesClient(conn, subscriptionID)

	resp, err := namespacesClient.CreateOrUpdateAuthorizationRule(
		ctx,
		resourceGroupName,
		namespaceName,
		authorizationRuleName,
		armservicebus.SBAuthorizationRule{
			Properties: &armservicebus.SBAuthorizationRuleProperties{
				Rights: []*armservicebus.AccessRights{
					armservicebus.AccessRightsListen.ToPtr(),
					armservicebus.AccessRightsSend.ToPtr(),
				},
			},
		},
		nil,
	)
	if err != nil {
		return ar, err
	}

	return resp, nil
}

func createMigrationConfig(ctx context.Context, conn *arm.Connection, secondNamespaceID string) (*armservicebus.MigrationConfigProperties, error) {
	migrationConfigsClient := armservicebus.NewMigrationConfigsClient(conn, subscriptionID)

	pollerResp, err := migrationConfigsClient.BeginCreateAndStartMigration(
		ctx,
		resourceGroupName,
		namespaceName,
		armservicebus.MigrationConfigurationNameDefault,
		armservicebus.MigrationConfigProperties{
			Properties: &armservicebus.MigrationConfigPropertiesProperties{
				TargetNamespace:   to.StringPtr(secondNamespaceID),
				PostMigrationName: to.StringPtr(postMigrationName),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.MigrationConfigProperties, nil
}

func completeMigrationConfig(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	migrationConfigsClient := armservicebus.NewMigrationConfigsClient(conn, subscriptionID)

	resp, err := migrationConfigsClient.CompleteMigration(
		ctx,
		resourceGroupName,
		namespaceName,
		armservicebus.MigrationConfigurationNameDefault,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}

func createResourceGroup(ctx context.Context, conn *arm.Connection) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.StringPtr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}
