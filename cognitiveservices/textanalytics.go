// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cognitiveservices
// <imports>
import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"

    "github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v2.1/textanalytics"
    "github.com/Azure/go-autorest/autorest"
    "github.com/Azure/go-autorest/autorest/to"
)
// </imports>

// <client>
func GetTextAnalyticsClient() textanalytics.BaseClient {
    subscriptionKeyVar := "TEXT_ANALYTICS_SUBSCRIPTION_KEY"
    if "" == os.Getenv(subscriptionKeyVar) {
        log.Fatal("Please set/export the environment variable " + subscriptionKeyVar + ".")
    }
    subscriptionKey := os.Getenv(subscriptionKeyVar)
    endpointVar := "TEXT_ANALYTICS_ENDPOINT"
    if "" == os.Getenv(endpointVar) {
        log.Fatal("Please set/export the environment variable " + endpointVar + ".")
    }
    endpoint := os.Getenv(endpointVar)

    textAnalyticsClient := textanalytics.New(endpoint)
    textAnalyticsClient.Authorizer = autorest.NewCognitiveServicesAuthorizer(subscriptionKey)

    return textAnalyticsClient
}
// </client>

// detects the sentiment of a set of text records
// <sentimentAnalysis>
func SentimentAnalysis() {
    textAnalyticsClient := GetTextAnalyticsClient()
    ctx := context.Background()
    inputDocuments := []textanalytics.MultiLanguageInput{
        {
            Language: to.StringPtr("en"),
            ID:       to.StringPtr("0"),
            Text:     to.StringPtr("I had the best day of my life."),
        },
        {
            Language: to.StringPtr("en"),
            ID:       to.StringPtr("1"),
            Text:     to.StringPtr("This was a waste of my time. The speaker put me to sleep."),
        },
        {
            Language: to.StringPtr("es"),
            ID:       to.StringPtr("2"),
            Text:     to.StringPtr("No tengo dinero ni nada que dar..."),
        },
        {
            Language: to.StringPtr("it"),
            ID:       to.StringPtr("3"),
            Text:     to.StringPtr("L'hotel veneziano era meraviglioso. È un bellissimo pezzo di architettura."),
        },
    }

    batchInput := textanalytics.MultiLanguageBatchInput{Documents: &inputDocuments}
    result, _ := textAnalyticsClient.Sentiment(ctx, to.BoolPtr(false), &batchInput)
    var batchResult textanalytics.SentimentBatchResult
    jsonString, _ := json.Marshal(result)
    _ = json.Unmarshal(jsonString, &batchResult)

    // Printing sentiment results
    for _, document := range *batchResult.Documents {
        fmt.Printf("Document ID: %s ", *document.ID)
        fmt.Printf("Sentiment Score: %f\n", *document.Score)
    }

    // Printing document errors
    fmt.Println("Document Errors")
    for _, err := range *batchResult.Errors {
        fmt.Printf("Document ID: %s Message : %s\n", *err.ID, *err.Message)
    }
}
// </sentimentAnalysis>

//detects the language of a text document
// <languageDetection>
func DetectLanguage() {
    textAnalyticsClient := GetTextAnalyticsClient()
    ctx := context.Background()
    inputDocuments := []textanalytics.LanguageInput{
        {
            ID:   to.StringPtr("0"),
            Text: to.StringPtr("This is a document written in English."),
        },
        {
            ID:   to.StringPtr("1"),
            Text: to.StringPtr("Este es un document escrito en Español."),
        },
        {
            ID:   to.StringPtr("2"),
            Text: to.StringPtr("这是一个用中文写的文件"),
        },
    }

    batchInput := textanalytics.LanguageBatchInput{Documents: &inputDocuments}
    result, _ := textAnalyticsClient.DetectLanguage(ctx, to.BoolPtr(false), &batchInput)

    // Printing language detection results
    for _, document := range *result.Documents {
        fmt.Printf("Document ID: %s ", *document.ID)
        fmt.Printf("Detected Languages with Score: ")
        for _, language := range *document.DetectedLanguages {
            fmt.Printf("%s %f,", *language.Name, *language.Score)
        }
        fmt.Println()
    }

    // Printing document errors
    fmt.Println("Document Errors")
    for _, err := range *result.Errors {
        fmt.Printf("Document ID: %s Message : %s\n", *err.ID, *err.Message)
    }
}
// </languageDetection>

