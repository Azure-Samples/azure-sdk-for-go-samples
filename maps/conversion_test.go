package maps

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/2.0/creator"
	"github.com/Azure/go-autorest/autorest/to"
)

func createConversion(client *creator.ConversionClient, ctx context.Context, resource string, shouldDelete bool) string {
	convertResp, convertErr := client.BeginConvert(ctx, resource, "facility-2.0", &creator.ConversionBeginConvertOptions{
		Description: to.StringPtr("sample conversion description"),
	})
	if convertErr != nil {
		util.LogAndPanic(convertErr)
	}
	util.PrintAndLog("conversion started")

	convertOpResp, convertOpErr := convertResp.PollUntilDone(ctx, 1*time.Second)
	if convertOpErr != nil {
		util.LogAndPanic(convertOpErr)
	}
	util.PrintAndLog("conversion completed")

	resourceLocation := *convertOpResp.ResourceLocation
	if len(resourceLocation) == 0 {
		util.LogAndPanic(errors.New("Resource location should not be empty."))
	}

	uuidExpr := regexp.MustCompile(`[0-9A-Fa-f\-]{36}`)
	match := uuidExpr.FindStringSubmatch(resourceLocation)
	if len(match) == 0 {
		util.LogAndPanic(errors.New("Unable to extract resource uuid from resource location."))
	}

	if shouldDelete {
		defer func() {
			_, deleteErr := client.Delete(ctx, match[0], nil)
			if deleteErr != nil {
				util.LogAndPanic(deleteErr)
			}
			util.PrintAndLog("conversion deleted")
		}()
	}

	return match[0]
}

func Example_conversionOperations() {
	ctx := context.Background()
	accountsClient := getAccountsClient()
	cred, credErr := Authenticate(&accountsClient, ctx, *mapsAccount.Name, *usesADAuth)
	if credErr != nil {
		util.LogAndPanic(credErr)
	}

	conn := creator.NewConnection(creator.GeographyUs.ToPtr(), cred, nil)
	// xmsClientId doesn't need to be supplied for SharedKey auth
	var xmsClientId *string
	if *usesADAuth {
		xmsClientId = mapsAccount.Properties.XMsClientID
	}

	dataClient := creator.NewDataClient(conn, xmsClientId)
	conversionClient := creator.NewConversionClient(conn, xmsClientId)

	// we need to upload resource first
	resourceUdid := uploadResource(dataClient, ctx, "resources/data_sample_upload.zip", creator.UploadDataFormatDwgzippackage, false)

	conversionId := createConversion(conversionClient, ctx, resourceUdid, false)
	defer func() {
		_, deleteErr := conversionClient.Delete(ctx, conversionId, nil)
		if deleteErr != nil {
			util.LogAndPanic(deleteErr)
		}
		util.PrintAndLog("conversion deleted")
	}()

	_, detailsErr := conversionClient.Get(ctx, conversionId, nil)
	if detailsErr != nil {
		util.LogAndPanic(detailsErr)
	}
	util.PrintAndLog("conversion details retrieved")

	respPager := conversionClient.List(nil)
	for respPager.NextPage(ctx) {
		if respPager.Err() != nil {
			util.LogAndPanic(respPager.Err())
		}

		// do something with conversions
		// conversions := respPager.PageResponse().ConversionListResponse.Conversions
	}
	util.PrintAndLog("conversions listed")

	// Output:
	// resource upload started: resources/data_sample_upload.zip
	// resource upload completed: resources/data_sample_upload.zip
	// conversion started
	// conversion completed
	// conversion details retrieved
	// conversions listed
	// conversion deleted
}
