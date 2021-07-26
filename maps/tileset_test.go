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

func Example_tilesetOperations() {
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
	tilesetClient := creator.NewTilesetClient(conn, xmsClientId)

	resourceUdid := uploadResource(dataClient, ctx, "resources/data_sample_upload.zip", creator.UploadDataFormatDwgzippackage, false)
	conversionUdid := createConversion(conversionClient, ctx, resourceUdid, false)
	datasetUdid := createDataset(datasetClient, ctx, conversionUdid, false)

	createResp, createErr := tilesetClient.BeginCreate(ctx, datasetUdid, &creator.TilesetBeginCreateOptions{
		Description: to.StringPtr("test tileset"),
	})
	if createErr != nil {
		util.LogAndPanic(createErr)
	}
	util.PrintAndLog("tileset creation started")

	createOpResp, createOpErr := createResp.PollUntilDone(ctx, 1*time.Second)
	if createOpErr != nil {
		util.LogAndPanic(createOpErr)
	}
	util.PrintAndLog("tileset creation completed")

	resourceLocation := *createOpResp.ResourceLocation
	if len(resourceLocation) == 0 {
		util.LogAndPanic(errors.New("Resource location should not be empty."))
	}

	uuidExpr := regexp.MustCompile(`[0-9A-Fa-f\-]{36}`)
	match := uuidExpr.FindStringSubmatch(resourceLocation)
	if len(match) == 0 {
		util.LogAndPanic(errors.New("Unable to extract resource uuid from resource location."))
	}

	tilesetUdid := match[0]
	defer func() {
		_, deleteErr := tilesetClient.Delete(ctx, tilesetUdid, nil)
		if deleteErr != nil {
			util.LogAndPanic(deleteErr)
		}
		util.PrintAndLog("tileset deleted")
	}()

	_, detailsErr := tilesetClient.Get(ctx, tilesetUdid, nil)
	if detailsErr != nil {
		util.LogAndPanic(detailsErr)
	}
	util.PrintAndLog("tileset details retrieved")

	respPager := tilesetClient.List(nil)
	for respPager.NextPage(ctx) {
		if respPager.Err() != nil {
			util.LogAndPanic(respPager.Err())
		}

		// do something with tileset
		util.PrintAndLog(fmt.Sprintf("tileset listed: %d tileset", len(respPager.PageResponse().TilesetListResponse.Tilesets)))
	}

	// Output:
	// resource upload started: resources/data_sample_upload.zip
	// resource upload completed: resources/data_sample_upload.zip
	// conversion started
	// conversion completed
	// dataset creation started
	// dataset creation completed
	// tileset creation started
	// tileset creation completed
	// tileset details retrieved
	// tileset listed: 1 tileset
	// tileset deleted
}
