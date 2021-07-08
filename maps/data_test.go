package maps

import (
	"context"
	"errors"
	"flag"
	"os"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/2.0/creator"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
)

var usesSharedKeyAuth = flag.Bool("sharedkey-auth", false, "uses Azure Maps shared key authentication inside of AD if set")

func uploadAndThenDeleteResource(client *creator.DataClient, ctx context.Context, resource string, dataFormat creator.UploadDataFormat) {
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

	_, deleteErr := client.DeletePreview(ctx, match[0], nil)
	if deleteErr != nil {
		util.LogAndPanic(deleteErr)
	}
	util.PrintAndLog("resource deleted: " + resource)
}

func Example_uploadOperations() {
	ctx := context.Background()
	account := CreateResourceGroupWithMapAccount()
	defer resources.Cleanup(ctx)

	accountsClient := getAccountsClient()

	var (
		cred    azcore.Credential
		credErr error
	)

	if *usesSharedKeyAuth {
		keysResp, keysErr := accountsClient.ListKeys(ctx, config.GroupName(), *account.Name)
		if keysErr != nil {
			credErr = keysErr
		} else {
			cred = creator.SharedKeyCredential{SubscriptionKey: *keysResp.PrimaryKey}
		}
	} else {
		// service principal explicit auth
		cred, credErr = azidentity.NewClientSecretCredential(config.TenantID(), config.ClientID(), config.ClientSecret(), nil)
		// cred, credErr = azidentity.NewDefaultAzureCredential(nil)
	}

	if credErr != nil {
		util.LogAndPanic(credErr)
	}

	client := creator.NewDataClient(
		creator.NewConnection(creator.GeographyUs.ToPtr(), cred, nil),
		// xmsClientId can be nil for SharedKey auth
		account.Properties.XMsClientID,
	)

	uploadAndThenDeleteResource(client, ctx, "resources/data_sample_upload.zip", creator.UploadDataFormatDwgzippackage)
	uploadAndThenDeleteResource(client, ctx, "resources/data_sample_upload.json", creator.UploadDataFormatGeojson)

	// Output:
	// resource group created
	// account created
	// resource upload started: resources/data_sample_upload.zip
	// resource upload completed: resources/data_sample_upload.zip
	// resource deleted: resources/data_sample_upload.zip
	// resource upload started: resources/data_sample_upload.json
	// resource upload completed: resources/data_sample_upload.json
	// resource deleted: resources/data_sample_upload.json
}