// extracts key-phrases from a text document
// <keyPhrases>
func ExtractKeyPhrases() {
    textAnalyticsClient := GetTextAnalyticsClient()
    ctx := context.Background()
    inputDocuments := []textanalytics.MultiLanguageInput{
        {
            Language: to.StringPtr("ja"),
            ID:       to.StringPtr("0"),
            Text:     to.StringPtr("猫は幸せ"),
        },
        {
            Language: to.StringPtr("de"),
            ID:       to.StringPtr("1"),
            Text:     to.StringPtr("Fahrt nach Stuttgart und dann zum Hotel zu Fu."),
        },
        {
            Language: to.StringPtr("en"),
            ID:       to.StringPtr("2"),
            Text:     to.StringPtr("My cat might need to see a veterinarian."),
        },
        {
            Language: to.StringPtr("es"),
            ID:       to.StringPtr("3"),
            Text:     to.StringPtr("A mi me encanta el fútbol!"),
        },
    }

    batchInput := textanalytics.MultiLanguageBatchInput{Documents: &inputDocuments}
    result, _ := textAnalyticsClient.KeyPhrases(ctx, to.BoolPtr(false), &batchInput)

    // Printing extracted key phrases results
    for _, document := range *result.Documents {
        fmt.Printf("Document ID: %s\n", *document.ID)
        fmt.Printf("\tExtracted Key Phrases:\n")
        for _, keyPhrase := range *document.KeyPhrases {
            fmt.Printf("\t\t%s\n", keyPhrase)
        }
        fmt.Println()
    }

    // Printing document errors
    fmt.Println("Document Errors")
    for _, err := range *result.Errors {
        fmt.Printf("Document ID: %s Message : %s\n", *err.ID, *err.Message)
    }
}
// </keyPhrases>

//  identifies well-known entities in a text document
// <entityRecognition>
func ExtractEntities() {
    textAnalyticsClient := GetTextAnalyticsClient()
    ctx := context.Background()
    inputDocuments := []textanalytics.MultiLanguageInput{
        {
            Language: to.StringPtr("en"),
            ID:       to.StringPtr("0"),
            Text:     to.StringPtr("Microsoft was founded by Bill Gates and Paul Allen on April 4, 1975, to develop and sell BASIC interpreters for the Altair 8800."),
        },
        {
            Language: to.StringPtr("es"),
            ID:       to.StringPtr("1"),
            Text:     to.StringPtr("La sede principal de Microsoft se encuentra en la ciudad de Redmond, a 21 kilómetros de Seattle."),
        },
    }

    batchInput := textanalytics.MultiLanguageBatchInput{Documents: &inputDocuments}
    result, _ := textAnalyticsClient.Entities(ctx, to.BoolPtr(false), &batchInput)

    // Printing extracted entities results
    for _, document := range *result.Documents {
        fmt.Printf("Document ID: %s\n", *document.ID)
        fmt.Printf("\tExtracted Entities:\n")
        for _, entity := range *document.Entities {
            fmt.Printf("\t\tName: %s\tType: %s", *entity.Name, *entity.Type)
            if entity.SubType != nil {
                fmt.Printf("\tSub-Type: %s\n", *entity.SubType)
            }
            fmt.Println()
            for _, match := range *entity.Matches {
                fmt.Printf("\t\t\tOffset: %v\tLength: %v\tScore: %f\n", *match.Offset, *match.Length, *match.EntityTypeScore)
            }
        }
        fmt.Println()
    }

    // Printing document errors
    fmt.Println("Document Errors")
    for _, err := range *result.Errors {
        fmt.Printf("Document ID: %s Message : %s\n", *err.ID, *err.Message)
    }
}
// </entityRecognition>