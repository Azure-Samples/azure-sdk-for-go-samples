package maps

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/2.0/creator"
	"github.com/Azure/go-autorest/autorest/to"
)

func Example_spatialOperations() {
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
	spatialClient := creator.NewSpatialClient(conn, xmsClientId)

	resourceUdid := uploadResource(dataClient, ctx, "resources/data_sample_upload.json", creator.UploadDataFormatGeojson, false)
	bufferResp, getBufferErr := spatialClient.GetBuffer(ctx, creator.ResponseFormatJSON, resourceUdid, "176.3", nil)
	if getBufferErr != nil {
		util.LogAndPanic(getBufferErr)
	}

	switch geojson := bufferResp.BufferResponse.Result.(type) {
	case *creator.ExtendedGeoJSONFeatureCollection:
		if len(geojson.Features) != 2 {
			util.LogAndPanic(fmt.Errorf("Expected two features but found: %d", len(geojson.Features)))
		}

		switch geometry := geojson.Features[0].GeoJSONFeatureData.Geometry.(type) {
		case *creator.GeoJSONPolygon:
			util.PrintAndLog("fetched polygon buffer")
		default:
			util.LogAndPanic(fmt.Errorf("Encountered %s geometry while Polygon is expected", *geometry.GetGeoJSONGeometry().Type))
		}

	default:
		util.LogAndPanic(fmt.Errorf("Encountered %s which while FeatureCollection was expected", *geojson.GetGeoJSONObject().Type))
	}

	featureCollection, _ := bufferResp.BufferResponse.Result.(*creator.ExtendedGeoJSONFeatureCollection)
	json, jsonErr := featureCollection.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	closesPointsResp, getClosesPointsErr := spatialClient.GetClosestPoint(ctx, creator.ResponseFormatJSON, resourceUdid, 47.622942, -122.316456, &creator.SpatialGetClosestPointOptions{
		NumberOfClosestPoints: to.Int32Ptr(2),
	})
	if getClosesPointsErr != nil {
		util.LogAndPanic(getClosesPointsErr)
	}
	util.PrintAndLog(fmt.Sprintf("fetched %d closest points", len(closesPointsResp.ClosestPointResponse.Result)))
	json, jsonErr = closesPointsResp.ClosestPointResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	geofenceResp, getGeofenceErr := spatialClient.GetGeofence(ctx, creator.ResponseFormatJSON, "some_unique_device_name", resourceUdid, 48.36, -124.63, &creator.SpatialGetGeofenceOptions{
		Mode:         creator.GeofenceModeEnterAndExit.ToPtr(),
		SearchBuffer: to.Float32Ptr(50.0),
	})
	if getGeofenceErr != nil {
		util.LogAndPanic(getGeofenceErr)
	}
	util.PrintAndLog(fmt.Sprintf("fetched proximity with %d points", len(geofenceResp.GeofenceResponse.Geometries)))
	json, jsonErr = geofenceResp.GeofenceResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	gcDistResp, getGCDErr := spatialClient.GetGreatCircleDistance(ctx, creator.ResponseFormatJSON, "47.622942,-122.316456:47.610378,-122.200676", nil)
	if getGCDErr != nil {
		util.LogAndPanic(jsonErr)
	}

	distance := gcDistResp.GreatCircleDistanceResponse.Result.DistanceInMeters
	util.PrintAndLog(fmt.Sprintf("fetched greater circle distance: %.0f", float64(*distance)))

	// polygon_sample_upload.json
	polygonUdid := uploadResource(dataClient, ctx, "resources/polygon_sample_upload.json", creator.UploadDataFormatGeojson, false)
	pipResp, getPIPErr := spatialClient.GetPointInPolygon(ctx, creator.ResponseFormatJSON, polygonUdid, 47.64519559145717, -122.13093280792236, nil)
	if getPIPErr != nil {
		util.LogAndPanic(getPIPErr)
	}

	util.PrintAndLog(fmt.Sprintf("fetched point in polygon result: %t", *pipResp.PointInPolygonResponse.Result.PointInPolygons))
	bytes, err := ioutil.ReadFile("resources/spatial_buffer_request_body.json")
	if err != nil {
		util.LogAndPanic(err)
	}

	bufferRequestBody := creator.BufferRequestBody{}
	err = bufferRequestBody.UnmarshalJSON(bytes)
	if err != nil {
		util.LogAndPanic(err)
	}

	postBuffResp, postBuffErr := spatialClient.PostBuffer(ctx, creator.ResponseFormatJSON, bufferRequestBody, nil)
	if postBuffErr != nil {
		util.LogAndPanic(postBuffErr)
	}
	util.PrintAndLog("fetched polygon buffer")

	featureCollection, _ = postBuffResp.BufferResponse.Result.(*creator.ExtendedGeoJSONFeatureCollection)
	json, jsonErr = featureCollection.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	bytes, err = ioutil.ReadFile("resources/spatial_closest_point_request_body.json")
	if err != nil {
		util.LogAndPanic(err)
	}

	featureCollectionPayload := creator.GeoJSONFeatureCollection{}
	err = featureCollectionPayload.UnmarshalJSON(bytes)
	if err != nil {
		util.LogAndPanic(err)
	}

	postClosestPointRest, postClosestPointErr := spatialClient.PostClosestPoint(ctx, creator.ResponseFormatJSON, 47.622942, -122.316456, &featureCollectionPayload, &creator.SpatialPostClosestPointOptions{
		NumberOfClosestPoints: to.Int32Ptr(1),
	})
	if postClosestPointErr != nil {
		util.LogAndPanic(postClosestPointErr)
	}

	util.PrintAndLog(fmt.Sprintf("fetched %d closest point", len(postClosestPointRest.ClosestPointResponse.Result)))
	json, jsonErr = postClosestPointRest.ClosestPointResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	bytes, err = ioutil.ReadFile("resources/spatial_geofence_request_body.json")
	if err != nil {
		util.LogAndPanic(err)
	}

	featureCollectionPayload = creator.GeoJSONFeatureCollection{}
	err = featureCollectionPayload.UnmarshalJSON(bytes)
	if err != nil {
		util.LogAndPanic(err)
	}

	postGeofenceResp, postGeofenceErr := spatialClient.PostGeofence(ctx, creator.ResponseFormatJSON, "unique_device_name_under_account", 48.36, -124.63, &featureCollectionPayload, &creator.SpatialPostGeofenceOptions{
		Mode:         creator.GeofenceModeEnterAndExit.ToPtr(),
		SearchBuffer: to.Float32Ptr(50.0),
	})
	if postGeofenceErr != nil {
		util.LogAndPanic(postGeofenceErr)
	}

	util.PrintAndLog(fmt.Sprintf("fetched proximity with %d points", len(postGeofenceResp.GeofenceResponse.Geometries)))
	json, jsonErr = postGeofenceResp.GeofenceResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	bytes, err = ioutil.ReadFile("resources/spatial_point_in_polygon_request_body.json")
	if err != nil {
		util.LogAndPanic(err)
	}

	featureCollectionPayload = creator.GeoJSONFeatureCollection{}
	err = featureCollectionPayload.UnmarshalJSON(bytes)
	if err != nil {
		util.LogAndPanic(err)
	}

	postPipResp, postPipErr := spatialClient.PostPointInPolygon(ctx, creator.ResponseFormatJSON, 48.36, -124.63, &featureCollectionPayload, nil)
	if postPipErr != nil {
		util.LogAndPanic(postPipErr)
	}
	util.PrintAndLog(fmt.Sprintf("fetched point in polygon result: %t", *postPipResp.PointInPolygonResponse.Result.PointInPolygons))

	// console.log(" --- Post point in polygon:");
	// const postSpatialPointInPolygonPayload = JSON.parse(fs.readFileSync(filePathForPostSpatialPointInPolygon, "utf8"));
	// console.log(await spatial.postPointInPolygon("json", 48.36, -124.63, postSpatialPointInPolygonPayload, operationOptions));

	// Output:
	// resource upload started: resources/data_sample_upload.json
	// resource upload completed: resources/data_sample_upload.json
	// fetched polygon buffer
	// fetched 2 closest points
	// fetched proximity with 2 points
	// fetched greater circle distance: 8798
	// resource upload started: resources/polygon_sample_upload.json
	// resource upload completed: resources/polygon_sample_upload.json
	// fetched point in polygon result: true
	// fetched polygon buffer
	// fetched 1 closest point
	// fetched proximity with 2 points
	// fetched point in polygon result: false
}
