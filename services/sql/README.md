---
services: mssql
platforms: go
author: joshgav
---

# Azure MSSQL Samples

This // Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package demonstrates how to manage SQL databases and their data.

## Contents

* [How to run all samples](#run)
* Management
    * CreateServer
    * CreateDatabase
    * CreateFirewallRules
* Data plane
    * Open
    * CreateTable
    * Insert
    * Query

<a id="run"></a>
## How to run all samples

1. Get this repo and all dependencies.

  ```bash
  export PROJECT=github.com/Azure-Samples/azure-sdk-for-go-samples/services/compute
  go get -u $PROJECT
  cd ${GOPATH}/src/${PROJECT}
  dep ensure
  ```
1. Create an Azure service principal with the [Azure CLI][] command `az ad sp
   create-for-rbac`.
1. Set the following environment variables based on output properties of this
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

1. Run the sample. `go test -v`

<a id="info"></a>
## More information

Please refer to [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go)
for more information.

---

This project has adopted the [Microsoft Open Source Code of
Conduct](https://opensource.microsoft.com/codeofconduct/). For more information
see the [Code of Conduct
FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact
[opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional
questions or comments.
