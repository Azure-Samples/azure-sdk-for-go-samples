package maps

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/1.0/traffic"
	"github.com/Azure/go-autorest/autorest/to"
)

func Example_trafficOperations() {
	ctx := context.Background()
	accountsClient := getAccountsClient()
	cred, credErr := Authenticate(&accountsClient, ctx, *mapsAccount.Name, *usesADAuth)
	if credErr != nil {
		util.LogAndPanic(credErr)
	}

	conn := traffic.NewConnection(traffic.GeographyUs.ToPtr(), cred, nil)
	// xmsClientId doesn't need to be supplied for SharedKey auth
	var xmsClientId *string
	if *usesADAuth {
		xmsClientId = mapsAccount.Properties.UniqueID
	}

	trafficClient := traffic.NewTrafficClient(conn, xmsClientId)
	trafficFlowSegmentResp, err := trafficClient.GetTrafficFlowSegment(ctx, traffic.TextFormatJSON, traffic.TrafficFlowSegmentStyleAbsolute, 10, "52.41072,4.84239", &traffic.TrafficGetTrafficFlowSegmentOptions{
		OpenLr:    to.BoolPtr(true),
		Thickness: to.Int32Ptr(10),
		Unit:      traffic.SpeedUnitKMPH.ToPtr(),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved traffic flow segment")
	jsonResp, jsonErr := json.Marshal(trafficFlowSegmentResp.TrafficFlowSegmentResult)
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	trafficFlowTileResp, err := trafficClient.GetTrafficFlowTile(ctx, traffic.TileFormatPNG, traffic.TrafficFlowTileStyleAbsolute, 12, 2044, 1360, &traffic.TrafficGetTrafficFlowTileOptions{
		Thickness: to.Int32Ptr(10),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	err = saveResponseIntoTempFile("tilecache/12/2044", "1360.png", &trafficFlowTileResp.RawResponse.Body)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved and saved traffic flow tile")

	incidentResp, err := trafficClient.GetTrafficIncidentDetail(ctx, traffic.TextFormatJSON, traffic.TrafficIncidentDetailStyleNight, "6841263.950712,511972.674418,6886056.049288,582676.925582", 11, "1335294634919", &traffic.TrafficGetTrafficIncidentDetailOptions{
		ExpandCluster:    to.BoolPtr(true),
		Language:         to.StringPtr("EN"),
		OriginalPosition: to.BoolPtr(true),
		Projection:       traffic.ProjectionStandardEPSG900913.ToPtr(),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved traffic incident detail")
	jsonResp, jsonErr = incidentResp.TrafficIncidentDetailResult.Tm.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	trafficIncidentTileResp, err := trafficClient.GetTrafficIncidentTile(ctx, traffic.TileFormatPNG, traffic.TrafficIncidentTileStyleNight, 10, 175, 408, &traffic.TrafficGetTrafficIncidentTileOptions{
		TrafficState: to.StringPtr("-1"),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	err = saveResponseIntoTempFile("tilecache/10/175", "408.png", &trafficIncidentTileResp.RawResponse.Body)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved and saved traffic incident tile")

	viewportBbox := "-939584.4813015489,-23954526.723651607,14675583.153020501,25043442.895825107"
	overviewBbox := "-939584.4813018347,-23954526.723651607,14675583.153020501,25043442.8958229083"
	trafficIncidentViewportResp, err := trafficClient.GetTrafficIncidentViewport(ctx, traffic.TextFormatJSON, viewportBbox, 2, overviewBbox, 2, &traffic.TrafficGetTrafficIncidentViewportOptions{
		Copyright: to.BoolPtr(true),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved traffic incident viewport")
	jsonResp, jsonErr = json.Marshal(trafficIncidentViewportResp.TrafficIncidentViewportResult.ViewpResp)
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	// Output:
	// retrieved traffic flow segment
	// retrieved and saved traffic flow tile
	// retrieved traffic incident detail
	// retrieved and saved traffic incident tile
	// retrieved traffic incident viewport
}
