---
page_type: sample
languages:
- go
products:
- azure
description: "These code samples will show you how to manage Container Service using Azure SDK for Golang."
urlFragment: container-service-managed-clusters
---

# Getting started - Managing Container Service using Azure Golang SDK

These code samples will show you how to manage Container Service using Azure SDK for Golang.

## Features

This project framework provides examples for the following services:

### Container Service
* Using the Azure SDK for Golang - Container Service Management Library [containerservice/armcontainerservice](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice) for the [Azure Container Service API](https://docs.microsoft.com/en-us/rest/api/aks/)

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
   # bash
   export AZURE_SUBSCRIPTION_ID=<your Azure subscription id> 
   # If no value is set, the created resource will be deleted by default.
   # anything other than empty to keep the resources
   export KEEP_RESOURCE=1 
   export AZURE_TENANT_ID=<your Azure Tenant id>          
   export AZURE_OBJECT_ID=<your Azure Client/Object id> 
   ```

3. Run containerservice sample.

    ```
    cd azure-sdk-for-go-samples/sdk/resourcemanager/containerservice/managedclusters
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
