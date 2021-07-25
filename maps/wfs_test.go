package maps

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/2.0/creator"
)

func Example_wfsOperations() {
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
	datasetClient := creator.NewDatasetClient(conn)
	wfsClient := creator.NewWFSClient(conn, xmsClientId)

	resourceUdid := uploadResource(dataClient, ctx, "resources/data_sample_upload.zip", creator.UploadDataFormatDwgzippackage, false)
	conversionUdid := createConversion(conversionClient, ctx, resourceUdid, false)
	datasetUdid := createDataset(datasetClient, ctx, conversionUdid, false)

	getConfResp, getConfErr := wfsClient.GetConformance(ctx, datasetUdid, nil)
	if getConfErr != nil {
		util.LogAndPanic(getConfErr)
	}
	util.PrintAndLog("requirements classes retrieved")
	for index := range getConfResp.ConformanceResponse.ConformsTo {
		util.PrintAndLog(*getConfResp.ConformanceResponse.ConformsTo[index])
	}

	getLandingResp, getLandingErr := wfsClient.GetLandingPage(ctx, datasetUdid, nil)
	if getLandingErr != nil {
		util.LogAndPanic(getLandingErr)
	}
	util.PrintAndLog("metadata(landing page api) retrieved")
	util.PrintAndLog(*getLandingResp.LandingPageResponse.Ontology)
	json, jsonErr := getLandingResp.LandingPageResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	getCollResp, getCollErr := wfsClient.GetCollections(ctx, datasetUdid, nil)
	if getCollErr != nil {
		util.LogAndPanic(getCollErr)
	}
	util.PrintAndLog("collections retrieved")
	util.PrintAndLog(*getCollResp.CollectionsResponse.Ontology)
	json, jsonErr = getCollResp.CollectionsResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	getFacilityResp, getFacilityErr := wfsClient.GetCollection(ctx, datasetUdid, "facility", nil)
	if getFacilityErr != nil {
		util.LogAndPanic(getFacilityErr)
	}
	util.PrintAndLog("collection description retrieved")
	util.PrintAndLog(*getFacilityResp.CollectionInfo.Name)
	json, jsonErr = getFacilityResp.CollectionInfo.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	getCollDefResp, getCollDefErr := wfsClient.GetCollectionDefinition(ctx, datasetUdid, "facility", nil)
	if getCollDefErr != nil {
		util.LogAndPanic(getCollDefErr)
	}
	util.PrintAndLog("collection definition retrieved")
	util.PrintAndLog(*getCollDefResp.CollectionDefinitionResponse.Name)
	json, jsonErr = getCollDefResp.CollectionDefinitionResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	featuresResp, featuresRespErr := wfsClient.GetFeatures(ctx, datasetUdid, "facility", &creator.WFSGetFeaturesOptions{
		// Bbox:
		// Filter:
		// Limit:
	})
	if featuresRespErr != nil {
		util.LogAndPanic(featuresRespErr)
	}
	util.PrintAndLog("collection features retrived")
	util.PrintAndLog(fmt.Sprintf("Features in facility collection: %d", len(featuresResp.ExtendedGeoJSONFeatureCollection.Features)))
	if len(featuresResp.ExtendedGeoJSONFeatureCollection.Features) == 0 {
		util.LogAndPanic(fmt.Errorf("facility collection contains no features, which is unexpected"))
	}

	featureId := *featuresResp.ExtendedGeoJSONFeatureCollection.Features[0].ID
	featureResp, featureRespErr := wfsClient.GetFeature(ctx, datasetUdid, "facility", featureId, nil)
	if featureRespErr != nil {
		util.LogAndPanic(featureRespErr)
	}
	util.PrintAndLog("collection feature retrived")
	util.PrintAndLog(*featureResp.FeatureResponse.Ontology)
	json, jsonErr = featureResp.FeatureResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	_, deleteErr := wfsClient.DeleteFeature(ctx, datasetUdid, "facility", featureId, nil)
	if deleteErr != nil {
		util.LogAndPanic(deleteErr)
	}
	util.PrintAndLog("feature deleted")

	// Output:
	// resource upload started: resources/data_sample_upload.zip
	// resource upload completed: resources/data_sample_upload.zip
	// conversion started
	// conversion completed
	// dataset creation started
	// dataset creation completed
	// requirements classes retrieved
	// http://www.opengis.net/spec/wfs-1/3.0/req/core
	// http://www.opengis.net/spec/wfs-1/3.0/req/oas30
	// http://www.opengis.net/spec/wfs-1/3.0/req/geojson
	// http://tempuri.org/wfs/3.0/edit
	// metadata(landing page api) retrieved
	// facility-2.0
	// collections retrieved
	// facility-2.0
	// collection description retrieved
	// facility
	// collection definition retrieved
	// facility
	// collection features retrived
	// Features in facility collection: 1
	// collection feature retrived
	// facility-2.0
	// feature deleted
}
