package main

import (
	"context"
	"log"
)

func deleteGroupQuota(groupName string) {
	ctx := context.Background()
	clientFactory := createClientFactory()

	poller, err := clientFactory.NewGroupQuotasClient().BeginDelete(ctx, managementGroupId, groupName, nil)
	if err != nil {
		log.Fatalf("failed to finish the request: %v", err)
	}
	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		log.Fatalf("failed to pull the result: %v", err)
	}
}
