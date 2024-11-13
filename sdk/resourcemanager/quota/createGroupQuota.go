package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/quota/armquota"
)

func createGroupQuota(groupName string, managementGroupId string) {
	clientFactory := createClientFactory()
	ctx := context.Background()

	groupQuotaBody := &armquota.GroupQuotasClientBeginCreateOrUpdateOptions{
		GroupQuotaPutRequestBody: &armquota.GroupQuotasEntity{
			Properties: &armquota.GroupQuotasEntityBase{
				AdditionalAttributes: &armquota.AdditionalAttributes{
					GroupID: &armquota.GroupingID{
						GroupingIDType: to.Ptr(armquota.GroupingIDTypeBillingID),
						Value:          to.Ptr("E7EC67B3-7657-4966-BFFC-41EFD36BAA09"),
					},
				},
				DisplayName: to.Ptr(groupName),
			},
		},
	}

	poller, err := clientFactory.NewGroupQuotasClient().BeginCreateOrUpdate(ctx, managementGroupId, groupName, groupQuotaBody)

	if err != nil {
		log.Fatalf("failed to create GroupQuota: %v", err)
	}

	fmt.Println("waiting for GroupQuota creation to complete")

	res, err := poller.PollUntilDone(ctx, nil)

	if err != nil {
		log.Fatalf("failed to poll the result of the request: %v", err)
	}

	_ = res
	//return res
}
