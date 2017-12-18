package autorest

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2017-06-01/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

var accountName = "testname01"

func requestInspect() autorest.PrepareDecorator {
	return func(p autorest.Preparer) autorest.Preparer {
		return autorest.PreparerFunc(func(r *http.Request) (*http.Request, error) {
			log.Printf("Inspecting Request: %s %s\n", r.Method, r.URL)
			return p.Prepare(r)
		})
	}
}

func respondInspect() autorest.RespondDecorator {
	return func(r autorest.Responder) autorest.Responder {
		return autorest.ResponderFunc(func(resp *http.Response) error {
			log.Printf("Inspecting Response: %s for %s %s\n", resp.Status, resp.Request.Method, resp.Request.URL)
			return r.Respond(resp)
		})
	}
}

// ExampleHookPipeline demonstrates how to add hooks to the request/response pipeline
func ExampleHookPipeline() {
	helpers.ParseArgs()
	var logfile = "./output.log"
	f, err := os.Create(logfile)
	if err != nil {
		log.Fatalf("failed to create file: %s", err)
	}
	defer os.Remove(logfile)
	log.SetOutput(f)

	token, _ := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	ac := storage.NewAccountsClient(helpers.SubscriptionID())
	ac.Authorizer = autorest.NewBearerAuthorizer(token)

	ac.Sender = autorest.CreateSender(
		autorest.WithLogging(log.New(f, "example: ", log.LstdFlags)),
	)
	ac.RequestInspector = requestInspect()
	ac.ResponseInspector = respondInspect()

	nameAvailabilityResponse, err := ac.CheckNameAvailability(
		storage.AccountCheckNameAvailabilityParameters{
			Name: to.StringPtr(accountName),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts"),
		},
	)

	if err != nil {
		log.Fatalf("Error: %v", err)
		return
	}

	if to.Bool(nameAvailabilityResponse.NameAvailable) {
		fmt.Printf("The storage account name '%s' is available\n", accountName)
	} else {
		fmt.Printf("The storage account name '%s' is unavailable because %s\n",
			accountName, to.String(nameAvailabilityResponse.Message),
		)
	}

	// Output:
	// The storage account name 'testname01' is available
}
