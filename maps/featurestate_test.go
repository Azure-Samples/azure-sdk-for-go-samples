package maps

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/2.0/creator"
	"github.com/Azure/go-autorest/autorest/to"
)

func createStateset(client *creator.FeatureStateClient, ctx context.Context, datasetUdid string, statesetPath string, shouldDelete bool) string {
	data, readRrr := ioutil.ReadFile(statesetPath)
	if readRrr != nil {
		util.LogAndPanic(readRrr)
	}
	stateSet := creator.StylesObject{}
	jsonErr := stateSet.UnmarshalJSON(data)
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}

	createResp, createErr := client.CreateStateset(ctx, datasetUdid, stateSet, &creator.FeatureStateCreateStatesetOptions{
		Description: to.StringPtr("Test feature state"),
	})
	if createErr != nil {
		util.LogAndPanic(createErr)
	}
	util.PrintAndLog("stateset created")

	statesetId := *createResp.StatesetCreatedResponse.StatesetID
	if shouldDelete {
		defer func() {
			_, deleteErr := client.DeleteStateset(ctx, statesetId, nil)
			if deleteErr != nil {
				util.LogAndPanic(deleteErr)
			}
			util.PrintAndLog("stateset deleted")
		}()
	}

	return statesetId
}

func Example_featurestateOperations() {
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
	featureStateClient := creator.NewFeatureStateClient(conn, xmsClientId)

	resourceUdid := uploadResource(dataClient, ctx, "resources/data_sample_upload.zip", creator.UploadDataFormatDwgzippackage, false)
	conversionUdid := createConversion(conversionClient, ctx, resourceUdid, false)
	datasetUdid := createDataset(datasetClient, ctx, conversionUdid, false)
	statesetId := createStateset(featureStateClient, ctx, datasetUdid, "resources/featurestate_sample_create.json", false)
	// this is taken from WFS GetFeatures call with "facility" as a collection
	const featureId = "FCL13"

	defer func() {
		_, deleteErr := featureStateClient.DeleteStateset(ctx, statesetId, nil)
		if deleteErr != nil {
			util.LogAndPanic(deleteErr)
		}
		util.PrintAndLog("stateset deleted")
	}()

	_, getErr := featureStateClient.GetStateset(ctx, statesetId, nil)
	if getErr != nil {
		util.LogAndPanic(getErr)
	}
	util.PrintAndLog("stateset retrieved")

	stateUpdate := creator.FeatureStatesStructure{
		States: []*creator.FeatureStateObject{
			{
				KeyName:        to.StringPtr("s1"),
				Value:          true,
				EventTimestamp: to.StringPtr(time.Now().Format("2006-01-02 15:04:05")),
			},
		},
	}
	_, updateErr := featureStateClient.UpdateStates(ctx, statesetId, featureId, stateUpdate, nil)
	if updateErr != nil {
		util.LogAndPanic(updateErr)
	}
	stateUpdate = creator.FeatureStatesStructure{
		States: []*creator.FeatureStateObject{
			{
				KeyName:        to.StringPtr("s3"),
				Value:          "stateValue2",
				EventTimestamp: to.StringPtr(time.Now().Format("2006-01-02 15:04:05")),
			},
		},
	}
	_, updateErr = featureStateClient.UpdateStates(ctx, statesetId, featureId, stateUpdate, nil)
	if updateErr != nil {
		util.LogAndPanic(updateErr)
	}
	util.PrintAndLog("states updated")

	stateResp, stateErr := featureStateClient.GetStates(ctx, statesetId, featureId, nil)
	if stateErr != nil {
		util.LogAndPanic(stateErr)
	}

	states := stateResp.FeatureStatesStructure.States
	util.PrintAndLog(fmt.Sprintf("states retrieved: %d state", len(states)))
	util.PrintAndLog(fmt.Sprintf("s3: %s", states[0].Value))

	data, readRrr := ioutil.ReadFile("resources/featurestate_sample_put.json")
	if readRrr != nil {
		util.LogAndPanic(readRrr)
	}
	stateSet := creator.StylesObject{}
	jsonErr := stateSet.UnmarshalJSON(data)
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}

	_, updateErr = featureStateClient.PutStateset(ctx, statesetId, stateSet, nil)
	if updateErr != nil {
		util.LogAndPanic(updateErr)
	}
	util.PrintAndLog("stateset updated")

	_, deleteErr := featureStateClient.DeleteState(ctx, statesetId, featureId, "s3", nil)
	if deleteErr != nil {
		util.LogAndPanic(deleteErr)
	}
	util.PrintAndLog("state deleted")

	stateResp, stateErr = featureStateClient.GetStates(ctx, statesetId, featureId, nil)
	if stateErr != nil {
		util.LogAndPanic(stateErr)
	}
	util.PrintAndLog(fmt.Sprintf("states retrieved: %d states", len(stateResp.FeatureStatesStructure.States)))

	respPager := featureStateClient.ListStateset(nil)
	for respPager.NextPage(ctx) {
		if respPager.Err() != nil {
			util.LogAndPanic(respPager.Err())
		}

		// do something with datasets
		util.PrintAndLog(fmt.Sprintf("statesets listed: %d stateset", len(respPager.PageResponse().StatesetListResponse.Statesets)))
	}

	// Output:
	// resource upload started: resources/data_sample_upload.zip
	// resource upload completed: resources/data_sample_upload.zip
	// conversion started
	// conversion completed
	// dataset creation started
	// dataset creation completed
	// stateset created
	// stateset retrieved
	// states updated
	// states retrieved: 1 state
	// s3: stateValue2
	// stateset updated
	// state deleted
	// states retrieved: 0 states
	// statesets listed: 1 stateset
	// stateset deleted
}
