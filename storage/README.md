---
services: storage
platforms: go
author: mcardosos,joshgav
---

# Manage Azure Storage

This package demonstrates how to manage storage accounts and blobs with the Azure SDK for Go. You can use the SDK to create and update storage accounts, list all accounts in a subscription or resource group, list and regenerate keys, and delete accounts. You can also use the SDK to create and delete containers and the blobs they contain.

If you need it, get an Azure trial [here](https://azure.microsoft.com/pricing/free-trial).

**On this page**

- [Run this sample](#run)
- [What does this sample do?](#sample)
    - [Check storage account name availability](#check)
    - [Create a new storage account](#createsa)
    - [Get the properties of a storage account](#get)
    - [List storage accounts by resource group](#listsarg)
    - [List storage accounts in subscription](#listsasyb)
    - [Get the storage account keys](#getkeys)
    - [Regenerate a storage account key](#regenkey)
    - [Update the storage account](#update)
    - [List usage](#listusage)
    - [Delete storage account](#delete)
- [More information](#info)

<a id="run"></a>

## Run this sample

1. If necessary [install Go](https://golang.org/dl/).
1. Clone this repository.

  ```bash
  export PROJECT=github.com/Azure-Samples/azure-sdk-for-go-samples/storage
  go get -u $PROJECT
  cd ${GOPATH}/src/${PROJECT}
  dep ensure
  ```
1. Create an Azure service principal either through
    [Azure CLI](https://azure.microsoft.com/documentation/articles/resource-group-authenticate-service-principal-cli/),
    [PowerShell](https://azure.microsoft.com/documentation/articles/resource-group-authenticate-service-principal/)
    or [the portal](https://azure.microsoft.com/documentation/articles/resource-group-create-service-principal-portal/).

1. Set the following environment variables based on the properties of this new service principal.

  |EnvVar | Value|
  |-------|------|
  |AZURE_TENANT_ID|your tenant id|
  |AZURE_SUBSCRIPTION_ID|your subscription ID|
  |AZURE_CLIENT_ID|service principal/application ID|
  |AZURE_CLIENT_SECRET|service principal/application secret|
  |AZURE_RG_NAME|name of new resource group|
  |AZURE_LOCATION|location for all resources|
  |AZURE_STORAGE_ACCOUNT_NAME|name for test storage account|
  |AZURE_STORAGE_ACCOUNT_GROUP_NAME|name for storage account group|

1. Run the sample.

```
go test
```

<a id="sample"></a>

## What does example.go do?

<a id="check"></a>

### Check storage account name availability

Check the validity and availability of a string as a storage account name.

```go
result, err := storageAccountsClient.CheckNameAvailability(
  storage.AccountCheckNameAvailabilityParameters{
    Name: to.StringPtr(accountName),
    Type: to.StringPtr("Microsoft.Storage/storageAccounts")})
if err != nil {
  log.Fatalf("%s: %v", "storage account creation failed", err)
}
if *result.NameAvailable != true {
  log.Fatalf("%s: %v", "storage account name not available", err)
}
```

<a id="createsa"></a>

### Create a new storage account

```go
// CreateStorageAccount creates a new storage account.
func CreateStorageAccount() (<-chan storage.Account, <-chan error) {
	storageAccountsClient, _ := getStorageAccountsClient()

	result, err := storageAccountsClient.CheckNameAvailability(
		storage.AccountCheckNameAvailabilityParameters{
			Name: to.StringPtr(accountName),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts")})
	if err != nil {
		log.Fatalf("%s: %v", "storage account creation failed", err)
	}
	if *result.NameAvailable != true {
		log.Fatalf("%s: %v", "storage account name not available", err)
	}

	return storageAccountsClient.Create(
		helpers.ResourceGroupName,
		accountName,
		storage.AccountCreateParameters{
			Sku: &storage.Sku{
				Name: storage.StandardLRS},
			Location: to.StringPtr(helpers.Location),
			AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{}},
		nil /* cancel <-chan struct{} */)
}
```

<a id="get"></a>

### Get the properties of a storage account

```go
account, err := storageClient.GetProperties(groupName, accountName)
```

<a id="listsarg"></a>

### List storage accounts by resource group

```go
listGroupAccounts, err := storageClient.ListByResourceGroup(groupName)
onErrorFail(err, "ListByResourceGroup failed")

for _, acc := range *listGroupAccounts.Value {
     fmt.Printf("\t%s\n", *acc.Name)
}
```

<a id="listsasub"></a>

### List storage accounts in subscription

```go
listSubAccounts, err := storageClient.List()
onErrorFail(err, "List failed")

for _, acc := range *listSubAccounts.Value {
    fmt.Printf("\t%s\n", *acc.Name)
}
```

<a id="getkeys"></a>

### Get the storage account keys

```go
keys, err := storageClient.ListKeys(groupName, accountName)
onErrorFail(err, "ListKeys failed")

fmt.Printf("'%s' storage account keys\n", accountName)
for _, key := range *keys.Keys {
    fmt.Printf("\tKey name: %s\n\tValue: %s...\n\tPermissions: %s\n",
        *key.KeyName,
        (*key.Value)[:5],
        key.Permissions)
    fmt.Println("\t----------------")
}
```

<a id="regenkey"></a>

### Regenerate a storage account key

```go
keys, err = storageClient.RegenerateKey(groupName, accountName, storage.AccountRegenerateKeyParameters{
    KeyName: (*keys.Keys)[0].KeyName},
)
```

<a id="update"></a>

### Update the storage account

Just like all resources, storage accounts can be updated.

```go
storageClient.Update(groupName, accountName, storage.AccountUpdateParameters{
    Tags: &map[string]*string{
        "who rocks": to.StringPtr("golang"),
        "where":     to.StringPtr("on azure")},
})
```

<a id="listusage"></a>

### List usage

```go
usageList, err := usageClient.List()
onErrorFail(err, "List failed")

for _, usage := range *usageList.Value {
    fmt.Printf("\t%v: %v / %v\n", *usage.Name.Value, *usage.CurrentValue, *usage.Limit)
}
```

<a id="delete"></a>

### Delete storage account

```go
storageClient.Delete(groupName, accountName)
```

<a id="info"></a>

## More information

Please refer to [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go) for more information.
***

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.