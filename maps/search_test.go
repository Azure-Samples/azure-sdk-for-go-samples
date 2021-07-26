package maps

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/1.0/search"
	"github.com/Azure/go-autorest/autorest/to"
)

func Example_searchOperations() {
	ctx := context.Background()
	accountsClient := getAccountsClient()
	cred, credErr := Authenticate(&accountsClient, ctx, *mapsAccount.Name, *usesADAuth)
	if credErr != nil {
		util.LogAndPanic(credErr)
	}

	conn := search.NewConnection(search.GeographyUs.ToPtr(), cred, nil)
	// xmsClientId doesn't need to be supplied for SharedKey auth
	var xmsClientId *string
	if *usesADAuth {
		xmsClientId = mapsAccount.Properties.UniqueID
	}

	searchClient := search.NewSearchClient(conn, xmsClientId)
	searchAddrResp, err := searchClient.GetSearchAddress(ctx, search.TextFormatJSON, "15127 NE 24th Street, Redmond, WA 98052", &search.SearchGetSearchAddressOptions{
		// make sure to check options available
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved search address")
	jsonResp, jsonErr := searchAddrResp.SearchCommonResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	reverseAddrResp, err := searchClient.GetSearchAddressReverse(ctx, search.TextFormatJSON, "37.337,-121.89", &search.SearchGetSearchAddressReverseOptions{
		// make sure to check options available
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved reverse geocoding result")
	jsonResp, jsonErr = reverseAddrResp.SearchAddressReverseResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	reverseCrossStreetResp, err := searchClient.GetSearchAddressReverseCrossStreet(ctx, search.TextFormatJSON, "37.337,-121.89", &search.SearchGetSearchAddressReverseCrossStreetOptions{
		// make sure to check options available
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved reverse geocoding cross-street result")
	jsonResp, jsonErr = reverseCrossStreetResp.SearchAddressReverseCrossStreetResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	searchAddrStrResp, err := searchClient.GetSearchAddressStructured(ctx, search.TextFormatJSON, &search.SearchGetSearchAddressStructuredOptions{
		CountryCode:        to.StringPtr("US"),
		StreetNumber:       to.StringPtr("15127"),
		StreetName:         to.StringPtr("NE 24th Street"),
		Municipality:       to.StringPtr("Redmond"),
		CountrySubdivision: to.StringPtr("WA"),
		PostalCode:         to.StringPtr("98052"),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved search address structured")
	jsonResp, jsonErr = searchAddrStrResp.SearchCommonResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	searchFuzzyResp, err := searchClient.GetSearchFuzzy(ctx, search.TextFormatJSON, "Seattle", &search.SearchGetSearchFuzzyOptions{
		// make sure to check options available
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved search fuzzy results")
	jsonResp, jsonErr = searchFuzzyResp.SearchCommonResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	nearbySearchResp, err := searchClient.GetSearchNearby(ctx, search.TextFormatJSON, 40.706270, -74.011454, &search.SearchGetSearchNearbyOptions{
		Radius: to.Float32Ptr(8046),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved nearby results")
	jsonResp, jsonErr = nearbySearchResp.SearchCommonResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	searchPOIResp, err := searchClient.GetSearchPOI(ctx, search.TextFormatJSON, "juice bars", &search.SearchGetSearchPOIOptions{
		Limit:  to.Int32Ptr(5),
		Lat:    to.Float32Ptr(47.606038),
		Lon:    to.Float32Ptr(-122.333345),
		Radius: to.Float32Ptr(8046),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved searched POIs")
	jsonResp, jsonErr = searchPOIResp.SearchCommonResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	searchPOICategoryResp, err := searchClient.GetSearchPOICategory(ctx, search.TextFormatJSON, "atm", &search.SearchGetSearchPOICategoryOptions{
		Limit:  to.Int32Ptr(5),
		Lat:    to.Float32Ptr(47.606038),
		Lon:    to.Float32Ptr(-122.333345),
		Radius: to.Float32Ptr(8046),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrived search POI category")
	jsonResp, jsonErr = searchPOICategoryResp.SearchCommonResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	geomIds := make([]string, 0, len(searchAddrResp.SearchCommonResponse.Results))
	for _, result := range searchFuzzyResp.SearchCommonResponse.Results {
		geomIds = append(geomIds, *result.DataSources.Geometry.ID)
	}
	searchPolygonResp, err := searchClient.GetSearchPolygon(ctx, search.ResponseFormatJSON, geomIds, nil)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved search polygon")
	jsonResp, jsonErr = searchPolygonResp.SearchPolygonResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	bytes, err := ioutil.ReadFile("resources/search_address_batch_request_body.json")
	if err != nil {
		util.LogAndPanic(err)
	}
	batchRequestBody := search.BatchRequestBody{}
	err = json.Unmarshal(bytes, &batchRequestBody)
	if err != nil {
		util.LogAndPanic(err)
	}
	searchAddrBatchLRO, err := searchClient.BeginPostSearchAddressBatch(ctx, search.ResponseFormatJSON, batchRequestBody, nil)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("begin post search address batch")
	searchAddrBatchResp, err := searchAddrBatchLRO.PollUntilDone(ctx, 1*time.Second)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("post search address batch complete")
	jsonResp, jsonErr = searchAddrBatchResp.SearchAddressBatchResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	bytes, err = ioutil.ReadFile("resources/search_address_reverse_batch_request_body.json")
	if err != nil {
		util.LogAndPanic(err)
	}
	batchRequestBody = search.BatchRequestBody{}
	err = json.Unmarshal(bytes, &batchRequestBody)
	if err != nil {
		util.LogAndPanic(err)
	}
	searchAddrReverseBatchLRO, err := searchClient.BeginPostSearchAddressReverseBatch(ctx, search.ResponseFormatJSON, batchRequestBody, nil)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("begin post search address reverse batch")
	searchAddReverseBatchResp, err := searchAddrReverseBatchLRO.PollUntilDone(ctx, 1*time.Second)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("post search address reverse batch complete")
	jsonResp, jsonErr = searchAddReverseBatchResp.SearchAddressReverseBatchResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	bytes, err = ioutil.ReadFile("resources/search_fuzzy_batch_request_body.json")
	if err != nil {
		util.LogAndPanic(err)
	}
	batchRequestBody = search.BatchRequestBody{}
	err = json.Unmarshal(bytes, &batchRequestBody)
	if err != nil {
		util.LogAndPanic(err)
	}
	searchFuzzyBatchLRO, err := searchClient.BeginPostSearchFuzzyBatch(ctx, search.ResponseFormatJSON, batchRequestBody, nil)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("begin post search fuzzy batch")
	searchFuzzyBatchResp, err := searchFuzzyBatchLRO.PollUntilDone(ctx, 1*time.Second)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("post search fuzzy batch complete")
	jsonResp, jsonErr = searchFuzzyBatchResp.SearchFuzzyBatchResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	bytes, err = ioutil.ReadFile("resources/search_along_route_request_body.json")
	if err != nil {
		util.LogAndPanic(err)
	}
	searchAlongRouteBody := search.SearchAlongRouteRequestBody{}
	err = json.Unmarshal(bytes, &searchAlongRouteBody)
	postSearchResp, err := searchClient.PostSearchAlongRoute(ctx, search.TextFormatJSON, "burger", 1000, searchAlongRouteBody, &search.SearchPostSearchAlongRouteOptions{
		Limit: to.Int32Ptr(2),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved search along route results")
	jsonResp, jsonErr = postSearchResp.SearchCommonResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	bytes, err = ioutil.ReadFile("resources/search_inside_geometry_request_body.json")
	if err != nil {
		util.LogAndPanic(err)
	}
	searchInsideGeomBody := search.SearchInsideGeometryRequestBody{}
	err = json.Unmarshal(bytes, &searchInsideGeomBody)
	postSearchInsideGeomResp, err := searchClient.PostSearchInsideGeometry(ctx, search.TextFormatJSON, "burger", searchInsideGeomBody, &search.SearchPostSearchInsideGeometryOptions{
		Limit: to.Int32Ptr(2),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved search inside geometry response")
	jsonResp, jsonErr = postSearchInsideGeomResp.SearchCommonResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	// Output:
	// retrieved search address
	// retrieved reverse geocoding result
	// retrieved reverse geocoding cross-street result
	// retrieved search address structured
	// retrieved search fuzzy results
	// retrieved nearby results
	// retrieved searched POIs
	// retrived search POI category
	// retrieved search polygon
	// begin post search address batch
	// post search address batch complete
	// begin post search address reverse batch
	// post search address reverse batch complete
	// begin post search fuzzy batch
	// post search fuzzy batch complete
	// retrieved search along route results
	// retrieved search inside geometry response
}
