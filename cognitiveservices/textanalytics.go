// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cognitiveservices

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v2.1/textanalytics"
	"github.com/Azure/go-autorest/autorest"
)

func GetTextAnalyticsClient(apiEndPoint string, apiKey string) textanalytics.BaseClient {
	textAnalyticsClient := textanalytics.New(apiEndPoint)
	subscriptionKey := apiKey
	textAnalyticsClient.Authorizer = autorest.NewCognitiveServicesAuthorizer(subscriptionKey)
	return textAnalyticsClient
}

// returns a pointer to the string value passed in.
func StringPointer(v string) *string {
	return &v
}

// returns a pointer to the bool value passed in.
func BoolPointer(v bool) *bool {
	return &v
}

func SentimentAnalysis(apiEndPoint string, apiKey string) {
	textAnalyticsclient := GetTextAnalyticsClient(apiEndPoint, apiKey)
	ctx := context.Background()
	inputDocuments := []textanalytics.MultiLanguageInput {
		textanalytics.MultiLanguageInput {
			Language: StringPointer("en"),
			ID:StringPointer("0"),
			Text:StringPointer("I had the best day of my life."),
		},
		textanalytics.MultiLanguageInput {
			Language: StringPointer("en"),
			ID:StringPointer("1"),
			Text:StringPointer("This was a waste of my time. The speaker put me to sleep."),
		},
		textanalytics.MultiLanguageInput {
			Language: StringPointer("es"),
			ID:StringPointer("2"),
			Text:StringPointer("No tengo dinero ni nada que dar..."),
		},
		textanalytics.MultiLanguageInput {
			Language: StringPointer("it"),
			ID:StringPointer("3"),
			Text:StringPointer("L'hotel veneziano era meraviglioso. È un bellissimo pezzo di architettura."),
		},
	}

	batchInput := textanalytics.MultiLanguageBatchInput{Documents:&inputDocuments}
	result, _ := textAnalyticsclient.Sentiment(ctx, BoolPointer(false), &batchInput)
	batchResult := textanalytics.SentimentBatchResult{}
	jsonString, _ := json.Marshal(result.Value)
	json.Unmarshal(jsonString, &batchResult)

	// Printing sentiment results
	for _,document := range *batchResult.Documents {
		fmt.Printf("Document ID: %s " , *document.ID)
		fmt.Printf("Sentiment Score: %f\n",*document.Score)
	}

	// Printing document errors
	fmt.Println("Document Errors")
	for _,error := range *batchResult.Errors {
		fmt.Printf("Document ID: %s Message : %s\n" ,*error.ID, *error.Message)
	}
}

func DetectLanguage(apiEndPoint string, apiKey string) {
	textAnalyticsclient := GetTextAnalyticsClient(apiEndPoint, apiKey)
	ctx := context.Background()
	inputDocuments := []textanalytics.LanguageInput {
		textanalytics.LanguageInput {
			ID:StringPointer("0"),
			Text:StringPointer("This is a document written in English."),
		},
		textanalytics.LanguageInput {
			ID:StringPointer("1"),
			Text:StringPointer("Este es un document escrito en Español."),
		},
		textanalytics.LanguageInput {
			ID:StringPointer("2"),
			Text:StringPointer("这是一个用中文写的文件"),
		},
	}

	batchInput := textanalytics.LanguageBatchInput{Documents:&inputDocuments}
	result, _ := textAnalyticsclient.DetectLanguage(ctx, BoolPointer(false), &batchInput)

	// Printing language detection results
	for _,document := range *result.Documents {
		fmt.Printf("Document ID: %s " , *document.ID)
		fmt.Printf("Detected Languages with Score: ")
		for _,language := range *document.DetectedLanguages{
			fmt.Printf("%s %f,",*language.Name, *language.Score)
		}
		fmt.Println()
	}

	// Printing document errors
	fmt.Println("Document Errors")
	for _,error := range *result.Errors {
		fmt.Printf("Document ID: %s Message : %s\n" ,*error.ID, *error.Message)
	}
}

