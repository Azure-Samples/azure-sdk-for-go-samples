package keyvault

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/Azure/go-autorest/autorest/to"
)

var (
	keyName = "az-samples-go-" + helpers.GetRandomLetterSequence(10)
)

func getKeysClient() keyvault.BaseClient {
	token, _ := iam.GetKeyvaultToken(iam.AuthGrantType())
	vmClient := keyvault.New()
	vmClient.Authorizer = token
	vmClient.AddToUserAgent(helpers.UserAgent())
	return vmClient
}

func CreateKeyBundle(ctx context.Context, vaultName string) (keyvault.KeyBundle, error) {
	// vaultName = "az-samples-go-zdjykYvTmF"
	vaultURL := fmt.Sprintf("https://%s.vault.azure.net/", vaultName)
	keyClient := getKeysClient()

	return keyClient.CreateKey(
		ctx,
		vaultURL,
		keyName,
		keyvault.KeyCreateParameters{
			KeyAttributes: &keyvault.KeyAttributes{
				Enabled: to.BoolPtr(true),
			},
			KeySize: to.Int32Ptr(2048), // As of writing this sample, 2048 is the only supported KeySize.
			KeyOps: &[]keyvault.JSONWebKeyOperation{
				keyvault.Encrypt,
				keyvault.Decrypt,
			},
			Kty: keyvault.RSA,
		})
}
