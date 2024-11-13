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

// Generated from example definition: https://github.com/Azure/azure-rest-api-specs/blob/106483d9f698ac3b6c0d481ab0c5fab14152e21f/specification/quota/resource-manager/Microsoft.Quota/preview/2023-06-01-preview/examples/SubscriptionQuotaAllocationRequests/PutSubscriptionQuotaAllocationRequest-Compute.json
func createSubscriptionAllocationRequest(groupName string, provider string, resourceName string, region string, allocVal int64) {
	ctx := context.Background()
	clientFactory := createClientFactory()
	subscriptionQuotaAllocationRequestClient := clientFactory.NewGroupQuotaSubscriptionAllocationRequestClient()
	poller, err := subscriptionQuotaAllocationRequestClient.BeginCreateOrUpdate(ctx, managementGroupId, groupName, provider, resourceName, armquota.AllocationRequestStatus{
		Properties: &armquota.AllocationRequestStatusProperties{
			RequestedResource: &armquota.AllocationRequestBase{
				Properties: &armquota.AllocationRequestBaseProperties{
					Limit:  to.Ptr[int64](allocVal),
					Region: to.Ptr(region),
				},
			},
		},
	}, nil)
	if err != nil {
		log.Fatalf("failed to finish the request: %v", err)
	}
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
		allocation_res, err := subscriptionQuotaAllocationRequestClient.Get(ctx, managementGroupId, groupName, id, nil)
		if err != nil {
			panic(err)
		}

		provisioningState = *allocation_res.Properties.ProvisioningState
		if provisioningState == armquota.RequestStateSucceeded {
			subscriptionQuotaAllocationClient := clientFactory.NewGroupQuotaSubscriptionAllocationClient()
			filterString := fmt.Sprintf("location eq %s", region)
			allocation_res, err := subscriptionQuotaAllocationClient.Get(ctx, managementGroupId, groupName, resourceName, filterString, nil)
			if err != nil {
				log.Fatalf("failed to finish the request: %v", err)
			}
			// You could use response here. We use blank identifier for just demo purposes.
			_ = allocation_res
			break
		} else if provisioningState == armquota.RequestStateFailed {
			fmt.Println("Group Quota Subscription Allocation Failed")
			log.Fatalf("Request failed")
		}
		fmt.Println("Polling...")
		time.Sleep(30 * time.Second)
	}

	if provisioningState == armquota.RequestStateInProgress || provisioningState == armquota.RequestStateAccepted {
		fmt.Println("Did not reach terminal state within 2 minutes. Please perform a get on this URL: %s", opStatus)
	}

	if provisioningState == "Escalated" {
		fmt.Println("Request was escalated please contact your capacity manager. Please perform a get on this URL: %s", opStatus)
	}

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
