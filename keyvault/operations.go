package keyvault

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/Azure/go-autorest/autorest/to"
)

var (
	keyName = "az-samples-go-" + internal.GetRandomLetterSequence(10)
)

func getKeysClient() keyvault.BaseClient {
	token, _ := iam.GetKeyvaultToken(iam.AuthGrantType())
	vmClient := keyvault.New()
	vmClient.Authorizer = token
	vmClient.AddToUserAgent(internal.UserAgent())
	return vmClient
}

// CreateKeyBundle creates a key in the specified keyvault
func CreateKeyBundle(ctx context.Context, vaultName string) (key keyvault.KeyBundle, err error) {
	vaultsClient := getVaultsClient()
	vault, err := vaultsClient.Get(ctx, internal.ResourceGroupName(), vaultName)
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
