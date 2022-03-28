// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package resources

import (
	"context"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/config"
)

// Cleanup deletes the resource group created for the sample
func Cleanup(ctx context.Context) {
	if config.KeepResources() {
		log.Println("keeping resources")
		return
	}
	log.Println("deleting resources")
	_, _ = DeleteGroup(ctx, config.GroupName())
}
