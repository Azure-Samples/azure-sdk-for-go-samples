package keyvault

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/Azure/go-autorest/autorest/to"
)

var (
	keyName = "az-samples-go-" + helpers.GetRandomLetterSequence(10)
)

func getKeysClient() keyvault.BaseClient {
	keyClient := keyvault.New()
	auth, _ := iam.GetKeyvaultAuthorizer(iam.AuthGrantType())
	keyClient.Authorizer = auth
	keyClient.AddToUserAgent(helpers.UserAgent())
	return keyClient
}

// CreateKeyBundle creates a key in the specified keyvault
func CreateKeyBundle(ctx context.Context, vaultName string) (key keyvault.KeyBundle, err error) {
	vaultsClient := getVaultsClient()
	vault, err := vaultsClient.Get(ctx, helpers.ResourceGroupName(), vaultName)
	if err != nil {
		return
	}
	vaultURL := *vault.Properties.VaultURI

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
