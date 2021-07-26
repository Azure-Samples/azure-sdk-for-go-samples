package maps

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/1.0/elevation"
	"github.com/Azure/go-autorest/autorest/to"
)

func Example_elevationOperations() {
	ctx := context.Background()
	accountsClient := getAccountsClient()
	cred, credErr := Authenticate(&accountsClient, ctx, *mapsAccount.Name, *usesADAuth)
	if credErr != nil {
		util.LogAndPanic(credErr)
	}

	conn := elevation.NewConnection(elevation.GeographyUs.ToPtr(), cred, nil)
	// xmsClientId doesn't need to be supplied for SharedKey auth
	var xmsClientId *string
	if *usesADAuth {
		xmsClientId = mapsAccount.Properties.UniqueID
	}

	elevationClient := elevation.NewElevationClient(conn, xmsClientId)
	dataForBboxResp, err := elevationClient.GetDataForBoundingBox(ctx, elevation.ResponseFormatJSON, []string{"-121.66853362143818", "46.84646479863713", "-121.65853362143818", "46.85646479863713"}, 3, 3, nil)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("fetched data for bounding box")
	util.PrintAndLog(fmt.Sprintf("elevation: %.0f", float64(*dataForBboxResp.BoundingBoxResult.Data[0].ElevationInMeter)))
	json, jsonErr := dataForBboxResp.BoundingBoxResult.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	dataForPointsResp, err := elevationClient.GetDataForPoints(ctx, elevation.ResponseFormatJSON, []string{"-121.66853362143818,46.84646479863713", "-121.65853362143818,46.85646479863713"}, nil)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("fetched data for points")
	util.PrintAndLog(fmt.Sprintf("elevation: %.0f", float64(*dataForPointsResp.PointsResult.Data[0].ElevationInMeter)))
	json, jsonErr = dataForPointsResp.PointsResult.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	dataForPolyResp, err := elevationClient.GetDataForPolyline(ctx, elevation.ResponseFormatJSON, []string{"-121.66853362143818,46.84646479863713", "-121.65853362143818,46.85646479863713"}, &elevation.ElevationGetDataForPolylineOptions{
		Samples: to.Int32Ptr(10),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("fetched data for polyline")
	util.PrintAndLog(fmt.Sprintf("elevation: %.0f", float64(*dataForPolyResp.LinesResult.Data[0].ElevationInMeter)))
	json, jsonErr = dataForPolyResp.LinesResult.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	postDataForPointsResp, err := elevationClient.PostDataForPoints(ctx, elevation.ResponseFormatJSON, []*elevation.CoordinatesPairAbbreviated{
		{
			Lat: to.Float64Ptr(46.84646479863713),
			Lon: to.Float64Ptr(-121.66853362143818),
		},
		{
			Lat: to.Float64Ptr(46.85646479863713),
			Lon: to.Float64Ptr(-121.65853362143818),
		},
	}, nil)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("posted data for points")
	util.PrintAndLog(fmt.Sprintf("elevation: %.0f", float64(*postDataForPointsResp.PointsResult.Data[0].ElevationInMeter)))
	json, jsonErr = postDataForPointsResp.PointsResult.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	postDataForPolylineResp, err := elevationClient.PostDataForPolyline(ctx, elevation.ResponseFormatJSON, []*elevation.CoordinatesPairAbbreviated{
		{
			Lat: to.Float64Ptr(46.84646479863713),
			Lon: to.Float64Ptr(-121.66853362143818),
		},
		{
			Lat: to.Float64Ptr(46.85646479863713),
			Lon: to.Float64Ptr(-121.65853362143818),
		},
	}, nil)
	util.PrintAndLog("posted data for polyline")
	util.PrintAndLog(fmt.Sprintf("elevation: %.0f", float64(*postDataForPolylineResp.LinesResult.Data[0].ElevationInMeter)))
	json, jsonErr = postDataForPolylineResp.LinesResult.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(json))

	// Output:
	// fetched data for bounding box
	// elevation: 2299
	// fetched data for points
	// elevation: 2299
	// fetched data for polyline
	// elevation: 2299
	// posted data for points
	// elevation: 2299
	// posted data for polyline
	// elevation: 2299
}
