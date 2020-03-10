// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package network

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-11-01/network"
	"github.com/Azure/go-autorest/autorest/to"
)

func getNsgClient() network.SecurityGroupsClient {
	nsgClient := network.NewSecurityGroupsClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	nsgClient.Authorizer = a
	nsgClient.AddToUserAgent(config.UserAgent())
	return nsgClient
}

// CreateNetworkSecurityGroup creates a new network security group with rules set for allowing SSH and HTTPS use
func CreateNetworkSecurityGroup(ctx context.Context, nsgName string) (nsg network.SecurityGroup, err error) {
	nsgClient := getNsgClient()
	future, err := nsgClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		nsgName,
		network.SecurityGroup{
			Location: to.StringPtr(config.Location()),
			SecurityGroupPropertiesFormat: &network.SecurityGroupPropertiesFormat{
				SecurityRules: &[]network.SecurityRule{
					{
						Name: to.StringPtr("allow_ssh"),
						SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
							Protocol:                 network.SecurityRuleProtocolTCP,
							SourceAddressPrefix:      to.StringPtr("0.0.0.0/0"),
							SourcePortRange:          to.StringPtr("1-65535"),
							DestinationAddressPrefix: to.StringPtr("0.0.0.0/0"),
							DestinationPortRange:     to.StringPtr("22"),
							Access:                   network.SecurityRuleAccessAllow,
							Direction:                network.SecurityRuleDirectionInbound,
							Priority:                 to.Int32Ptr(100),
						},
					},
					{
						Name: to.StringPtr("allow_https"),
						SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
							Protocol:                 network.SecurityRuleProtocolTCP,
							SourceAddressPrefix:      to.StringPtr("0.0.0.0/0"),
							SourcePortRange:          to.StringPtr("1-65535"),
							DestinationAddressPrefix: to.StringPtr("0.0.0.0/0"),
							DestinationPortRange:     to.StringPtr("443"),
							Access:                   network.SecurityRuleAccessAllow,
							Direction:                network.SecurityRuleDirectionInbound,
							Priority:                 to.Int32Ptr(200),
						},
					},
				},
			},
		},
	)

	if err != nil {
		return nsg, fmt.Errorf("cannot create nsg: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, nsgClient.Client)
	if err != nil {
		return nsg, fmt.Errorf("cannot get nsg create or update future response: %v", err)
	}

	return future.Result(nsgClient)
}

// CreateSimpleNetworkSecurityGroup creates a new network security group, without rules (rules can be set later)
func CreateSimpleNetworkSecurityGroup(ctx context.Context, nsgName string) (nsg network.SecurityGroup, err error) {
	nsgClient := getNsgClient()
	future, err := nsgClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		nsgName,
		network.SecurityGroup{
			Location: to.StringPtr(config.Location()),
		},
	)

	if err != nil {
		return nsg, fmt.Errorf("cannot create nsg: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, nsgClient.Client)
	if err != nil {
		return nsg, fmt.Errorf("cannot get nsg create or update future response: %v", err)
	}

	return future.Result(nsgClient)
}

// DeleteNetworkSecurityGroup deletes an existing network security group
func DeleteNetworkSecurityGroup(ctx context.Context, nsgName string) (result network.SecurityGroupsDeleteFuture, err error) {
	nsgClient := getNsgClient()
	return nsgClient.Delete(ctx, config.GroupName(), nsgName)
}

// GetNetworkSecurityGroup returns an existing network security group
func GetNetworkSecurityGroup(ctx context.Context, nsgName string) (network.SecurityGroup, error) {
	nsgClient := getNsgClient()
	return nsgClient.Get(ctx, config.GroupName(), nsgName, "")
}

// Network security group rules

func getSecurityRulesClient() network.SecurityRulesClient {
	rulesClient := network.NewSecurityRulesClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	rulesClient.Authorizer = a
	rulesClient.AddToUserAgent(config.UserAgent())
	return rulesClient
}

// CreateSSHRule creates an inbound network security rule that allows using port 22
func CreateSSHRule(ctx context.Context, nsgName string) (rule network.SecurityRule, err error) {
	rulesClient := getSecurityRulesClient()
	future, err := rulesClient.CreateOrUpdate(ctx,
		config.GroupName(),
		nsgName,
		"ALLOW-SSH",
		network.SecurityRule{
			SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
				Access: network.SecurityRuleAccessAllow,
				DestinationAddressPrefix: to.StringPtr("*"),
				DestinationPortRange:     to.StringPtr("22"),
				Direction:                network.SecurityRuleDirectionInbound,
				Description:              to.StringPtr("Allow SSH"),
				Priority:                 to.Int32Ptr(103),
				Protocol:                 network.SecurityRuleProtocolTCP,
				SourceAddressPrefix:      to.StringPtr("*"),
				SourcePortRange:          to.StringPtr("*"),
			},
		})
	if err != nil {
		return rule, fmt.Errorf("cannot create SSH security rule: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, rulesClient.Client)
	if err != nil {
		return rule, fmt.Errorf("cannot get security rule create or update future response: %v", err)
	}

	return future.Result(rulesClient)
}

// CreateHTTPRule creates an inbound network security rule that allows using port 80
func CreateHTTPRule(ctx context.Context, nsgName string) (rule network.SecurityRule, err error) {
	rulesClient := getSecurityRulesClient()
	future, err := rulesClient.CreateOrUpdate(ctx,
		config.GroupName(),
		nsgName,
		"ALLOW-HTTP",
		network.SecurityRule{
			SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
				Access: network.SecurityRuleAccessAllow,
				DestinationAddressPrefix: to.StringPtr("*"),
				DestinationPortRange:     to.StringPtr("80"),
				Direction:                network.SecurityRuleDirectionInbound,
				Description:              to.StringPtr("Allow HTTP"),
				Priority:                 to.Int32Ptr(101),
				Protocol:                 network.SecurityRuleProtocolTCP,
				SourceAddressPrefix:      to.StringPtr("*"),
				SourcePortRange:          to.StringPtr("*"),
			},
		})
	if err != nil {
		return rule, fmt.Errorf("cannot create HTTP security rule: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, rulesClient.Client)
	if err != nil {
		return rule, fmt.Errorf("cannot get security rule create or update future response: %v", err)
	}

	return future.Result(rulesClient)
}

// CreateSQLRule creates an inbound network security rule that allows using port 1433
func CreateSQLRule(ctx context.Context, nsgName, frontEndAddressPrefix string) (rule network.SecurityRule, err error) {
	rulesClient := getSecurityRulesClient()
	future, err := rulesClient.CreateOrUpdate(ctx,
		config.GroupName(),
		nsgName,
		"ALLOW-SQL",
		network.SecurityRule{
			SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
				Access: network.SecurityRuleAccessAllow,
				DestinationAddressPrefix: to.StringPtr("*"),
				DestinationPortRange:     to.StringPtr("1433"),
				Direction:                network.SecurityRuleDirectionInbound,
				Description:              to.StringPtr("Allow SQL"),
				Priority:                 to.Int32Ptr(102),
				Protocol:                 network.SecurityRuleProtocolTCP,
				SourceAddressPrefix:      &frontEndAddressPrefix,
				SourcePortRange:          to.StringPtr("*"),
			},
		})
	if err != nil {
		return rule, fmt.Errorf("cannot create SQL security rule: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, rulesClient.Client)
	if err != nil {
		return rule, fmt.Errorf("cannot get security rule create or update future response: %v", err)
	}

	return future.Result(rulesClient)
}

// CreateDenyOutRule creates an network security rule that denies outbound traffic
func CreateDenyOutRule(ctx context.Context, nsgName string) (rule network.SecurityRule, err error) {
	rulesClient := getSecurityRulesClient()
	future, err := rulesClient.CreateOrUpdate(ctx,
		config.GroupName(),
		nsgName,
		"DENY-OUT",
		network.SecurityRule{
			SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
				Access: network.SecurityRuleAccessDeny,
				DestinationAddressPrefix: to.StringPtr("*"),
				DestinationPortRange:     to.StringPtr("*"),
				Direction:                network.SecurityRuleDirectionOutbound,
				Description:              to.StringPtr("Deny outbound traffic"),
				Priority:                 to.Int32Ptr(100),
				Protocol:                 network.SecurityRuleProtocolAsterisk,
				SourceAddressPrefix:      to.StringPtr("*"),
				SourcePortRange:          to.StringPtr("*"),
			},
		})
	if err != nil {
		return rule, fmt.Errorf("cannot create deny out security rule: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, rulesClient.Client)
	if err != nil {
		return rule, fmt.Errorf("cannot get security rule create or update future response: %v", err)
	}

	return future.Result(rulesClient)
}
