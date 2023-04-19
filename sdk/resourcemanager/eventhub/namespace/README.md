---
page_type: sample
languages:
- go
products:
- azure
description: "These code samples will show you how to manage Event Hubs using Azure SDK for Golang."
urlFragment: event-hub-name-space
---

# Getting started - Managing Event Hubs using Azure Golang SDK

These code samples will show you how to manage Event Hubs using Azure SDK for Golang.

## Features

This project framework provides examples for the following services:

### Event Hubs
* Using the Azure SDK for Golang - Event Hubs Management Library [eventhub/armeventhub](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub) for the [Azure Event Hubs API](https://docs.microsoft.com/en-us/rest/api/eventhub/)

### Prerequisites
* an [Azure subscription](https://azure.microsoft.com)
* Go 1.18 or above

### Quickstart

1. Clone the repository.

    ```
    git clone https://github.com/Azure-Samples/azure-sdk-for-go-samples.git
    ```
2. Set the environment variable.

   ```
   
   export AZURE_SUBSCRIPTION_ID=<your Azure subscription id> 
   # If no value is set, the created resource will be deleted by default.
   # anything other than empty to keep the resources
   export KEEP_RESOURCE=1 
   export AZURE_TENANT_ID=<your Azure Tenant id>          
   export AZURE_OBJECT_ID=<your Azure Client/Object id>          
   ```

3. Run eventhub sample.

    ```
    cd azure-sdk-for-go-samples/sdk/resourcemanager/eventhub/namespace
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
