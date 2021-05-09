package main

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v4.0/qnamaker"
	"github.com/Azure/go-autorest/autorest"
	"log"
	"time"
)

// Replace this with a valid subscription key.
var subscription_key string = "INSERT KEY HERE"

// Replace this with the endpoint for your subscription key.
var endpoint string = "https://westus.api.cognitive.microsoft.com"

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
Error
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/models.go#L174
ErrorResponse
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/models.go#L189
ErrorResponseError
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/models.go#L196
InnerErrorModel
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/models.go#L218
*/
func print_inner_error (error qnamaker.InnerErrorModel) {
	if error.Code != nil {
		fmt.Println (*error.Code)
	}
	if error.InnerError != nil {
		print_inner_error (*error.InnerError)
	}
}

func print_error_details (errors []qnamaker.Error) {
	for _, err := range errors {
		if err.Message != nil {
			fmt.Println (*err.Message)
		}
		if err.Details != nil {
			print_error_details (*err.Details)
		}
		if err.InnerError != nil {
			print_inner_error (*err.InnerError)
		}
	}
}

func handle_error (result qnamaker.Operation) {
	if result.ErrorResponse != nil {
		response := *result.ErrorResponse
		if response.Error != nil {
			err := *response.Error
			if err.Message != nil {
				fmt.Println (*err.Message)
			}
			if err.Details != nil {
				print_error_details (*err.Details)
			}
			if err.InnerError != nil {
				print_inner_error (*err.InnerError)
			}
		}
	}
}

/* See:
Create
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/knowledgebase.go#L39
CreateKbDTO
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/models.go#L131
MetadataDTO
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/models.go#L258
QnADTO
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/models.go#L321
FileDTO
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/models.go#L210
Operation
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/models.go#L266
*/
func add_kb (ctx context.Context, kb_client qnamaker.KnowledgebaseClient, ops_client qnamaker.OperationsClient) {
	name := "QnA Maker FAQ"

	/*
	The fields of QnADTO are pointers, and we cannot get the addresses of literal values,
	so we declare helper variables.
	*/
	id := int32(0)
	answer := "You can use our REST APIs to manage your Knowledge Base. See here for details: https://westus.dev.cognitive.microsoft.com/docs/services/58994a073d9e04097c7ba6fe/operations/58994a073d9e041ad42d9baa"
	source := "Custom Editorial"
	questions := []string{ "How do I programmatically update my Knowledge Base?" }

	// The fields of MetadataDTO are also pointers.
	metadata_name_1 := "category"
	metadata_value_1 := "api"
	metadata := []qnamaker.MetadataDTO{ qnamaker.MetadataDTO{ Name: &metadata_name_1, Value: &metadata_value_1 } }
	qna_list := []qnamaker.QnADTO{ qnamaker.QnADTO{
		ID: &id,
		Answer: &answer,
		Source: &source,
		Questions: &questions,
		Metadata: &metadata,
	} }

	urls := []string{}
	files := []qnamaker.FileDTO{}

	// The fields of CreateKbDTO are all pointers, so we get the addresses of our variables.
	createKbPayload := qnamaker.CreateKbDTO{ Name: &name, QnaList: &qna_list, Urls: &urls, Files: &files }

	// Create the KB.
	kb_result, kb_err := kb_client.Create (ctx, createKbPayload)
	if kb_err != nil {
		log.Fatal(kb_err)
	}

	// Wait for the KB create operation to finish.
	fmt.Println ("Waiting for KB create operation to finish...")
	// Operation.OperationID is a pointer, so we need to dereference it.
	operation_id := *kb_result.OperationID
	done := false
	for done == false {
		op_result, op_err := ops_client.GetDetails (ctx, operation_id)
		if op_err != nil {
			log.Fatal(op_err)
		}
		// If the operation isn't finished, wait and query again.
		if op_result.OperationState == "Running" || op_result.OperationState == "NotStarted" {
			fmt.Println ("Operation is not finished. Waiting 10 seconds...")
			time.Sleep (time.Duration(10) * time.Second)
		} else {
			done = true
			fmt.Print ("Operation result: " + op_result.OperationState)
			fmt.Println ()
			if op_result.OperationState == "Failed" {
				handle_error (op_result)
			}
		}
	}
}

// Delete the specified KB. You can use this method to delete excess KBs created with this quickstart.
func delete_kb (ctx context.Context, kb_client qnamaker.KnowledgebaseClient, kb_id string) {
	// Delete the KB.
	_, kb_err := kb_client.Delete (ctx, kb_id)
	if kb_err != nil {
		log.Fatal(kb_err)
	}
	fmt.Println ("KB " + kb_id + " deleted.")
}

/* See:
NewKnowledgebaseClient
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/knowledgebase.go#L34
NewOperationsClient
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/operations.go#L33
*/
func main() {
	// Get the context, which is required by the SDK methods.
	ctx := context.Background()

	kb_client := qnamaker.NewKnowledgebaseClient(endpoint)
	// Set the subscription key on the client.
	kb_client.Authorizer = autorest.NewCognitiveServicesAuthorizer(subscription_key)

	// We use this to check on the status of the create KB request.
	ops_client := qnamaker.NewOperationsClient(endpoint)
	// Set the subscription key on the client.
	ops_client.Authorizer = autorest.NewCognitiveServicesAuthorizer(subscription_key)

	add_kb (ctx, kb_client, ops_client)
	list_kbs (ctx, kb_client)
}
