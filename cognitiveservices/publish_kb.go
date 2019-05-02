package main

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v4.0/qnamaker"
	"github.com/Azure/go-autorest/autorest"
	"log"
)

// Replace this with a valid subscription key.
var subscription_key string = "INSERT KEY HERE"

// Replace this with the endpoint for your subscription key.
var endpoint string = "https://westus.api.cognitive.microsoft.com"

// The ID of the KB to publish.
var kb_id string = "INSERT KB ID HERE"

/* See:
ListAll
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/knowledgebase.go#L335
KnowledgebasesDTO
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/models.go#L251
KnowledgebaseDTO
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/models.go#L229
*/
func list_kbs (ctx context.Context, client qnamaker.KnowledgebaseClient) {
	result, err := client.ListAll (ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println ("Existing knowledge bases:\n")
	// KnowledgebasesDTO.Knowledgebases is a pointer, so we need to dereference it.
	for _, item := range (*result.Knowledgebases) {
		// Most fields of KnowledgebaseDTO are pointers, so we need to dereference them.
        fmt.Println ("ID: " + *item.ID)
		fmt.Println ("Name: " + *item.Name)
		fmt.Println ()
    }
}

/* See:
Publish
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/knowledgebase.go#L401
*/
func publish_kb (ctx context.Context, kb_client qnamaker.KnowledgebaseClient, kb_id string) {
	// Publish the KB.
	_, kb_err := kb_client.Publish (ctx, kb_id)
	if kb_err != nil {
		log.Fatal(kb_err)
	}
	fmt.Println ("KB " + kb_id + " published.")
}

/* See:
NewKnowledgebaseClient
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/knowledgebase.go#L34
*/
func main() {
	// Get the context, which is required by the SDK methods.
	ctx := context.Background()

	kb_client := qnamaker.NewKnowledgebaseClient(endpoint)
	// Set the subscription key on the client.
	kb_client.Authorizer = autorest.NewCognitiveServicesAuthorizer(subscription_key)

	// You can use this method to get the IDs of existing KBs that you can publish.
	list_kbs (ctx, kb_client)
	publish_kb (ctx, kb_client, kb_id)
}
