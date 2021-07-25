package maps

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/2.0/creator"
)

func Example_aliasOperations() {
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
	aliasClient := creator.NewAliasClient(conn)

	// we need to upload resource first
	resourceUdid := uploadResource(dataClient, ctx, "resources/data_sample_upload.json", creator.UploadDataFormatGeojson, false)
	aliasCreateResp, createErr := aliasClient.Create(ctx, nil)
	if createErr != nil {
		util.LogAndPanic(createErr)
	}
	util.PrintAndLog("alias created")

	aliasId := *aliasCreateResp.AliasesCreateResponse.AliasID
	_, assignErr := aliasClient.Assign(ctx, aliasId, resourceUdid, nil)
	if assignErr != nil {
		util.LogAndPanic(assignErr)
	}
	util.PrintAndLog("alias assigned")

	_, detailsErr := aliasClient.GetDetails(ctx, aliasId, nil)
	if detailsErr != nil {
		util.LogAndPanic(detailsErr)
	}
	util.PrintAndLog("alias details retrieved")

	defer func() {
		_, deleteErr := aliasClient.Delete(ctx, aliasId, nil)
		if deleteErr != nil {
			util.LogAndPanic(deleteErr)
		}
		util.PrintAndLog("alias deleted")
	}()

	respPager := aliasClient.List(nil)
	for respPager.NextPage(ctx) {
		if respPager.Err() != nil {
			util.LogAndPanic(respPager.Err())
		}

		// do something with aliases
		// aliases := respPager.PageResponse().AliasListResponse.Aliases
	}
	util.PrintAndLog("aliases listed")

	// Output:
	// resource upload started: resources/data_sample_upload.json
	// resource upload completed: resources/data_sample_upload.json
	// alias created
	// alias assigned
	// alias details retrieved
	// aliases listed
	// alias deleted
}
