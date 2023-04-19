---
page_type: sample
languages:
- go
products:
- azure
description: "These code samples will show you how to manage Automation using Azure SDK for Golang."
urlFragment: automation-module
---

# Getting started - Managing Automation using Azure Golang SDK

These code samples will show you how to manage Automation using Azure SDK for Golang.

## Features

This project framework provides examples for the following services:

### Automation
* Using the Azure SDK for Golang - Automation Management Library [automation/armautomation](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/automation/armautomation) for the [Azure Automation API](https://docs.microsoft.com/en-us/rest/api/automation/)

### Prerequisites
* an [Azure subscription](https://azure.microsoft.com)
* Go 1.13 or above

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

3. Run automation sample.

    ```
    cd azure-sdk-for-go-samples/sdk/resourcemanager/automation/module
    go run main.go
    ```
   
## Resources

- https://github.com/Azure/azure-sdk-for-go
- https://docs.microsoft.com/en-us/azure/developer/go/
- https://docs.microsoft.com/en-us/rest/api/
- https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk

## Need help?

Post issue on Github (https://github.com/Azure/azure-sdk-for-go/issues)
