package maps

import (
	"context"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/1.0/weather"
	"github.com/Azure/go-autorest/autorest/to"
)

func Example_weatherOperations() {
	ctx := context.Background()
	accountsClient := getAccountsClient()
	cred, credErr := Authenticate(&accountsClient, ctx, *mapsAccount.Name, *usesADAuth)
	if credErr != nil {
		util.LogAndPanic(credErr)
	}

	conn := weather.NewConnection(weather.GeographyUs.ToPtr(), cred, nil)
	// xmsClientId doesn't need to be supplied for SharedKey auth
	var xmsClientId *string
	if *usesADAuth {
		xmsClientId = mapsAccount.Properties.UniqueID
	}

	weatherClient := weather.NewWeatherClient(conn, xmsClientId)
	currentCondResp, err := weatherClient.GetCurrentConditions(ctx, weather.ResponseFormatJSON, "47.641268,-122.125679", &weather.WeatherGetCurrentConditionsOptions{
		Details:  to.StringPtr("true"),
		Duration: to.Int32Ptr(0),
		Language: to.StringPtr("EN"),
		Unit:     weather.WeatherDataUnitMetric.ToPtr(),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved current weather conditions")
	jsonResp, jsonErr := currentCondResp.CurrentConditionsResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	dailyForecastResp, err := weatherClient.GetDailyForecast(ctx, weather.ResponseFormatJSON, "62.6490341,30.0734812", &weather.WeatherGetDailyForecastOptions{
		Duration: to.Int32Ptr(1),
		Language: to.StringPtr("EN"),
		Unit:     weather.WeatherDataUnitMetric.ToPtr(),
	})
	util.PrintAndLog("retrieved daily weather forecast")
	jsonResp, jsonErr = dailyForecastResp.DailyForecastResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	dailyIndicesResp, err := weatherClient.GetDailyIndices(ctx, weather.ResponseFormatJSON, "43.84745,-79.37849", &weather.WeatherGetDailyIndicesOptions{
		Duration:     to.Int32Ptr(1),
		IndexGroupID: to.Int32Ptr(11),
		Language:     to.StringPtr("EN"),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved daily indices")
	jsonResp, jsonErr = dailyIndicesResp.DailyIndicesResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	hourlyForecastResp, err := weatherClient.GetHourlyForecast(ctx, weather.ResponseFormatJSON, "47.632346,-122.138874", &weather.WeatherGetHourlyForecastOptions{
		Duration: to.Int32Ptr(1),
		Language: to.StringPtr("EN"),
		Unit:     weather.WeatherDataUnitMetric.ToPtr(),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved hourly forecast")
	jsonResp, jsonErr = hourlyForecastResp.HourlyForecastResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	minForecastResp, err := weatherClient.GetMinuteForecast(ctx, weather.ResponseFormatJSON, "47.632346,-122.138874", &weather.WeatherGetMinuteForecastOptions{
		Interval: to.Int32Ptr(15),
		Language: to.StringPtr("EN"),
	})
	util.PrintAndLog("retrieved minute forecast")
	jsonResp, jsonErr = minForecastResp.MinuteForecastResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	quarterDayForecast, err := weatherClient.GetQuarterDayForecast(ctx, weather.ResponseFormatJSON, "47.632346,-122.138874", &weather.WeatherGetQuarterDayForecastOptions{
		Duration: to.Int32Ptr(1),
		Language: to.StringPtr("EN"),
		Unit:     weather.WeatherDataUnitMetric.ToPtr(),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved quarter day forecast")
	jsonResp, jsonErr = quarterDayForecast.QuarterDayForecastResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	alertsResp, err := weatherClient.GetSevereWeatherAlerts(ctx, weather.ResponseFormatJSON, "48.057,-81.091", &weather.WeatherGetSevereWeatherAlertsOptions{
		Details:  to.StringPtr("true"),
		Language: to.StringPtr("EN"),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved severe weather alerts")
	jsonResp, jsonErr = alertsResp.SevereWeatherAlertsResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	weatherAlongRouteResp, err := weatherClient.GetWeatherAlongRoute(ctx, weather.ResponseFormatJSON, "38.907,-77.037,0:38.907,-77.009,10:38.926,-76.928,20:39.033,-76.852,30:39.168,-76.732,40:39.269,-76.634,50:39.287,-76.612,60", &weather.WeatherGetWeatherAlongRouteOptions{
		Language: to.StringPtr("EN"),
	})
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("retrieved weather along route")
	jsonResp, jsonErr = weatherAlongRouteResp.WeatherAlongRouteResponse.MarshalJSON()
	if jsonErr != nil {
		util.LogAndPanic(jsonErr)
	}
	log.Println(string(jsonResp))

	// Output:
	// retrieved current weather conditions
	// retrieved daily weather forecast
	// retrieved daily indices
	// retrieved hourly forecast
	// retrieved minute forecast
	// retrieved quarter day forecast
	// retrieved severe weather alerts
	// retrieved weather along route
}