func ExtractKeyPhrases(apiEndPoint string, apiKey string) {
	textAnalyticsclient := GetTextAnalyticsClient(apiEndPoint, apiKey)
	ctx := context.Background()
	inputDocuments := []textanalytics.MultiLanguageInput {
		textanalytics.MultiLanguageInput {
			Language: StringPointer("ja"),
			ID:StringPointer("0"),
			Text:StringPointer("猫は幸せ"),
		},
		textanalytics.MultiLanguageInput {
			Language: StringPointer("de"),
			ID:StringPointer("1"),
			Text:StringPointer("Fahrt nach Stuttgart und dann zum Hotel zu Fu."),
		},
		textanalytics.MultiLanguageInput {
			Language: StringPointer("en"),
			ID:StringPointer("2"),
			Text:StringPointer("My cat might need to see a veterinarian."),
		},
		textanalytics.MultiLanguageInput {
			Language: StringPointer("es"),
			ID:StringPointer("3"),
			Text:StringPointer("A mi me encanta el fútbol!"),
		},
	}

	batchInput := textanalytics.MultiLanguageBatchInput{Documents:&inputDocuments}
	result, _ := textAnalyticsclient.KeyPhrases(ctx, BoolPointer(false), &batchInput)

	// Printing extracted key phrases results
	for _,document := range *result.Documents {
		fmt.Printf("Document ID: %s\n" , *document.ID)
		fmt.Printf("\tExtracted Key Phrases:\n")
		for _,keyPhrase := range *document.KeyPhrases{
			fmt.Printf("\t\t%s\n",keyPhrase)
		}
		fmt.Println()
	}

	// Printing document errors
	fmt.Println("Document Errors")
	for _,error := range *result.Errors {
		fmt.Printf("Document ID: %s Message : %s\n" ,*error.ID, *error.Message)
	}
}

func ExtractEntities(apiEndPoint string, apiKey string) {
	textAnalyticsclient := GetTextAnalyticsClient(apiEndPoint, apiKey)
	ctx := context.Background()
	inputDocuments := []textanalytics.MultiLanguageInput {
		textanalytics.MultiLanguageInput {
			Language: StringPointer("en"),
			ID:StringPointer("0"),
			Text:StringPointer("Microsoft was founded by Bill Gates and Paul Allen on April 4, 1975, to develop and sell BASIC interpreters for the Altair 8800."),
		},
		textanalytics.MultiLanguageInput {
			Language: StringPointer("es"),
			ID:StringPointer("1"),
			Text:StringPointer("La sede principal de Microsoft se encuentra en la ciudad de Redmond, a 21 kilómetros de Seattle."),
		},
	}

	batchInput := textanalytics.MultiLanguageBatchInput{Documents:&inputDocuments}
	result, _ := textAnalyticsclient.Entities(ctx, BoolPointer(false), &batchInput)

	// Printing extracted entities results
	for _,document := range *result.Documents {
		fmt.Printf("Document ID: %s\n" , *document.ID)
		fmt.Printf("\tExtracted Entities:\n")
		for _,entity := range *document.Entities{
			fmt.Printf("\t\tName: %s\tType: %s",*entity.Name, *entity.Type)
			if entity.SubType != nil{
				fmt.Printf("\tSub-Type: %s\n", *entity.SubType)
			}
			fmt.Println()
			for _,match := range *entity.Matches{
				fmt.Printf("\t\t\tOffset: %v\tLength: %v\tScore: %f\n", *match.Offset, *match.Length, *match.EntityTypeScore)
			}
		}
		fmt.Println()
	}

	// Printing document errors
	fmt.Println("Document Errors")
	for _,error := range *result.Errors {
		fmt.Printf("Document ID: %s Message : %s\n" ,*error.ID, *error.Message)
	}
}
