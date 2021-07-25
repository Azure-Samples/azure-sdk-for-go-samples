package maps

import (
	"context"
	"errors"
	"os"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/maps/2.0/creator"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

func uploadResource(client *creator.DataClient, ctx context.Context, resource string, dataFormat creator.UploadDataFormat, shouldDelete bool) string {
	file, openErr := os.Open(resource)
	if openErr != nil {
		util.LogAndPanic(openErr)
	}
	defer file.Close()

	uploadOperation, uploadErr := client.BeginUploadPreview(ctx, dataFormat, file, nil)
	if uploadErr != nil {
		util.LogAndPanic(uploadErr)
	}
	util.PrintAndLog("resource upload started: " + resource)

	resp, uploadErr := uploadOperation.PollUntilDone(ctx, 1*time.Second)
	if uploadErr != nil {
		util.LogAndPanic(uploadErr)
	}
	util.PrintAndLog("resource upload completed: " + resource)

	resourceLocation := *resp.ResourceLocation
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
			_, deleteErr := client.DeletePreview(ctx, match[0], nil)
			if deleteErr != nil {
				util.LogAndPanic(deleteErr)
			}
			util.PrintAndLog("resource deleted: " + resource)
		}()
	}

	return match[0]
}

func Example_uploadOperations() {
	ctx := context.Background()
	defer resources.Cleanup(ctx)
	accountsClient := getAccountsClient()
	cred, credErr := Authenticate(&accountsClient, ctx, *mapsAccount.Name, *usesADAuth)
	if credErr != nil {
		util.LogAndPanic(credErr)
	}

	// xmsClientId doesn't need to be supplied for SharedKey auth
	var xmsClientId *string
	if *usesADAuth {
		xmsClientId = mapsAccount.Properties.XMsClientID
	}

	client := creator.NewDataClient(creator.NewConnection(creator.GeographyUs.ToPtr(), cred, nil), xmsClientId)
	uploadResource(client, ctx, "resources/data_sample_upload.zip", creator.UploadDataFormatDwgzippackage, true)
	uploadResource(client, ctx, "resources/data_sample_upload.json", creator.UploadDataFormatGeojson, true)

	_, listErr := client.ListPreview(ctx, nil)
	if listErr != nil {
		util.LogAndPanic(listErr)
	}
	util.PrintAndLog("resources listed")

	// Output:
	// resource upload started: resources/data_sample_upload.zip
	// resource upload completed: resources/data_sample_upload.zip
	// resource deleted: resources/data_sample_upload.zip
	// resource upload started: resources/data_sample_upload.json
	// resource upload completed: resources/data_sample_upload.json
	// resource deleted: resources/data_sample_upload.json
	// resources listed
}
