---
page_type: sample
languages:
- golang
products:
- azure
description: "These code samples will show you how to manage Storage using Azure SDK for Golang."
urlFragment: storage
---

# Getting started - Managing Storage using Azure Golang SDK

These code samples will show you how to manage Storage using Azure SDK for Golang.

## Features

This project framework provides examples for the following services:

### Storage
* Using the Azure SDK for Golang - Storage Library [storage/armstorage](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage) for the [Azure Storage API](https://docs.microsoft.com/en-us/rest/api/storage/)

### Prerequisites
* an [Azure subscription](https://azure.microsoft.com)
* Go 1.18 or above

### Quickstart

1. Clone the repository.

    ```
    git clone https://github.com/Azure-Samples/azure-sdk-for-go-samples.git --branch new-version
    ```
   
2. Set the environment variable.

   ```
   # bash
   export AZURE_SUBSCRIPTION_ID=<your Azure subscription id> 
   # If no value is set, the created resource will be deleted by default.
   # anything other than empty to keep the resources
   export KEEP_RESOURCE=1 
   export AZURE_TENANT_ID=<your Azure Tenant id>          
   export AZURE_OBJECT_ID=<your Azure Client/Object id> 
   ```

3. Run storage sample.

    ```
    cd azure-sdk-for-go-samples/storage/table
    go mod tidy
    go run main.go
    ```
   
## Resources

- https://github.com/Azure/azure-sdk-for-go
- https://docs.microsoft.com/en-us/azure/developer/go/
- https://docs.microsoft.com/en-us/rest/api/
- https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk

## Need help?

Post issue on Github (https://github.com/Azure/azure-sdk-for-go/issues)
