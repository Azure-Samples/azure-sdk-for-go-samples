package main

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/armcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/to"
	"log"
	"net/http"
	"os"
	"time"
)

var subscriptionID string
var location = "westus"
var resourceGroupName = "sample-resource-group"
var securityGroupName = "sample-network-security-group"

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}

	conn := armcore.NewDefaultConnection(cred, &armcore.ConnectionOptions{
		Logging: azcore.LogOptions{
			IncludeBody: true,
		},
	})
	ctx := context.Background()

	resourceGroup,err := createResourceGroup(ctx,conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resource group:",*resourceGroup.ID)

	networkSecurityGroup,err := createNetworkSecurityGroup(ctx,conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("network security group:",*networkSecurityGroup.ID)

	sshRule,err := createSSHRule(ctx,conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("SSH:",*sshRule.ID)

	httpRule,err := createHTTPRule(ctx,conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("HTTP:",*httpRule.ID)

	sqlRule,err := createSQLRule(ctx,conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("SQL:",*sqlRule.ID)

	denyOutRule,err := createDenyOutRule(ctx,conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Deny Out:",*denyOutRule.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_,err := cleanup(ctx,conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createNetworkSecurityGroup(ctx context.Context,conn *armcore.Connection) (*armnetwork.NetworkSecurityGroup,error) {
	networkSecurityGroupClient := armnetwork.NewNetworkSecurityGroupsClient(conn,subscriptionID)

	pollerResp,err := networkSecurityGroupClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		securityGroupName,
		armnetwork.NetworkSecurityGroup{
			Resource: armnetwork.Resource{
				Location: to.StringPtr(location),
			},
			Properties: &armnetwork.NetworkSecurityGroupPropertiesFormat{
				SecurityRules: []*armnetwork.SecurityRule{
					{
						Name: to.StringPtr("allow_ssh"),
						Properties: &armnetwork.SecurityRulePropertiesFormat{
							Protocol:                 armnetwork.SecurityRuleProtocolTCP.ToPtr(),
							SourceAddressPrefix:      to.StringPtr("0.0.0.0/0"),
							SourcePortRange:          to.StringPtr("1-65535"),
							DestinationAddressPrefix: to.StringPtr("0.0.0.0/0"),
							DestinationPortRange:     to.StringPtr("22"),
							Access:                   armnetwork.SecurityRuleAccessAllow.ToPtr(),
							Direction:                armnetwork.SecurityRuleDirectionInbound.ToPtr(),
							Priority:                 to.Int32Ptr(100),
						},
					},
					{
						Name: to.StringPtr("allow_https"),
						Properties: &armnetwork.SecurityRulePropertiesFormat{
							Protocol:                 armnetwork.SecurityRuleProtocolTCP.ToPtr(),
							SourceAddressPrefix:      to.StringPtr("0.0.0.0/0"),
							SourcePortRange:          to.StringPtr("1-65535"),
							DestinationAddressPrefix: to.StringPtr("0.0.0.0/0"),
							DestinationPortRange:     to.StringPtr("443"),
							Access:                   armnetwork.SecurityRuleAccessAllow.ToPtr(),
							Direction:                armnetwork.SecurityRuleDirectionInbound.ToPtr(),
							Priority:                 to.Int32Ptr(200),
						},
					},
				},
			},
		},
		nil)

	if err != nil {
		return nil,err
	}

	resp,err := pollerResp.PollUntilDone(ctx, 10 * time.Second)
	if err != nil {
		return nil,err
	}
	return resp.NetworkSecurityGroup,nil
}

func createSSHRule(ctx context.Context, conn *armcore.Connection) (*armnetwork.SecurityRule, error) {
	securityRules := armnetwork.NewSecurityRulesClient(conn,subscriptionID)

	pollerResp, err := securityRules.BeginCreateOrUpdate(ctx,
		resourceGroupName,
		securityGroupName,
		"ALLOW-SSH",
		armnetwork.SecurityRule{
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Access: armnetwork.SecurityRuleAccessAllow.ToPtr(),
				DestinationAddressPrefix: to.StringPtr("*"),
				DestinationPortRange:     to.StringPtr("22"),
				Direction:                armnetwork.SecurityRuleDirectionInbound.ToPtr(),
				Description:              to.StringPtr("Allow SSH"),
				Priority:                 to.Int32Ptr(103),
				Protocol:                 armnetwork.SecurityRuleProtocolTCP.ToPtr(),
				SourceAddressPrefix:      to.StringPtr("*"),
				SourcePortRange:          to.StringPtr("*"),
			},
		},
		nil)

	if err != nil {
		return nil, fmt.Errorf("cannot create SSH security rule: %v", err)
	}

	resp,err := pollerResp.PollUntilDone(ctx, 10 * time.Second)
	if err != nil {
		return nil, fmt.Errorf("cannot get security rule create or update future response: %v", err)
	}

	return resp.SecurityRule,nil
}

func createHTTPRule(ctx context.Context, conn *armcore.Connection) (*armnetwork.SecurityRule, error) {
	securityRules := armnetwork.NewSecurityRulesClient(conn,subscriptionID)

	pollerResp, err := securityRules.BeginCreateOrUpdate(ctx,
		resourceGroupName,
		securityGroupName,
		"ALLOW-HTTP",
		armnetwork.SecurityRule{
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Access: armnetwork.SecurityRuleAccessAllow.ToPtr(),
				DestinationAddressPrefix: to.StringPtr("*"),
				DestinationPortRange:     to.StringPtr("80"),
				Direction:                armnetwork.SecurityRuleDirectionInbound.ToPtr(),
				Description:              to.StringPtr("Allow HTTP"),
				Priority:                 to.Int32Ptr(101),
				Protocol:                 armnetwork.SecurityRuleProtocolTCP.ToPtr(),
				SourceAddressPrefix:      to.StringPtr("*"),
				SourcePortRange:          to.StringPtr("*"),
			},
		},
		nil)

	if err != nil {
		return nil, fmt.Errorf("cannot create HTTP security rule: %v", err)
	}

	resp,err := pollerResp.PollUntilDone(ctx, 10 * time.Second)
	if err != nil {
		return nil, fmt.Errorf("cannot get security rule create or update future response: %v", err)
	}

	return resp.SecurityRule,nil
}

func createSQLRule(ctx context.Context, conn *armcore.Connection) (*armnetwork.SecurityRule, error) {
	securityRules := armnetwork.NewSecurityRulesClient(conn,subscriptionID)

	pollerResp, err := securityRules.BeginCreateOrUpdate(ctx,
		resourceGroupName,
		securityGroupName,
		"ALLOW-SQL",
		armnetwork.SecurityRule{
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Access:                   armnetwork.SecurityRuleAccessAllow.ToPtr(),
				DestinationAddressPrefix: to.StringPtr("*"),
				DestinationPortRange:     to.StringPtr("1433"),
				Direction:                armnetwork.SecurityRuleDirectionInbound.ToPtr(),
				Description:              to.StringPtr("Allow SQL"),
				Priority:                 to.Int32Ptr(102),
				Protocol:                 armnetwork.SecurityRuleProtocolTCP.ToPtr(),
				SourceAddressPrefix:      to.StringPtr("*"),
				SourcePortRange:          to.StringPtr("*"),
			},
		},
		nil)

	if err != nil {
		return nil, fmt.Errorf("cannot create SQL security rule: %v", err)
	}

	resp,err := pollerResp.PollUntilDone(ctx, 10 * time.Second)
	if err != nil {
		return nil, fmt.Errorf("cannot get security rule create or update future response: %v", err)
	}

	return resp.SecurityRule,nil
}

func createDenyOutRule(ctx context.Context, conn *armcore.Connection) (*armnetwork.SecurityRule, error) {
	securityRules := armnetwork.NewSecurityRulesClient(conn,subscriptionID)

	pollerResp, err := securityRules.BeginCreateOrUpdate(ctx,
		resourceGroupName,
		securityGroupName,
		"DENY-OUT",
		armnetwork.SecurityRule{
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Access: armnetwork.SecurityRuleAccessDeny.ToPtr(),
				DestinationAddressPrefix: to.StringPtr("*"),
				DestinationPortRange:     to.StringPtr("*"),
				Direction:                armnetwork.SecurityRuleDirectionOutbound.ToPtr(),
				Description:              to.StringPtr("Deny outbound traffic"),
				Priority:                 to.Int32Ptr(100),
				Protocol:                 armnetwork.SecurityRuleProtocolAsterisk.ToPtr(),
				SourceAddressPrefix:      to.StringPtr("*"),
				SourcePortRange:          to.StringPtr("*"),
			},
		},
		nil)

	if err != nil {
		return nil, fmt.Errorf("cannot create deny out security rule: %v", err)
	}

	resp,err := pollerResp.PollUntilDone(ctx, 10 * time.Second)
	if err != nil {
		return nil, fmt.Errorf("cannot get security rule create or update future response: %v", err)
	}

	return resp.SecurityRule,nil
}

func createResourceGroup(ctx context.Context,conn *armcore.Connection) (*armresources.ResourceGroup,error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn,subscriptionID)

	resourceGroupResp,err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.StringPtr(location),
		},
		nil)
	if err != nil {
		return nil,err
	}
	return resourceGroupResp.ResourceGroup,nil
}

func cleanup(ctx context.Context, conn *armcore.Connection) (*http.Response,error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn,subscriptionID)
	log.Println("cleanup...")

	pollerResp,err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return nil,err
	}

	resp,err := pollerResp.PollUntilDone(ctx, 10 * time.Second)
	if err != nil {
		return nil, err
	}
	return resp,nil
}