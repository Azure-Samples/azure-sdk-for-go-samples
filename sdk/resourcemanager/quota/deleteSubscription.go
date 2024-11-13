package main

import (
	"context"
	"log"
)

func deleteSubscription(groupName string) {
	ctx := context.Background()
	clientFactory := createClientFactory()

	poller, err := clientFactory.NewGroupQuotaSubscriptionsClient().BeginDelete(ctx, managementGroupId, groupName, nil)
	if err != nil {
		log.Fatalf("failed to finish the request: %v", err)
	}
	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		log.Fatalf("failed to pull the result: %v", err)
	}
}
