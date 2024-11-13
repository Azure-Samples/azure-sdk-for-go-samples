package main

import (
	"context"
	"log"
)

func addSubscription(groupName string) {
	ctx := context.Background()
	clientFactory := createClientFactory()

	poller, err := clientFactory.NewGroupQuotaSubscriptionsClient().BeginCreateOrUpdate(ctx, managementGroupId, groupName, nil)
	if err != nil {
		log.Fatalf("failed to finish the request: %v", err)
	}
	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		log.Fatalf("failed to pull the result: %v", err)
	}
	// You could use response here. We use blank identifier for just demo purposes.
	_ = res
	// If the HTTP response code is 200 as defined in example definition, your response structure would look as follows. Please pay attention that all the values in the output are fake values for just demo purposes.
	// res.GroupQuotaSubscriptionID = armquota.GroupQuotaSubscriptionID{
	// 	Name: to.Ptr("00000000-0000-0000-0000-000000000000"),
	// 	Type: to.Ptr("Microsoft.Quota/groupQuotas/subscriptions"),
	// 	ID: to.Ptr("/providers/Microsoft.Management/managementGroups/E7EC67B3-7657-4966-BFFC-41EFD36BAA09/providers/Microsoft.Quota/groupQuotas/groupquota1/subscriptions/00000000-0000-0000-0000-000000000000"),
	// 	Properties: &armquota.GroupQuotaSubscriptionIDProperties{
	// 		ProvisioningState: to.Ptr(armquota.RequestStateSucceeded),
	// 		SubscriptionID: to.Ptr("00000000-0000-0000-0000-000000000000"),
	// 	},
	// }
}
