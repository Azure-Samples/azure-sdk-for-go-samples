# Azure SDK for Go - Hybrid Samples 

The goal of hybrid samples is to provide a library of code snippets for common
operations in Azure Stack via the Go SDK. Currently "compute", "network", 
"resource", and "storage" services are supported on Azure Stack. 
Hybrid sample code is organized by snippet type and is placed in "hybrid" folder
in the supported services folders.

Note: Device authentication is not been enabled for Hybrid samples. 

## To run tests

1. Set the following environment variables (those marked * are required). Use
the following instructions to find or create these values if necessary.

    * `AZURE_SUBSCRIPTION_ID`*
    * `AZURE_TENANT_ID`*
    * `AZURE_CLIENT_ID`*
    * `AZURE_CLIENT_SECRET`*
    * `AZURE_LOCATION`*
    * `AZURE_ENVIRONMENT`*
    * `AZURE_RESOURCE_GROUP_NAME`
    * `AZURE_SAMPLES_KEEP_RESOURCES`

    Using [the Azure CLI][azure-cli], you can get your subscription ID by running `az account
    list`. You can check your tenant ID and get a client ID and secret by
    running `az ad sp create-for-rbac -n "<yourAppName>"`.

    Using [the Azure CLI][azure-cli], you can get your ARM endpoint and storage suffix by running
    `az cloud show`.   

    If `AZURE_RESOURCE_GROUP_NAME` isn't specified a random name will be used.

    If `AZURE_KEEP_SAMPLE_RESOURCES` is set to `1` tests won't clean up resources
    they create when done. This can be helpful if you want to further experiment
    with those resources.

1. Run `dep ensure` to get dependencies.
1. Run tests with `go test` as follows:

    1. To run individual samples, refer to that folder, e.g. `go test ./storage/hybrid/`, `go test ./network/hybrid/`.
    
# Resources

- SDK code is at [Azure/azure-sdk-for-go][].
- SDK docs are at [godoc.org](https://godoc.org/github.com/Azure/azure-sdk-for-go/).
- SDK notifications are published via the [Azure update feed][].
- Azure API docs are at [docs.microsoft.com/rest/api](https://docs.microsoft.com/rest/api/).
- General Azure docs are at [docs.microsoft.com/azure](https://docs.microsoft.com/azure).

# License

This code is provided under the MIT license. See [LICENSE][] for details.

# Contribute

We welcome your contributions! For instructions and our code of conduct see [CONTRIBUTING.md][]. And thank you!

[azure-cli]: https://github.com/Azure/azure-cli
[LICENSE]: ./LICENSE.md
[CONTRIBUTING.md]: ./CONTRIBUTING.md
