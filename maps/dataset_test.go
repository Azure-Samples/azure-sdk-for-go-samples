package maps

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/2.0/creator"
	"github.com/Azure/go-autorest/autorest/to"
)

func createDataset(client *creator.DatasetClient, ctx context.Context, conversion string, shouldDelete bool) string {
	createResp, createErr := client.BeginCreate(ctx, conversion, &creator.DatasetBeginCreateOptions{
		DescriptionDataset: to.StringPtr("test dataset"),
	})
	if createErr != nil {
		util.LogAndPanic(createErr)
	}
	util.PrintAndLog("dataset creation started")

	createOpResp, createOpErr := createResp.PollUntilDone(ctx, 1*time.Second)
	if createOpErr != nil {
		util.LogAndPanic(createOpErr)
	}
	util.PrintAndLog("dataset creation completed")

	resourceLocation := *createOpResp.ResourceLocation
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
			util.PrintAndLog("dataset deleted")
		}()
	}

	return match[0]
}

func Example_datasetOperations() {
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
		xmsClientId = mapsAccount.Properties.UniqueID
	}

	dataClient := creator.NewDataClient(conn, xmsClientId)
	conversionClient := creator.NewConversionClient(conn, xmsClientId)
	datasetClient := creator.NewDatasetClient(conn, xmsClientId)

	resourceUdid := uploadResource(dataClient, ctx, "resources/data_sample_upload.zip", creator.UploadDataFormatDwgzippackage, false)
	conversionUdid := createConversion(conversionClient, ctx, resourceUdid, false)
	datasetUdid := createDataset(datasetClient, ctx, conversionUdid, false)
	defer func() {
		_, deleteErr := datasetClient.Delete(ctx, datasetUdid, nil)
		if deleteErr != nil {
			util.LogAndPanic(deleteErr)
		}
		util.PrintAndLog("dataset deleted")
	}()

	_, detailsErr := datasetClient.Get(ctx, datasetUdid, nil)
	if detailsErr != nil {
		util.LogAndPanic(detailsErr)
	}
	util.PrintAndLog("dataset details retrieved")

	respPager := datasetClient.List(nil)
	for respPager.NextPage(ctx) {
		if respPager.Err() != nil {
			util.LogAndPanic(respPager.Err())
		}

		// do something with datasets
		util.PrintAndLog(fmt.Sprintf("datasets listed: %d dataset", len(respPager.PageResponse().DatasetListResponse.Datasets)))
	}

	// Output:
	// resource upload started: resources/data_sample_upload.zip
	// resource upload completed: resources/data_sample_upload.zip
	// conversion started
	// conversion completed
	// dataset creation started
	// dataset creation completed
	// dataset details retrieved
	// datasets listed: 1 dataset
	// dataset deleted
}
