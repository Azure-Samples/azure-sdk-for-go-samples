package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/quota/armquota"
)

func createGroupQuotaLimitRequest(groupName string, provider string, resourceName string, region string, limitVal int64) {
	ctx := context.Background()
	clientFactory := createClientFactory()

	groupQuotaLimitRequestBody := &armquota.GroupQuotaLimitsRequestClientBeginCreateOrUpdateOptions{
		GroupQuotaRequest: &armquota.SubmittedResourceRequestStatus{
			Properties: &armquota.SubmittedResourceRequestStatusProperties{
				RequestedResource: &armquota.GroupQuotaRequestBase{
					Properties: &armquota.GroupQuotaRequestBaseProperties{
						Limit:  &limitVal,
						Region: to.Ptr(region),
					},
				},
			},
		},
	}
	groupQuotaLimitsRequestClient := clientFactory.NewGroupQuotaLimitsRequestClient()
	poller, err := groupQuotaLimitsRequestClient.BeginCreateOrUpdate(ctx, managementGroupId, groupName, provider, resourceName, groupQuotaLimitRequestBody)
	if err != nil {
		log.Fatalf("failed to finish the request: %v", err)
	}

	// Get the HTTP response from the quota limit request
	res, err := poller.Poll(ctx)
	if err != nil {
		panic(err)
	}
	opStatus := res.Header["Location"][0]
	opStatusURI, err := url.Parse(opStatus)
	if err != nil {
		panic(err)
	}

	// Get the request ID and the Request URL in case the request doesn't reach a terminal state in 2 minutes
	opStatusURISegments := strings.Split(opStatusURI.Path, "/")
	id := opStatusURISegments[len(opStatusURISegments)-1]

	// Poll for 2 minutes until the group quota limit request has completed
	start := time.Now()
	duration := 2 * time.Minute
	provisioningState := armquota.RequestStateInProgress
	for time.Since(start) < duration {
		group_limit_request_res, err := groupQuotaLimitsRequestClient.Get(ctx, managementGroupId, groupName, id, nil)
		if err != nil {
			panic(err)
		}

		provisioningState = *group_limit_request_res.Properties.ProvisioningState
		if provisioningState == armquota.RequestStateSucceeded {
			groupQuotaLimitClient := clientFactory.NewGroupQuotaLimitsClient()
			filterString := fmt.Sprintf("location eq %s", region)
			limit_res, err := groupQuotaLimitClient.Get(ctx, managementGroupId, groupName, provider, resourceName, filterString, nil)
			if err != nil {
				log.Fatalf("failed to finish the request: %v", err)
			}
			// You could use response here. We use blank identifier for just demo purposes.
			_ = limit_res
			break
		} else if provisioningState == armquota.RequestStateFailed {
			fmt.Println("Group Quota Limit Request Failed")
			log.Fatalf("Request failed")
		}
		fmt.Println("Polling...")
		time.Sleep(30 * time.Second)
	}

	if provisioningState == armquota.RequestStateInProgress || provisioningState == "Escalated" {
		fmt.Println("Did not reach terminal state within 2 minutes. Please perform a get on this URL: %s", opStatus)
	}

	if provisioningState == "Escalated" {
		fmt.Println("Request was escalated please contact your capacity manager. Please perform a get on this URL: %s", opStatus)
	}

	// You could use response here. We use blank identifier for just demo purposes.
	//_ = limit_res
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
