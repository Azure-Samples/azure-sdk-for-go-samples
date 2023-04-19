---
page_type: sample
languages:
- go
products:
- azure
description: "These code samples will show you how to manage Monitor Manager using Azure SDK for Golang."
urlFragment: monitor-autoscale-setting
---

# Getting started - Managing Monitor using Azure Golang SDK

These code samples will show you how to manage Monitor using Azure SDK for Golang.

## Features

This project framework provides examples for the following services:

### Monitor
* Using the Azure SDK for Golang - Monitor Management Library [monitor/armmonitor](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor) for the [Azure Monitor API](https://docs.microsoft.com/en-us/rest/api/monitor/)

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

3. Run monitor sample.

    ```
    cd azure-sdk-for-go-samples/sdk/resourcemanager/monitor/autoscale_setting
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
