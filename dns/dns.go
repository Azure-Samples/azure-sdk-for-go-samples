package dns

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"

	dns "github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2017-09-01/dns"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

// A dns.ZonesClient manages DNS zones (as opposed to records within zones).
func getZonesClient() dns.ZonesClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	zonesClient := dns.NewZonesClient(helpers.SubscriptionID())
	zonesClient.Authorizer = autorest.NewBearerAuthorizer(token)
	zonesClient.AddToUserAgent(helpers.UserAgent())
	return zonesClient
}

func getRecordSetsClient() dns.RecordSetsClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	recordsClient := dns.NewRecordSetsClient(helpers.SubscriptionID())
	recordsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	recordsClient.AddToUserAgent(helpers.UserAgent())
	return recordsClient
}

// CreateZone creates a new DNS zone (the top-level mgmt object for DNS in Azure).
func CreateZone(ctx context.Context, zoneName string) (dns.Zone, error) {
	zonesClient := getZonesClient()
	return zonesClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		zoneName,
		dns.Zone{
			Location: to.StringPtr("global"),
		},
		"",  // if-match
		"*", // if-none-match
	)
}

// DeleteZone deletes a DNS zone with the given name.
func DeleteZone(ctx context.Context, zoneName string) (r autorest.Response, err error) {
	zonesClient := getZonesClient()
	future, err := zonesClient.Delete(ctx, helpers.ResourceGroupName(), zoneName, "")

	if err != nil {
		return r, fmt.Errorf("cannot delete zone: %v", err)
	}

	err = future.WaitForCompletion(ctx, zonesClient.Client)
	if err != nil {
		return r, fmt.Errorf("cannot get the zone delete future response: %v", err)
	}

	return future.Result(zonesClient)
}

func CreateARecordSet(ctx context.Context, zoneName, hostName, address string) (dns.RecordSet, error) {
	recordSetsClient := getRecordSetsClient()
	return recordSetsClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		zoneName,
		hostName,
		dns.A,
		dns.RecordSet{
			RecordSetProperties: &dns.RecordSetProperties{
				TTL: to.Int64Ptr(3600),
				ARecords: &[]dns.ARecord{
					{Ipv4Address: to.StringPtr(address)},
				},
			},
		},
		"", // if-match
		"", // if-none-match
	)
}

func CreateCNAMERecordSet(ctx context.Context, zoneName, hostName, cName string) (dns.RecordSet, error) {
	recordSetsClient := getRecordSetsClient()
	return recordSetsClient.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		zoneName,
		hostName,
		dns.CNAME,
		dns.RecordSet{
			RecordSetProperties: &dns.RecordSetProperties{
				TTL: to.Int64Ptr(3600),
				CnameRecord: &dns.CnameRecord{
					Cname: to.StringPtr(cName),
				},
			},
		},
		"", // if-match
		"", // if-none-match
	)
}

func GetRecordsForZone(ctx context.Context, zoneName string) ([]dns.RecordSet, error) {
	recordSetsClient := getRecordSetsClient()
	records, err := recordSetsClient.ListByDNSZone(
		ctx,
		helpers.ResourceGroupName(),
		zoneName,
		to.Int32Ptr(1000),
		"", // recordSetNameSuffic
	)
	return records.Values(), err
}
