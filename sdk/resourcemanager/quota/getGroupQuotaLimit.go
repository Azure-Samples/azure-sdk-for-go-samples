package main

import (
	"context"
	"fmt"
	"log"
)

func getGroupQuotaLimit(groupName string, provider string, resourceName string, region string) {
	ctx := context.Background()
	clientFactory := createClientFactory()
	filterString := fmt.Sprintf("location eq %s", region)
	res, err := clientFactory.NewGroupQuotaLimitsClient().Get(ctx, managementGroupId, groupName, provider, resourceName, filterString, nil)
	if err != nil {
		log.Fatalf("failed to finish the request: %v", err)
	}
	// You could use response here. We use blank identifier for just demo purposes.
	_ = res
	// If the HTTP response code is 200 as defined in example definition, your response structure would look as follows. Please pay attention that all the values in the output are fake values for just demo purposes.
	// res.GroupQuotaLimit = armquota.GroupQuotaLimit{
	// 	Name: to.Ptr("cores"),
	// 	Type: to.Ptr("Microsoft.Quota/groupQuotas/groupQuotaLimits"),
	// 	ID: to.Ptr("/providers/Microsoft.Management/managementGroups/E7EC67B3-7657-4966-BFFC-41EFD36BAA09/providers/Microsoft.Quota/groupQuotas/groupquota1/providers/Microsoft.Compute/locations/westus/groupQuotaLimits/cores"),
	// 	Properties: &armquota.GroupQuotaDetails{
	// 		Name: &armquota.GroupQuotaDetailsName{
	// 			LocalizedValue: to.Ptr("Total vCPUs Regional Cores"),
	// 			Value: to.Ptr("cores"),
	// 		},
	// 		AllocatedToSubscriptions: &armquota.AllocatedQuotaToSubscriptionList{
	// 			Value: []*armquota.AllocatedToSubscription{
	// 				{
	// 					QuotaAllocated: to.Ptr[int64](20),
	// 					SubscriptionID: to.Ptr("00000000-0000-0000-0000-000000000000"),
	// 			}},
	// 		},
	// 		AvailableLimit: to.Ptr[int64](80),
	// 		Limit: to.Ptr[int64](100),
	// 		Region: to.Ptr("westus"),
	// 		Unit: to.Ptr("count"),
	// 	},
	// }
}
