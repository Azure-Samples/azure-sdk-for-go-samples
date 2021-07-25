package maps

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/2.0/creator"
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/2.0/render"
	"github.com/Azure/go-autorest/autorest/to"
)

func saveResponseIntoTempFile(tempDirName string, fileName string, reader *io.ReadCloser) error {
	os.MkdirAll(filepath.Join(os.TempDir(), tempDirName), 0755)
	tilePath := filepath.Join(os.TempDir(), tempDirName, fileName)
	rasterTileFile, err := os.Create(tilePath)
	if err != nil {
		return err
	}

	defer rasterTileFile.Close()
	_, err = io.Copy(rasterTileFile, *reader)
	log.Println(fmt.Sprintf("written %s", tilePath))
	return err
}

func Example_renderV2Operations() {
	ctx := context.Background()
	accountsClient := getAccountsClient()
	cred, credErr := Authenticate(&accountsClient, ctx, *mapsAccount.Name, *usesADAuth)
	if credErr != nil {
		util.LogAndPanic(credErr)
	}

	conn := render.NewConnection(render.GeographyUs.ToPtr(), cred, nil)
	// xmsClientId doesn't need to be supplied for SharedKey auth
	var xmsClientId *string
	if *usesADAuth {
		xmsClientId = mapsAccount.Properties.XMsClientID
	}

	renderClient := render.NewRenderV2Client(conn, xmsClientId)
	tileData, err := renderClient.GetMapTilePreview(ctx, render.TilesetIDMicrosoftBase, 6, 10, 22, &render.RenderV2GetMapTilePreviewOptions{
		Language:  to.StringPtr("EN"),
		TileSize:  render.TileSizeFiveHundredTwelve.ToPtr(),
		TimeStamp: to.StringPtr(time.Now().Format(time.RFC3339)),
		View:      to.StringPtr("Auto"),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	err = saveResponseIntoTempFile("tilecache/6/10", "22.mvt", &tileData.RawResponse.Body)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved and saved vector tile")

	// Output:
	// retrieved and saved vector tile
}

func Example_renderOperations() {
	ctx := context.Background()
	accountsClient := getAccountsClient()
	cred, credErr := Authenticate(&accountsClient, ctx, *mapsAccount.Name, *usesADAuth)
	if credErr != nil {
		util.LogAndPanic(credErr)
	}

	conn := render.NewConnection(render.GeographyUs.ToPtr(), cred, nil)
	// xmsClientId doesn't need to be supplied for SharedKey auth
	var xmsClientId *string
	if *usesADAuth {
		xmsClientId = mapsAccount.Properties.XMsClientID
	}

	renderClient := render.NewRenderClient(conn, xmsClientId)

	captionResp, err := renderClient.GetCopyrightCaption(ctx, render.TextFormatJSON, nil)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved copyright caption")
	log.Println(*captionResp.GetCopyrightCaptionResult.CopyrightsCaption)

	copyrightResp, err := renderClient.GetCopyrightForTile(ctx, render.TextFormatJSON, 6, 9, 22, &render.RenderGetCopyrightForTileOptions{
		Text: render.IncludeTextNo.ToPtr(),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved copyright for tile")
	json, jsonErr := copyrightResp.GetCopyrightForTileResult.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	copyrightWorldResp, err := renderClient.GetCopyrightForWorld(ctx, render.TextFormatJSON, &render.RenderGetCopyrightForWorldOptions{
		Text: render.IncludeTextNo.ToPtr(),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved copyright for world")
	json, jsonErr = copyrightWorldResp.GetCopyrightForWorldResult.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	copyrightForBboxResp, err := renderClient.GetCopyrightFromBoundingBox(ctx, render.TextFormatJSON, "52.41064,4.84228", "52.41072,4.84239", &render.RenderGetCopyrightFromBoundingBoxOptions{
		Text: render.IncludeTextNo.ToPtr(),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved copyright for bbox")
	json, jsonErr = copyrightForBboxResp.GetCopyrightFromBoundingBoxResult.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	imageTileResp, err := renderClient.GetMapImageryTile(ctx, render.RasterTileFormatPNG, render.MapImageryStyleSatellite, 6, 10, 22, nil)
	if err != nil {
		util.LogAndPanic(err)
	}
	err = saveResponseIntoTempFile("tilecache/6/10", "22.png", &imageTileResp.RawResponse.Body)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved and saved raster tile")

	mapStaticImgResp, err := renderClient.GetMapStaticImage(ctx, render.RasterTileFormatPNG, &render.RenderGetMapStaticImageOptions{
		Layer: render.StaticMapLayerHybrid.ToPtr(),
		Zoom:  to.Int32Ptr(2),
		Bbox:  to.StringPtr("1.355233,42.982261,24.980233,56.526017"),
		Style: render.MapImageStyleDark.ToPtr(),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	err = saveResponseIntoTempFile("tilecache", "eu_staticsample.png", &mapStaticImgResp.RawResponse.Body)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved and saved map static image")

	mapTileResp, err := renderClient.GetMapTile(ctx, render.TileFormatPNG, render.MapTileLayerBasic, render.MapTileStyleDark, 6, 10, 22, &render.RenderGetMapTileOptions{
		Language: to.StringPtr("EN"),
		TileSize: render.MapTileSizeFiveHundredTwelve.ToPtr(),
		View:     to.StringPtr("Auto"),
	})
	err = saveResponseIntoTempFile("tilecache", "wa_tile.png", &mapTileResp.RawResponse.Body)
	util.PrintAndLog("retrieved and saved map tile")

	// stateset creation
	creatorConn := creator.NewConnection(creator.GeographyUs.ToPtr(), cred, nil)
	dataClient := creator.NewDataClient(creatorConn, xmsClientId)
	conversionClient := creator.NewConversionClient(creatorConn, xmsClientId)
	datasetClient := creator.NewDatasetClient(creatorConn)
	featureStateClient := creator.NewFeatureStateClient(creatorConn)
	resourceUdid := uploadResource(dataClient, ctx, "resources/data_sample_upload.zip", creator.UploadDataFormatDwgzippackage, false)
	conversionUdid := createConversion(conversionClient, ctx, resourceUdid, false)
	datasetUdid := createDataset(datasetClient, ctx, conversionUdid, false)
	statesetId := createStateset(featureStateClient, ctx, datasetUdid, "resources/featurestate_sample_create.json", false)
	stateTileResp, err := renderClient.GetMapStateTilePreview(ctx, 6, 10, 22, statesetId, nil)
	if err != nil {
		util.LogAndPanic(err)
	}
	err = saveResponseIntoTempFile("tilecache/6/10", "22.state.mvt", &stateTileResp.RawResponse.Body)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved and saved state tile")

	// Output:
	// retrieved copyright caption
	// retrieved copyright for tile
	// retrieved copyright for world
	// retrieved copyright for bbox
	// retrieved and saved raster tile
	// retrieved and saved map static image
	// retrieved and saved map tile
	// resource upload started: resources/data_sample_upload.zip
	// resource upload completed: resources/data_sample_upload.zip
	// conversion started
	// conversion completed
	// dataset creation started
	// dataset creation completed
	// stateset created
	// retrieved and saved state tile
}
