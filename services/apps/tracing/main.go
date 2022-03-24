package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/dns/mgmt/2018-03-01-preview/dns"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	"go.opencensus.io/exporter/zipkin"
	"go.opencensus.io/trace"
)

const (
	subID      = "7127e532-e730-40dd-acda-0ca1105c1e55"
	rg         = "goTest"             // specify any preffered  value
	zoneName   = "azuregotestsdk.com" // specify any preffered value
	iterations = 1
	rrCount    = 2
	deleteRR   = true
	getRR      = true
	deleteZone = true
)

func main() {
	// enables Azure SDK tracing. This can be done also by setting the AZURE_SDK_TRACING_ENABLED value
	// If you are running an AI Local Forwarder you can use instead `tracing.EnableWithAIForwarding()`
	initZipkin()
	executeAzureAPICalls()
}

func executeAzureAPICalls() {
	zonesClient := dns.NewZonesClient(subID)
	recordsClient := dns.NewRecordSetsClient(subID)
	authorizer, _ := auth.NewAuthorizerFromEnvironment()
	zonesClient.Authorizer = authorizer
	recordsClient.Authorizer = authorizer

	// create the initial span for our app which will nest all the other ones through the context object.
	ctx, span := trace.StartSpan(context.Background(), "executeAzureAPICalls", trace.WithSampler(trace.AlwaysSample()))
	defer span.End()

	executeZonesAPICalls(ctx, zonesClient)

	// execute record set API calls
	for i := 0; i < iterations; i++ {
		executeRecordSetsAPICalls(ctx, recordsClient)
		time.Sleep(10 * time.Second)
	}

	if deleteZone {
		cleanupZones(ctx, zonesClient)
		time.Sleep(10 * time.Second)
	}
}

func executeZonesAPICalls(ctx context.Context, zonesClient dns.ZonesClient) {
	ctx, span := trace.StartSpan(ctx, "executeZonesAPICalls")
	defer span.End()
	zone, _ := zonesClient.CreateOrUpdate(ctx, rg, zoneName, dns.Zone{
		Location: to.StringPtr("global"),
	}, "", "")

	zoneGet, _ := zonesClient.Get(ctx, rg, *zone.Name)

	if *zoneGet.Name == *zone.Name {
		fmt.Println("Created zone: ", *zone.Name)
	}
}

func cleanupZones(ctx context.Context, zonesClient dns.ZonesClient) {
	ctx, span := trace.StartSpan(ctx, "cleanupZones")
	defer span.End()
	future, _ := zonesClient.Delete(ctx, rg, zoneName, "")
	//done, _ := future.Done(zonesClient)

	done := future.WaitForCompletionRef(ctx, zonesClient.Client)

	if done == nil {
		fmt.Println("Delete zone: ", done, future.Status())
	}

	fmt.Println("Delete finished: ", done)
}
func executeRecordSetsAPICalls(ctx context.Context, recordsClient dns.RecordSetsClient) {
	ctx, span := trace.StartSpan(ctx, "executeRecordSetsAPICalls")
	defer span.End()
	executeRRCreate(ctx, recordsClient)
	if getRR {
		executeRRGet(ctx, recordsClient)
	}
	if deleteRR {
		executeRRDelete(ctx, recordsClient)
	}
}

func executeRRCreate(ctx context.Context, recordsClient dns.RecordSetsClient) {
	ctx, span := trace.StartSpan(ctx, "executeRRCreate")
	defer span.End()
	for i := 0; i < rrCount; i++ {
		rr, _ := recordsClient.CreateOrUpdate(ctx, rg, zoneName, fmt.Sprintf("rr%d", i), dns.CNAME, dns.RecordSet{
			RecordSetProperties: &dns.RecordSetProperties{
				TTL: to.Int64Ptr(3600),
				CnameRecord: &dns.CnameRecord{
					Cname: to.StringPtr("vladdbCname"),
				},
			},
		},
			"", // if-match
			"", // if-none-match)
		)

		fmt.Println("Create RR: ", *rr.Name)
	}
}

func executeRRGet(ctx context.Context, recordsClient dns.RecordSetsClient) {
	ctx, span := trace.StartSpan(ctx, "executeRRGet")
	defer span.End()
	for i := 0; i < rrCount; i++ {
		rr, _ := recordsClient.Get(ctx, rg, zoneName, fmt.Sprintf("rr%d", i), dns.CNAME)
		fmt.Println("Get RR: ", *rr.Name)
	}
}

func executeRRDelete(ctx context.Context, recordsClient dns.RecordSetsClient) {
	ctx, span := trace.StartSpan(ctx, "executeRRDelete")
	defer span.End()
	for i := 0; i < rrCount; i++ {
		rrName := fmt.Sprintf("rr%d", i)
		_, _ = recordsClient.Delete(ctx, rg, zoneName, rrName, dns.CNAME, "")
		fmt.Println("Delete RR: ", rrName)
	}
}

func initZipkin() {
	localEndpoint, err := openzipkin.NewEndpoint("azureSDKZipkinTracing", "192.168.1.5:5454")
	if err != nil {
		log.Fatalf("Failed to create the local zipkinEndpoint: %v", err)
	}
	reporter := zipkinHTTP.NewReporter("http://localhost:9411/api/v2/spans")
	ze := zipkin.NewExporter(reporter, localEndpoint)
	// Register the Zipkin exporter.
	// This step is needed so that traces can be exported.
	trace.RegisterExporter(ze)
}
