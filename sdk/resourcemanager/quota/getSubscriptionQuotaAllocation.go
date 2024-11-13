package main

import (
	"context"
	"fmt"
	"log"
)

// Generated from example definition: https://github.com/Azure/azure-rest-api-specs/blob/106483d9f698ac3b6c0d481ab0c5fab14152e21f/specification/quota/resource-manager/Microsoft.Quota/preview/2023-06-01-preview/examples/SubscriptionQuotaAllocation/SubscriptionQuotaAllocation_Get-Compute.json
func getSubscriptionQuotaAllocation(groupName string, resourceName string, region string) {
	ctx := context.Background()
	clientFactory := createClientFactory()
	filterString := fmt.Sprintf("location eq %s", region)
	res, err := clientFactory.NewGroupQuotaSubscriptionAllocationClient().Get(ctx, managementGroupId, groupName, resourceName, filterString, nil)
	if err != nil {
		log.Fatalf("failed to finish the request: %v", err)
	}
	// You could use response here. We use blank identifier for just demo purposes.
	_ = res
	// If the HTTP response code is 200 as defined in example definition, your response structure would look as follows. Please pay attention that all the values in the output are fake values for just demo purposes.
	// res.SubscriptionQuotaAllocations = armquota.SubscriptionQuotaAllocations{
	// 	Name: to.Ptr("standardav2family"),
	// 	Type: to.Ptr("Microsoft.Quota/groupQuotas/quotaAllocations"),
	// 	ID: to.Ptr("/providers/Microsoft.Management/managementGroups/E7EC67B3-7657-4966-BFFC-41EFD36BAA09/subscriptions/00000000-0000-0000-0000-000000000000/providers/Microsoft.Quota/groupQuotas/groupquota1/providers/Microsoft.Compute/locations/westus/quotaAllocations/standardav2family"),
	// 	Properties: &armquota.SubscriptionQuotaDetails{
	// 		Name: &armquota.SubscriptionQuotaDetailsName{
	// 			LocalizedValue: to.Ptr("standard Av2 Family vCPUs"),
	// 			Value: to.Ptr("standardav2family"),
	// 		},
	// 		Limit: to.Ptr[int64](100),
	// 		Region: to.Ptr("westus"),
	// 		ShareableQuota: to.Ptr[int64](25),
	// 	},
	// }
}