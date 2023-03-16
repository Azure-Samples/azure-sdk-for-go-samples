---
page_type: sample
languages:
- golang
products:
- azure
description: "These code samples will show you how to get started using Azure SDK for Golang."
urlFragment: go
---

# Getting started - Azure SDK for Go

These code sample will show you the basic usage pattern for Azure SDK for Go.

## Features

This project framework provides examples for the following usage pattern:

- How to create management plane clients - [`ExampleUsingARMClients`](azstart.go?plain=1#L14)
- How to create data plane clients - [`ExampleUsingDataPlaneClients`](azstart.go?plain=1#L44)
- How to page over responses - [`ExamplePagingOverACollection`](azstart.go?plain=1#L70)
- How to use long running operations - [`ExampleLongRunningOperation`](azstart.go?plain=1#L103)

### Prerequisites
* An [Azure subscription](https://azure.microsoft.com)
* Go 1.18 or above

### Quickstart

1. Clone the repository.

    ```bash
    git clone https://github.com/Azure-Samples/azure-sdk-for-go-samples.git --branch new-version
    ```

1. Run `azstart` sample.

    ```bash
    cd azure-sdk-for-go-samples/sdk/azstart
    go run azstart.go
    ```
   
## Resources

- https://github.com/Azure/azure-sdk-for-go
- https://docs.microsoft.com/en-us/azure/developer/go/
- https://docs.microsoft.com/en-us/rest/api/
- https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk

## Need help?

Post issue on Github (https://github.com/Azure/azure-sdk-for-go/issues)
