package main

// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.
//
//
// You need to set four environment variables before using the app:
// AZURE_TENANT_ID: Your Azure tenant ID
// AZURE_CLIENT_ID: Your Azure client ID. This will be an app ID from your AAD.
// KVAULT_NAME: The name of your vault (just the name, not the full URL/path)
//
// Optional command line argument:
// If you have a secret already, set KVAULT_SECRET_NAME to the secret's name.
//
// NOTE: Do NOT set AZURE_CLIENT_SECRET. This example uses Managed identities.
// The README.md provides more information.
//
//

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/keyvault"
	kvauth "github.com/Azure/azure-sdk-for-go/services/keyvault/auth"
	"github.com/Azure/go-autorest/autorest"
)

var (
	vaultName  string
	secretName string
)

func listSecrets(basicClient keyvault.BaseClient) {
	secretList, err := basicClient.GetSecrets(context.Background(), "https://"+vaultName+".vault.azure.net", nil)
	if err != nil {
		fmt.Printf("unable to get list of secrets: %v\n", err)
		os.Exit(1)
	}

	for ; secretList.NotDone(); secretList.NextWithContext(context.Background()) {
		// group by ContentType
		secWithType := make(map[string][]string)
		secWithoutType := make([]string, 1)
		for _, secret := range secretList.Values() {
			if secret.ContentType != nil {
				_, exists := secWithType[*secret.ContentType]
				if exists {
					secWithType[*secret.ContentType] = append(secWithType[*secret.ContentType], path.Base(*secret.ID))
				} else {
					tempSlice := make([]string, 1)
					tempSlice[0] = path.Base(*secret.ID)
					secWithType[*secret.ContentType] = tempSlice
				}
			} else {
				secWithoutType = append(secWithoutType, path.Base(*secret.ID))
			}
		}

		for k, v := range secWithType {
			fmt.Println(k)
			for _, sec := range v {
				fmt.Println(sec)
			}
		}
		for _, wov := range secWithoutType {
			fmt.Println(wov)
		}
	}
}

func getSecret(basicClient keyvault.BaseClient, secname string) {
	secretResp, err := basicClient.GetSecret(context.Background(), "https://"+vaultName+".vault.azure.net", secname, "")
	if err != nil {
		fmt.Printf("unable to get value for secret: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(*secretResp.Value)
}

func createUpdateSecret(basicClient keyvault.BaseClient, secname, secvalue string) {
	var secParams keyvault.SecretSetParameters
	secParams.Value = &secvalue
	newBundle, err := basicClient.SetSecret(context.Background(), "https://"+vaultName+".vault.azure.net", secname, secParams)
	if err != nil {
		fmt.Printf("unable to add/update secret: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("added/updated: " + *newBundle.ID)
}

func deleteSecret(basicClient keyvault.BaseClient, secname string) {
	_, err := basicClient.DeleteSecret(context.Background(), "https://"+vaultName+".vault.azure.net", secname)
	if err != nil {
		fmt.Printf("error deleting secret: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(secname + " deleted successfully")
}

func logRequest() autorest.PrepareDecorator {
	return func(p autorest.Preparer) autorest.Preparer {
		return autorest.PreparerFunc(func(r *http.Request) (*http.Request, error) {
			r, err := p.Prepare(r)
			if err != nil {
				log.Println(err)
			}
			dump, _ := httputil.DumpRequestOut(r, true)
			log.Println(string(dump))
			return r, err
		})
	}
}

func logResponse() autorest.RespondDecorator {
	return func(p autorest.Responder) autorest.Responder {
		return autorest.ResponderFunc(func(r *http.Response) error {
			err := p.Respond(r)
			if err != nil {
				log.Println(err)
			}
			dump, _ := httputil.DumpResponse(r, true)
			log.Println(string(dump))
			return err
		})
	}
}

func main() {
	vaultName = os.Getenv("KVAULT_NAME")
	fmt.Printf("KVAULT_NAME: %s\n", vaultName)

	authorizer, err := kvauth.NewAuthorizerFromEnvironment()
	if err != nil {
		fmt.Printf("unable to create vault authorizer: %v\n", err)
		os.Exit(1)
	}

	basicClient := keyvault.New()
	basicClient.Authorizer = authorizer

	fmt.Println("\nListing secret names in keyvault:")
	listSecrets(basicClient)

	if secretName = os.Getenv("KVAULT_SECRET_NAME"); secretName != "" {
		fmt.Printf("KVAULT_SECRET_NAME: %s\n", secretName)
		fmt.Print("KVAULT_SECRET Value: ")
		getSecret(basicClient, secretName)
	} else {
		fmt.Println("KVAULT_SECRET_NAME not set.\n")
	}

	fmt.Println("Setting 'newsecret' to 'newvalue'")
	createUpdateSecret(basicClient, "newsecret", "newvalue")
	fmt.Println("\nListing secret names in keyvault:")
	listSecrets(basicClient)
}
