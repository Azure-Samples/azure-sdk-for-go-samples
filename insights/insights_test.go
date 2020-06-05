// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package insights

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/web"
	"github.com/marstr/randname"
)

var (
	siteName = randname.GenerateWithPrefix("web-site-go-samples", 10)
)

// TestMain sets up the environment and initiates tests.
func TestMain(m *testing.M) {
	var err error
	err = config.ParseEnvironment()
	if err != nil {
		log.Fatalf("failed to parse env: %v\n", err)
	}
	err = config.AddFlags()
	if err != nil {
		log.Fatalf("failed to parse env: %v\n", err)
	}
	flag.Parse()

	os.Exit(m.Run())
}

// ExampleGetMetricsForWebsite creates a website then uses the insights package
// to retrieve the queryable metric names and values.
func TestGetMetricsForWebsite(t *testing.T) {
	var groupName = config.GenerateGroupName("GetMetricsForWebsite")
	config.SetGroupName(groupName)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	_, err := resources.CreateGroup(ctx, config.GroupName())
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created resource group")
	defer resources.Cleanup(ctx)

	webSite, err := web.CreateWebApp(ctx, siteName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("created web site")

	// retrieve the list of metric definitions.  each resource type
	// will have its own set of queryable metrics.
	metrics, err := ListMetricDefinitions(*webSite.ID)
	if err != nil {
		util.LogAndPanic(err)
	}

	util.PrintAndLog("available metrics:")
	util.PrintAndLog(strings.Join(metrics, "\n"))

	// here, CpuTime and Requests are the non-localized metric names
	metricData, err := GetMetricsData(ctx, *webSite.ID, []string{"CpuTime", "Requests"})
	if err != nil {
		util.LogAndPanic(err)
	}

	util.PrintAndLog("metric data:")
	for _, md := range metricData {
		util.PrintAndLog(md)
	}
}
