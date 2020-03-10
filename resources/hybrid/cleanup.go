// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package resources

import (
	"context"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
)

// Cleanup deletes the resource group created for the sample
func Cleanup(ctx context.Context) {
	if config.KeepResources() {
		log.Println("Hybrid resources cleanup: keeping resources")
		return
	}
	log.Println("Hybrid resources cleanup: deleting resources")
	_, _ = DeleteGroup(ctx)
}
