---
page_type: sample
languages:
- golang
products:
- azure
description: "These code samples will show you how to manage Compute using Azure SDK for Golang."
urlFragment: compute
---

# Getting started - Managing Compute using Azure Golang SDK

These code samples will show you how to manage Compute using Azure SDK for Golang.

## Features

This project framework provides examples for the following services:

### Compute
* Using the Azure SDK for Golang - Compute Management Library [compute/armcompute](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/compute/armcompute) for the [Azure Compute API](https://docs.microsoft.com/en-us/rest/api/compute/)

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
   #Linux
   export AZURE_SUBSCRIPTION_ID=<your Azure subscription id> 
   # If no value is set, the created resource will be deleted by default.
   export KEEP_RESOURCE=
   
   #PowerShell
   $env:AZURE_SUBSCRIPTION_ID=<your Azure subscription id> 
   export KEEP_RESOURCE=
   ```

3. Run compute sample.

    ```
    cd azure-sdk-for-go-samples/compute
    go run main.go
    ```
   
## Resources

- https://github.com/Azure/azure-sdk-for-go
