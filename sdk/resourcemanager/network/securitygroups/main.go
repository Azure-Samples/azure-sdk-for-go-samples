// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resources-group"
	securityGroupName = "sample-network-security-group"
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
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	networkSecurityGroup, err := createNetworkSecurityGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("network security group:", *networkSecurityGroup.ID)

	sshRule, err := createSSHRule(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("SSH:", *sshRule.ID)

	httpRule, err := createHTTPRule(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("HTTP:", *httpRule.ID)

	sqlRule, err := createSQLRule(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("SQL:", *sqlRule.ID)

	denyOutRule, err := createDenyOutRule(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Deny Out:", *denyOutRule.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createNetworkSecurityGroup(ctx context.Context, cred azcore.TokenCredential) (*armnetwork.SecurityGroup, error) {
	networkSecurityGroupClient, err := armnetwork.NewSecurityGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := networkSecurityGroupClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		securityGroupName,
		armnetwork.SecurityGroup{
			Location: to.Ptr(location),
			Properties: &armnetwork.SecurityGroupPropertiesFormat{
				SecurityRules: []*armnetwork.SecurityRule{
					{
						Name: to.Ptr("allow_ssh"),
						Properties: &armnetwork.SecurityRulePropertiesFormat{
							Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
							SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
							SourcePortRange:          to.Ptr("1-65535"),
							DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
							DestinationPortRange:     to.Ptr("22"),
							Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
							Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
							Priority:                 to.Ptr[int32](100),
						},
					},
					{
						Name: to.Ptr("allow_https"),
						Properties: &armnetwork.SecurityRulePropertiesFormat{
							Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
							SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
							SourcePortRange:          to.Ptr("1-65535"),
							DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
							DestinationPortRange:     to.Ptr("443"),
							Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
							Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
							Priority:                 to.Ptr[int32](200),
						},
					},
				},
			},
		},
		nil)

	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.SecurityGroup, nil
}

func createSSHRule(ctx context.Context, cred azcore.TokenCredential) (*armnetwork.SecurityRule, error) {
	securityRules, err := armnetwork.NewSecurityRulesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := securityRules.BeginCreateOrUpdate(ctx,
		resourceGroupName,
		securityGroupName,
		"ALLOW-SSH",
		armnetwork.SecurityRule{
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
				DestinationAddressPrefix: to.Ptr("*"),
				DestinationPortRange:     to.Ptr("22"),
				Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
				Description:              to.Ptr("Allow SSH"),
				Priority:                 to.Ptr[int32](103),
				Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
				SourceAddressPrefix:      to.Ptr("*"),
				SourcePortRange:          to.Ptr("*"),
			},
		},
		nil)

	if err != nil {
		return nil, fmt.Errorf("cannot create SSH security rule: %v", err)
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot get security rule create or update future response: %v", err)
	}

	return &resp.SecurityRule, nil
}

func createHTTPRule(ctx context.Context, cred azcore.TokenCredential) (*armnetwork.SecurityRule, error) {
	securityRules, err := armnetwork.NewSecurityRulesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := securityRules.BeginCreateOrUpdate(ctx,
		resourceGroupName,
		securityGroupName,
		"ALLOW-HTTP",
		armnetwork.SecurityRule{
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
				DestinationAddressPrefix: to.Ptr("*"),
				DestinationPortRange:     to.Ptr("80"),
				Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
				Description:              to.Ptr("Allow HTTP"),
				Priority:                 to.Ptr[int32](101),
				Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
				SourceAddressPrefix:      to.Ptr("*"),
				SourcePortRange:          to.Ptr("*"),
			},
		},
		nil)

	if err != nil {
		return nil, fmt.Errorf("cannot create HTTP security rule: %v", err)
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot get security rule create or update future response: %v", err)
	}

	return &resp.SecurityRule, nil
}

func createSQLRule(ctx context.Context, cred azcore.TokenCredential) (*armnetwork.SecurityRule, error) {
	securityRules, err := armnetwork.NewSecurityRulesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := securityRules.BeginCreateOrUpdate(ctx,
		resourceGroupName,
		securityGroupName,
		"ALLOW-SQL",
		armnetwork.SecurityRule{
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
				DestinationAddressPrefix: to.Ptr("*"),
				DestinationPortRange:     to.Ptr("1433"),
				Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
				Description:              to.Ptr("Allow SQL"),
				Priority:                 to.Ptr[int32](102),
				Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
				SourceAddressPrefix:      to.Ptr("*"),
				SourcePortRange:          to.Ptr("*"),
			},
		},
		nil)

	if err != nil {
		return nil, fmt.Errorf("cannot create SQL security rule: %v", err)
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot get security rule create or update future response: %v", err)
	}

	return &resp.SecurityRule, nil
}

func createDenyOutRule(ctx context.Context, cred azcore.TokenCredential) (*armnetwork.SecurityRule, error) {
	securityRules, err := armnetwork.NewSecurityRulesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := securityRules.BeginCreateOrUpdate(ctx,
		resourceGroupName,
		securityGroupName,
		"DENY-OUT",
		armnetwork.SecurityRule{
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Access:                   to.Ptr(armnetwork.SecurityRuleAccessDeny),
				DestinationAddressPrefix: to.Ptr("*"),
				DestinationPortRange:     to.Ptr("*"),
				Direction:                to.Ptr(armnetwork.SecurityRuleDirectionOutbound),
				Description:              to.Ptr("Deny outbound traffic"),
				Priority:                 to.Ptr[int32](100),
				Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolAsterisk),
				SourceAddressPrefix:      to.Ptr("*"),
				SourcePortRange:          to.Ptr("*"),
			},
		},
		nil)

	if err != nil {
		return nil, fmt.Errorf("cannot create deny out security rule: %v", err)
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot get security rule create or update future response: %v", err)
	}

	return &resp.SecurityRule, nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.Ptr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, cred azcore.TokenCredential) error {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return err
	}
	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
