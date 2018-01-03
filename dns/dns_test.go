package dns

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

var (
	zoneName        = "az-sample-zone.local"
	aRecordName     = "testA"
	cnameRecordName = "testB"
)

func TestMain(m *testing.M) {
	flag.StringVar(&zoneName, "zoneName", zoneName, "Specify name of DNS zone to create.")

	err := helpers.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}
	os.Exit(m.Run())
}

func ExampleDnsZone() {
	helpers.SetResourceGroupName("")
	ctx := context.Background()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	_, err = CreateZone(ctx, zoneName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	} else {
		helpers.PrintAndLog("zone created")
	}

	_, err = CreateARecordSet(ctx, zoneName, aRecordName, "127.0.0.127")
	if err != nil {
		helpers.PrintAndLog(err.Error())
	} else {
		helpers.PrintAndLog("A record set created")
	}

	_, err = CreateCNAMERecordSet(ctx, zoneName, cnameRecordName, "127.0.0.128")
	if err != nil {
		helpers.PrintAndLog(err.Error())
	} else {
		helpers.PrintAndLog("CNAME record set created")
	}

	records, err := GetRecordsForZone(ctx, zoneName)
	if err != nil {
		helpers.PrintAndLog(err.Error())
	} else {
		helpers.PrintAndLog("records retrieved")
	}

	for _, record := range records {
		log.Printf("%s\n",
			*record.RecordSetProperties.Fqdn)
	}

	// Output:
	// zone created
	// A record set created
	// CNAME record set created
	// records retrieved
}
