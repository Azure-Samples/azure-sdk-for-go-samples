package maps

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/1.0/route"
	"github.com/Azure/go-autorest/autorest/to"
)

func Example_routeOperations() {
	ctx := context.Background()
	accountsClient := getAccountsClient()
	cred, credErr := Authenticate(&accountsClient, ctx, *mapsAccount.Name, *usesADAuth)
	if credErr != nil {
		util.LogAndPanic(credErr)
	}

	conn := route.NewConnection(route.GeographyUs.ToPtr(), cred, nil)
	// xmsClientId doesn't need to be supplied for SharedKey auth
	var xmsClientId *string
	if *usesADAuth {
		xmsClientId = mapsAccount.Properties.UniqueID
	}

	routeClient := route.NewRouteClient(conn, xmsClientId)
	directionsResp, err := routeClient.GetRouteDirections(ctx, route.TextFormatJSON, "52.50931,13.42936:52.50274,13.43872", &route.RouteGetRouteDirectionsOptions{
		// make sure to check options available
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved route directions")
	jsonDirections, jsonErr := directionsResp.RouteDirectionsResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonDirections))

	routeRangeResp, err := routeClient.GetRouteRange(ctx, route.TextFormatJSON, "50.97452,5.86605", &route.RouteGetRouteRangeOptions{
		// make sure to check options available
		TimeBudgetInSec: to.Float32Ptr(6000.0),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved route range")
	jsonRouteRange, jsonErr := routeRangeResp.GetRouteRangeResponse.ReachableRange.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonRouteRange))

	bytes, err := ioutil.ReadFile("resources/route_directions_request_body.json")
	if err != nil {
		util.LogAndPanic(err)
	}
	postRouteDirectionBody := route.PostRouteDirectionsRequestBody{}
	err = json.Unmarshal(bytes, &postRouteDirectionBody)
	if err != nil {
		util.LogAndPanic(err)
	}
	postRouteDirectionResp, err := routeClient.PostRouteDirections(ctx, route.TextFormatJSON, "52.50931,13.42936:52.50274,13.43872", postRouteDirectionBody, &route.RoutePostRouteDirectionsOptions{
		// make sure to check options available
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("posted route direction")
	jsonDirections, err = postRouteDirectionResp.RouteDirectionsResponse.MarshalJSON()
	if err != nil {
		util.LogAndPanic(err)
	}
	log.Println(string(jsonDirections))

	bytes, err = ioutil.ReadFile("resources/route_directions_batch_request_body.json")
	if err != nil {
		util.LogAndPanic(err)
	}

	batchRequestBody := route.BatchRequestBody{}
	err = json.Unmarshal(bytes, &batchRequestBody)
	if err != nil {
		util.LogAndPanic(err)
	}
	resp, err := routeClient.BeginPostRouteDirectionsBatch(ctx, route.ResponseFormatJSON, batchRequestBody, nil)
	batchDirectionsResp, err := resp.PollUntilDone(ctx, 1*time.Second)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("posted route direction batch")
	jsonDirections, err = batchDirectionsResp.RouteDirectionsBatchResponse.MarshalJSON()
	if err != nil {
		util.LogAndPanic(err)
	}
	log.Println(string(jsonDirections))

	bytes, err = ioutil.ReadFile("resources/route_matrix_request_body.json")
	if err != nil {
		util.LogAndPanic(err)
	}

	postRouteMatrixBody := route.PostRouteMatrixRequestBody{}
	err = json.Unmarshal(bytes, &postRouteMatrixBody)
	if err != nil {
		util.LogAndPanic(err)
	}
	postMatrixResp, err := routeClient.BeginPostRouteMatrix(ctx, route.ResponseFormatJSON, postRouteMatrixBody, &route.RouteBeginPostRouteMatrixOptions{
		// make sure to check options available
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	matrixResp, err := postMatrixResp.PollUntilDone(ctx, 1*time.Second)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("posted route matrix")
	jsonMatrixResp, err := matrixResp.RouteMatrixResponse.MarshalJSON()
	if err != nil {
		util.LogAndPanic(err)
	}
	log.Println(string(jsonMatrixResp))

	// Output:
	// retrieved route directions
	// retrieved route range
	// posted route direction
	// posted route direction batch
	// posted route matrix
}
