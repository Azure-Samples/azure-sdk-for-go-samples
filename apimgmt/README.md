---
services: api management service
platforms: go
author: wimortl
---

# Azure API Management Service Samples

This package demonstrates how to create an instance of the API Management Service in Azure as well as how to create API endpoints using the GO SDK for Azure. Please note that the end-to-end test is set to _t.SkipNow()_ because it takes about 60 minutes to run (the API Mangement Service takes quite a while to spin-up). When executing the end-to-end test, please allow for enough time for it to complete, as such:

```bash
% go test -timeout 60m
```

# How to Install and Run the End-To-End Test

1. Get this repo and all dependencies.

  ```bash
  export PROJECT=github.com/Azure-Samples/azure-sdk-for-go-samples/apimgmt
  go get -u $PROJECT
  cd ${GOPATH}/src/${PROJECT}
  dep ensure
  ```
2. Create an Azure service principal with the [Azure CLI][] command `az ad sp
   create-for-rbac`.
3. Set the following environment variables based on output properties of this
   command. You can fill in these variables in `.env.tpl` in this directory and
   rename that to `.env`.

  |EnvVar | Value|
  |-------|------|
  |AZURE_CLIENT_ID|service principal/application ID|
  |AZURE_CLIENT_SECRET|service principal/application secret|
  |AZURE_TENANT_ID|your tenant id|
  |AZURE_SUBSCRIPTION_ID|your subscription ID|
  |AZURE_BASE_GROUP_NAME|base resource group name|
  |AZURE_LOCATION_DEFAULT|location for all resources|

4. Edit *endtoend_test.go* file and preface the `t.SkipNow()` line in the *TestEndToEnd* function with a comment (//).

5. Run the sample. `go test -timeout 60m`

# More information

Please refer to [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go)
for more information.

---

This project has adopted the [Microsoft Open Source Code of
Conduct](https://opensource.microsoft.com/codeofconduct/). For more information
see the [Code of Conduct
FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact
[opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional
questions or comments.
