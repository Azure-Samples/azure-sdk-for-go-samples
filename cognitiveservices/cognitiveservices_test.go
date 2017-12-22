package cognitiveservices

import (
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/cognitiveservices/mgmt/cognitiveservices"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/entitysearch"
	"github.com/subosito/gotenv"
)

var (
	accountName = "azuresamplesgo" + helpers.GetRandomLetterSequence(10)
)

func init() {
	gotenv.Load()
	err := helpers.ParseArgs()
	if err != nil {
		log.Fatalln("failed to parse args")
	}
}

// Example creates a resource group and a Cognitive Services account of type Search. Then it executes searches
// for web pages, images, videos, news and entities
func ExampleCognitiveServicesSearch() {
	defer resources.Cleanup()

	_, err := resources.CreateGroup(helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("resource group created")

	_, err = CreateCSAccount(accountName, cognitiveservices.BingSearchv7)

	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	helpers.PrintAndLog("cognitive services search resource created")

	searchWeb()
	searchImages()
	searchVideos()
	searchNews()
	searchEntities()

	// Output:
	// resource group created
	// cognitive services search resource created
	// completed web search and got results
	// completed image search and got results
	// completed image search and got pivot suggestions
	// completed image search and got suggestions
	// completed image search and got query expansions
	// completed video search and got results
	// completed trending video search and got results
	// completed news search and got results
	// completed entity search and got results
}

func ExampleCognitiveServicesSpellCheck() {
	defer resources.Cleanup()

	_, err := resources.CreateGroup(helpers.ResourceGroupName())
	if err != nil {
		helpers.PrintAndLog(err.Error())
	}
	helpers.PrintAndLog("resource group created")

	_, err = CreateCSAccount(accountName, cognitiveservices.BingSpellCheckv7)

	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	helpers.PrintAndLog("cognitive services spellcheck resource created")

	spellCheckResult, err := SpellCheck(accountName)

	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	if len(*spellCheckResult.FlaggedTokens) > 0 {
		helpers.PrintAndLog("completed spell check and found corrections")

		firstFlaggedToken := (*spellCheckResult.FlaggedTokens)[0]
		log.Printf("Number of flagged tokens in the input: %v \n", len(*spellCheckResult.FlaggedTokens))
		log.Printf("First flagged token: %v \n", *firstFlaggedToken.Token)
		log.Printf("First flagged token error type: %v \n", firstFlaggedToken.Type)
		log.Printf("First flagged token suggestions count: %v \n", len(*firstFlaggedToken.Suggestions))
	}
	// Output:
	// resource group created
	// cognitive services spellcheck resource created
	// completed spell check and found corrections
}

func searchWeb() {
	webPages, err := SearchWeb(accountName)

	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	if len(*webPages.Value) > 0 {
		helpers.PrintAndLog("completed web search and got results")

		firstWebPage := (*webPages.Value)[0]
		log.Printf("Number of web results: %v \n", len(*webPages.Value))
		log.Printf("First web page name: %v \n", *firstWebPage.Name)
		log.Printf("First web page url: %v \n", *firstWebPage.URL)
	}
}

func searchImages() {
	images, err := SearchImages(accountName)

	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	if len(*images.Value) > 0 {
		helpers.PrintAndLog("completed image search and got results")

		firstImage := (*images.Value)[0]
		log.Printf("Number of image results: %v \n", len(*images.Value))
		log.Printf("First image token: %v \n", *firstImage.ImageInsightsToken)
		log.Printf("First image thumbnail url: %v \n", *firstImage.ThumbnailURL)
		log.Printf("First image content url: %v \n", *firstImage.ContentURL)
	}

	if len(*images.PivotSuggestions) > 0 {
		helpers.PrintAndLog("completed image search and got pivot suggestions")

		firstPivot := (*images.PivotSuggestions)[0]
		log.Printf("Number of pivot suggestions results: %v \n", len(*images.PivotSuggestions))

		if len(*firstPivot.Suggestions) > 0 {
			helpers.PrintAndLog("completed image search and got suggestions")

			firstSuggestion := (*firstPivot.Suggestions)[0]
			log.Printf("Number of suggestions on first pivot: %v \n", len(*firstPivot.Suggestions))
			log.Printf("First suggestion text: %v \n", *firstSuggestion.Text)
			log.Printf("First suggestion web search url: %v \n", *firstSuggestion.WebSearchURL)
		}
	}

	if len(*images.QueryExpansions) > 0 {
		helpers.PrintAndLog("completed image search and got query expansions")

		firstQE := (*images.QueryExpansions)[0]
		log.Printf("Number of query expansions : %v \n", len(*images.QueryExpansions))
		log.Printf("First query expansion text : %v \n", *firstQE.Text)
		log.Printf("First query expansion search link: %v \n", *firstQE.SearchLink)
	}
}

func searchVideos() {
	videos, err := SearchVideos(accountName)

	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	if len(*videos.Value) > 0 {
		helpers.PrintAndLog("completed video search and got results")

		firstVideo := (*videos.Value)[0]
		log.Printf("Number of video results: %v \n", len(*videos.Value))
		log.Printf("First video id: %v \n", *firstVideo.VideoID)
		log.Printf("First video name: %v \n", *firstVideo.Name)
		log.Printf("First video url: %v \n", *firstVideo.ContentURL)
	}

	trendingVideos, err := TrendingVideos(accountName)

	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	if len(*trendingVideos.BannerTiles) > 0 {
		helpers.PrintAndLog("completed trending video search and got results")

		firstBannerTitle := (*trendingVideos.BannerTiles)[0]
		log.Printf("Number of trending titles : %v \n", len(*trendingVideos.BannerTiles))
		log.Printf("First banner title text: %v \n", *firstBannerTitle.Query.Text)
		log.Printf("First banner title url: %v \n", *firstBannerTitle.Query.WebSearchURL)
	}
}

func searchNews() {
	news, err := SearchNews(accountName)

	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	if len(*news.Value) > 0 {
		helpers.PrintAndLog("completed news search and got results")

		firstNewsResult := (*news.Value)[0]

		log.Printf("Number of news results: %v \n", len(*news.Value))
		log.Printf("First news name: %v \n", *firstNewsResult.Name)
		log.Printf("First news url: %v \n", *firstNewsResult.URL)
		log.Printf("First news description: %v \n", *firstNewsResult.Description)
		log.Printf("First news publish date: %v \n", *firstNewsResult.DatePublished)

		org, success := (*firstNewsResult.Provider)[0].AsOrganization()

		if success != true {
			helpers.PrintAndLog(err.Error())
		}

		log.Printf("First news provider: %v \n", *org.Name)
	}
}

func searchEntities() {
	entities, err := SearchEntities(accountName)

	if err != nil {
		helpers.PrintAndLog(err.Error())
	}

	if len(*entities.Value) > 0 {
		helpers.PrintAndLog("completed entity search and got results")

		dominantEntity := filter(*entities.Value, filterFunc)
		firstEntity, _ := dominantEntity[0].AsThing()

		log.Printf("Number of entities: %v \n", len(*entities.Value))
		log.Printf("First dominant entity description: %v \n", *firstEntity.Description)
	}
}

func filterFunc(entity entitysearch.BasicThing) bool {
	thingEntity, _ := entity.AsThing()

	if thingEntity.EntityPresentationInfo.EntityScenario == entitysearch.EntityScenarioDominantEntity {
		return true
	}
	return false
}

func filter(vs []entitysearch.BasicThing, f func(entitysearch.BasicThing) bool) []entitysearch.BasicThing {
	vsf := make([]entitysearch.BasicThing, 0)

	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}
