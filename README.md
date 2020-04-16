---
languages:
- go
products:
- azure
page_type: sample
description: "A collection of samples showing how to use the Azure SDK for Go."
---

# Azure SDK for Go Samples

azure-sdk-for-go-samples is a collection of sample usages of the [Azure/azure-sdk-for-go][].

[![Build Status](https://dev.azure.com/azure-sdk/public/_apis/build/status/Azure-Samples.azure-sdk-for-go-samples?branchName=master)](https://dev.azure.com/azure-sdk/public/_build/latest?definitionId=1666&branchName=master)

For general SDK help start with the [main SDK README][].

## To run tests

1. set up authentication (see following)
1. `go test -v ./network/` (or any package)

To use service principal authentication, create a principal by running `az ad sp create-for-rbac -n "<yourAppName>"` and set the following environment variables. You can copy `.env.tpl` to a `.env` file in each package for ease of use.

```bash
export AZURE_SUBSCRIPTION_ID=
export AZURE_TENANT_ID=
export AZURE_CLIENT_ID=
export AZURE_CLIENT_SECRET=

export AZURE_LOCATION_DEFAULT=westus2
export AZURE_BASE_GROUP_NAME=azure-samples-go
export AZURE_KEEP_SAMPLE_RESOURCES=0
```

For device flow authentication, create a "native" app by running `az ad app
create --display-name "<yourAppName>" --native-app --requiredResourceAccess
@manifest.json`; and specify the `-useDeviceFlow` flag when running tests.

## Other notes

`AZURE_SP_OBJECT_ID` represents a service principal ObjectID. It is needed to
run the Create VM with encrypted managed disks sample.

## Resources

- SDK code is at [Azure/azure-sdk-for-go][].
- SDK docs are at [godoc.org](https://godoc.org/github.com/Azure/azure-sdk-for-go/).
- SDK notifications are published via the [Azure update feed][].
- Azure API docs are at [docs.microsoft.com/rest/api](https://docs.microsoft.com/rest/api/).
- General Azure docs are at [docs.microsoft.com/azure](https://docs.microsoft.com/azure).

## License

This code is provided under the MIT license. See [LICENSE][] for details.

## Contribute

We welcome your contributions! For instructions and our code of conduct see [CONTRIBUTING.md][]. And thank you!

[main SDK README]: https://github.com/Azure/azure-sdk-for-go/blob/master/README.md
[Azure update feed]: https://azure.microsoft.com/updates/
[Azure/azure-sdk-for-go]: https://github.com/Azure/azure-sdk-for-go
[azure-cli]: https://github.com/Azure/azure-cli
[LICENSE]: ./LICENSE.md
[CONTRIBUTING.md]: ./CONTRIBUTING.md
