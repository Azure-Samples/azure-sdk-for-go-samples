package maps

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/1.0/geolocation"
)

func Example_geolocationOperations() {
	ctx := context.Background()
	accountsClient := getAccountsClient()
	cred, credErr := Authenticate(&accountsClient, ctx, *mapsAccount.Name, *usesADAuth)
	if credErr != nil {
		util.LogAndPanic(credErr)
	}

	conn := geolocation.NewConnection(geolocation.GeographyUs.ToPtr(), cred, nil)
	// xmsClientId doesn't need to be supplied for SharedKey auth
	var xmsClientId *string
	if *usesADAuth {
		xmsClientId = mapsAccount.Properties.XMsClientID
	}

	geolocationClient := geolocation.NewGeolocationClient(conn, xmsClientId)
	getIpToLocationResp, err := geolocationClient.GetIPToLocationPreview(ctx, geolocation.ResponseFormatJSON, "140.113.0.0", nil)
	if err != nil {
		util.LogAndPanic(credErr)
	}
	util.PrintAndLog("fetched country/region")
	util.PrintAndLog(fmt.Sprintf("ISO: %s", *getIpToLocationResp.IPAddressToLocationResult.CountryRegion.IsoCode))

	// Output:
	// fetched country/region
	// ISO: TW
}
