AZURE_BASE_GROUP_NAME=az-samples-go
AZURE_LOCATION_DEFAULT=westus2
AZURE_SAMPLES_KEEP_RESOURCES=0

# create with:
# `az ad sp create-for-rbac --name 'my-sp' --output json`
# sp must have Contributor role on subscription
AZURE_TENANT_ID=
AZURE_CLIENT_ID=
AZURE_CLIENT_SECRET=
AZURE_SUBSCRIPTION_ID=

# create with:
# `az ad sp create-for-rbac --name 'my-sp' --sdk-auth > $HOME/.azure/sdk_auth.json`
# sp must have Contributor role on subscription
AZURE_AUTH_LOCATION=$HOME/.azure/sdk_auth.json

AZURE_STORAGE_ACCOUNT_NAME=
AZURE_STORAGE_ACCOUNT_GROUP_NAME=
