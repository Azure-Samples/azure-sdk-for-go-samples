package maps

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/1.0/timezone"
)

func Example_timezoneOperations() {
	ctx := context.Background()
	accountsClient := getAccountsClient()
	cred, credErr := Authenticate(&accountsClient, ctx, *mapsAccount.Name, *usesADAuth)
	if credErr != nil {
		util.LogAndPanic(credErr)
	}

	conn := timezone.NewConnection(timezone.GeographyUs.ToPtr(), cred, nil)
	// xmsClientId doesn't need to be supplied for SharedKey auth
	var xmsClientId *string
	if *usesADAuth {
		xmsClientId = mapsAccount.Properties.UniqueID
	}

	timezoneClient := timezone.NewTimezoneClient(conn, xmsClientId)
	tzByCoordResp, err := timezoneClient.GetTimezoneByCoordinates(ctx, timezone.ResponseFormatJSON, "47.0,-122", &timezone.TimezoneGetTimezoneByCoordinatesOptions{
		Options: timezone.TimezoneOptionsAll.ToPtr(),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved timezone by coordinates")
	util.PrintAndLog(fmt.Sprintf("timezone alias: %s", *tzByCoordResp.TimezoneByCoordinatesResult.TimeZones[0].Aliases[0]))
	jsonResp, jsonErr := tzByCoordResp.TimezoneByCoordinatesResult.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	tzIANAVersionResp, err := timezoneClient.GetTimezoneIANAVersion(ctx, timezone.ResponseFormatJSON, nil)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved timezone IANA version")
	log.Println(*tzIANAVersionResp.TimezoneIanaVersionResult.Version)

	tzByIdResp, err := timezoneClient.GetTimezoneByID(ctx, timezone.ResponseFormatJSON, "Asia/Bahrain", &timezone.TimezoneGetTimezoneByIDOptions{
		Options: timezone.TimezoneOptionsAll.ToPtr(),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved timezone by id")
	util.PrintAndLog(fmt.Sprintf("timezone: %s", *tzByIdResp.TimezoneByIDResult.TimeZones[0].Names.Daylight))
	jsonResp, jsonErr = tzByIdResp.TimezoneByIDResult.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	timezoneWindowsResp, err := timezoneClient.GetTimezoneEnumWindows(ctx, timezone.ResponseFormatJSON, nil)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved timezone windows")
	jsonResp, jsonErr = json.Marshal(timezoneWindowsResp.TimezoneEnumWindowArray)
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	timezoneWindowsIANAResp, err := timezoneClient.GetTimezoneWindowsToIANA(ctx, timezone.ResponseFormatJSON, "Eastern Standard Time", &timezone.TimezoneGetTimezoneWindowsToIANAOptions{})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved timezone windows IANA ids")
	jsonResp, jsonErr = json.Marshal(timezoneWindowsIANAResp.IanaIDArray)
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	// Output:
	// retrieved timezone by coordinates
	// timezone alias: US/Pacific
	// retrieved timezone IANA version
	// retrieved timezone by id
	// timezone: Arabian Daylight Time
	// retrieved timezone windows
	// retrieved timezone windows IANA ids
}
