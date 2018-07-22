// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package compute

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/marstr/randname"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/network"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/Azure/go-autorest/autorest/to"
)

func ExampleDisk() {
	const groupName = config.GroupName()
	// TODO: remove and use local `groupName` only
	config.SetGroupName(groupName)

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	// don't delete resources so dataplane tests can reuse them
	// defer resources.Cleanup(ctx)

	// Disks
	_, err = AttachDataDisk(ctx, vmName)
	if err != nil {
		util.PrintAndLog(err.Error())
	}
	util.PrintAndLog("attached data disks")

	_, err = DetachDataDisks(ctx, vmName)
	if err != nil {
		util.PrintAndLog(err.Error())
	}
	util.PrintAndLog("detached data disks")

	_, err = UpdateOSDiskSize(ctx, vmName)
	if err != nil {
		util.PrintAndLog(err.Error())
	}
	util.PrintAndLog("updated OS disk size")

	// Output:
	// attached data disks
	// detached data disks
	// updated OS disk size
}
