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

// The ID of the KB to update.
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
Update
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/knowledgebase.go#L552
UpdateKbOperationDTO
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/models.go#L374
UpdateKbOperationDTOAdd
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/models.go#L384
UpdateKbOperationDTOUpdate 
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/models.go#L402
UpdateKbOperationDTODelete
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/models.go#L394
Operation
https://github.com/Azure/azure-sdk-for-go/blob/master/services/cognitiveservices/v4.0/qnamaker/models.go#L266
*/
func update_kb (ctx context.Context, kb_client qnamaker.KnowledgebaseClient, ops_client qnamaker.OperationsClient) {
	// Add new Q&A lists, URLs, and files to the KB.
	/*
	The fields of QnADTO are pointers, and we cannot get the addresses of literal values,
	so we declare helper variables.
	*/
	id := int32(1)
	answer := "You can change the default message if you use the QnAMakerDialog. See this for details: https://docs.botframework.com/en-us/azure-bot-service/templates/qnamaker/#navtitle"
	source := "Custom Editorial"
	questions := []string{ "How can I change the default message from QnA Maker?" }

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

	/*
	The fields of UpdateKbOperationDTOAdd, updateKBUpdatePayload, updateKBDeletePayload,
	and UpdateKbOperationDTO are all pointers, so we get the addresses of our variables.
	*/
	updateKBAddPayload := qnamaker.UpdateKbOperationDTOAdd{ QnaList: &qna_list, Urls: &urls, Files: &files }

	// Update the KB name.
	name := "New KB name"
	updateKBUpdatePayload := qnamaker.UpdateKbOperationDTOUpdate { Name: &name }

	// Delete the QnaList with ID 0.
	ids := []int32{ 0 }
	updateKBDeletePayload := qnamaker.UpdateKbOperationDTODelete { Ids: &ids }

	// Bundle the add, update, and delete requests.
	updateKbPayload := qnamaker.UpdateKbOperationDTO{ Add: &updateKBAddPayload, Update: &updateKBUpdatePayload, Delete: &updateKBDeletePayload }

	// Update the KB.
	kb_result, kb_err := kb_client.Update (ctx, kb_id, updateKbPayload)
	if kb_err != nil {
		log.Fatal(kb_err)
	}

	// Wait for the KB update operation to finish.
	fmt.Println ("Waiting for KB update operation to finish...")
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

	// We use this to check on the status of the update KB request.
	ops_client := qnamaker.NewOperationsClient(endpoint)
	// Set the subscription key on the client.
	ops_client.Authorizer = autorest.NewCognitiveServicesAuthorizer(subscription_key)

	// You can use this method to get the IDs of existing KBs that you can update.
	list_kbs (ctx, kb_client)
	update_kb (ctx, kb_client, ops_client)
}
