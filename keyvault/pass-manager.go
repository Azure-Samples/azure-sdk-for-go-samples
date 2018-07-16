package keyvault

/*
 * You need to set four environment variables before using the app:
 * AZURE_TENANT_ID: Your Azure tenant ID
 * AZURE_CLIENT_ID: Your Azure client ID. This will be an app ID from your AAD.
 * AZURE_CLIENT_SECRET: The secret for the client ID above.
 * KVAULT: The name of your vault (just the name, not the full URL/path)
 *
 * Usage
 * List the secrets currently in the vault (not the values though):
 * kv-pass
 *
 * Get the value for a secret in the vault:
 * kv-pass YOUR_SECRETS_NAME
 *
 * Add or Update a secret in the vault:
 * kv-pass -edit YOUR_NEW_VALUE YOUR_SECRETS_NAME
 *
 * Delete a secret in the vault:
 * kv-pass -delete YOUR_SECRETS_NAME
 */

import (
	"context"
	"flag"
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
	secretVal = flag.String("edit", "", "set value of secret")
	secretDel = flag.String("delete", "", "delete secret")
	setDebug  = flag.Bool("debug", false, "debug")
	vaultName string
)

func main() {
	flag.Parse()

	if os.Getenv("AZURE_TENANT_ID") == "" || os.Getenv("AZURE_CLIENT_ID") == "" || os.Getenv("AZURE_CLIENT_SECRET") == "" || os.Getenv("KVAULT") == "" {
		fmt.Println("env vars not set, exiting...")
		os.Exit(1)
	}
	vaultName = os.Getenv("KVAULT")

	authorizer, err := kvauth.NewAuthorizerFromEnvironment()
	if err != nil {
		fmt.Printf("unable to create vault authorizer: %v\n", err)
		os.Exit(1)
	}

	basicClient := keyvault.New()
	basicClient.Authorizer = authorizer

	if *setDebug {
		basicClient.RequestInspector = logRequest()
		basicClient.ResponseInspector = logResponse()
	}

	allArgs := flag.Args()

	if flag.NArg() <= 0 && flag.NFlag() <= 0 {
		listSecrets(basicClient)
	}

	if flag.NArg() == 1 && flag.NFlag() <= 0 {
		getSecret(basicClient, allArgs[0])
	}

	if *secretVal != "" && *secretDel == "" {
		createUpdateSecret(basicClient, allArgs[0], *secretVal)
	}

	if *secretDel != "" {
		deleteSecret(basicClient, *secretDel)
	}
}

func listSecrets(basicClient keyvault.BaseClient) {
	secretList, err := basicClient.GetSecrets(context.Background(), "https://"+vaultName+".vault.azure.net", nil)
	if err != nil {
		fmt.Printf("unable to get list of secrets: %v\n", err)
		os.Exit(1)
	}

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
			fmt.Println(" |--- " + sec)
		}
	}
	for _, wov := range secWithoutType {
		fmt.Println(wov)
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
