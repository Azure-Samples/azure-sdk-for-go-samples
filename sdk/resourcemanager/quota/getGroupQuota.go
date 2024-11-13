package main

import (
	"context"
	"log"
)

func getGroupQuota(groupName string) {
	clientFactory := createClientFactory()
	ctx := context.Background()
	res, err := clientFactory.NewGroupQuotasClient().Get(ctx, managementGroupId, groupName, nil)
	if err != nil {
		log.Fatalf("failed to finish the request: %v", err)
	}
	// You could use response here. We use blank identifier for just demo purposes.
	_ = res
	// If the HTTP response code is 200 as defined in example definition, your response structure would look as follows. Please pay attention that all the values in the output are fake values for just demo purposes.
	// res.GroupQuotasEntity = armquota.GroupQuotasEntity{
	// 	Name: to.Ptr("groupquota1"),
	// 	Type: to.Ptr("Microsoft.Quota/groupQuotas"),
	// 	ID: to.Ptr("/providers/Microsoft.Management/managementGroups/E7EC67B3-7657-4966-BFFC-41EFD36BAA09/providers/Microsoft.Quota/groupQuotas/groupquota1"),
	// 	Properties: &armquota.GroupQuotasEntityBase{
	// 		AdditionalAttributes: &armquota.AdditionalAttributes{
	// 			Environment: to.Ptr(armquota.EnvironmentTypeProduction),
	// 			GroupID: &armquota.GroupingID{
	// 				GroupingIDType: to.Ptr(armquota.GroupingIDTypeServiceTreeID),
	// 				Value: to.Ptr("yourServiceTreeIdHere"),
	// 			},
	// 		},
	// 		DisplayName: to.Ptr("GroupQuota1"),
	// 		ProvisioningState: to.Ptr(armquota.RequestStateSucceeded),
	// 	},
	// }
}
