---
page_type: sample
languages:
- Go
products:
- azure
description: "These code samples will show you how to manage Communication Service resources using Azure SDK for Go."
urlFragment: communication
---

# Getting started - Managing Azure Communication Services using Azure Go SDK

These code samples will show you how to manage Communication Service resources using Azure SDK for Go.

## Features

This project framework provides examples for the following services:

### Communication
* Using the Azure SDK for Go - Communication Management Library [azure-mgmt-communication](github.com/Azure/azure-sdk-for-go/services/preview/communication/mgmt/2020-08-20-preview/communication) for the [Azure Communication API](https://docs.microsoft.com/en-us/rest/api/communication/)

## Getting Started

### Prerequisites

1. Before we run the samples, we need to make sure we have setup the credentials. Follow the instructions in [register a new application using Azure portal](https://docs.microsoft.com/en-us/azure/active-directory/develop/howto-create-service-principal-portal) to obtain `subscription id`,`client id`,`client secret`, and `application id`

2. Store your credentials in the Go mgmt_quickstart.go file in the class variables.
```bash
var SubscriptionId = "xxx"
var TenantId = "xxx"
var ClientSecret = "xxx"
var ClientId = "xxx"
```

### Installation

1.  If you don't already have it, [install Go](https://golang.org/doc/install).

    This sample (and the SDK) is compatible with Go 1.15.2

### Quickstart

1.  Clone the repository.

    ```
    git clone https://github.com/Azure-Samples/azure-sdk-for-go-samples.git
    ```

2.  Build the sample.

    ```
    cd communication
    go run mgmt_quickstart.go
    ```

## Demo

A test app is included to show how to use the creation API.

To run the complete test follow the instructions in the [base of this repo.](https://github.com/Azure-Samples/azure-sdk-for-go-samples)

The sample files do not have dependency each other and each file represents an individual end-to-end scenario. Please look at the sample that contains the scenario you are interested in.

