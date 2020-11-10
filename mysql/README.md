---
services: mysql
platforms: go
author: gechris
---

# Azure mysql Samples

This package demonstrates how to manage Azure VMs, their disks and container
instances with the Go SDK.

The child package "hybrid" demonstrates how to manage Azure VMs using Azure's
Hybrid profile.

## Contents

* [How to run all samples](#run)
* Management
    * CreateServer - Create a PostgreSQL.
    * UpdateServer - Updates a PostgreSQL server.
    * DeleteServer - Deletes an existing PostgreSQL server.
    * CreateOrUpdateFirewallRules - Creates or updates a firewall rule on the server.
    * GetConfiguration - Get the configuration value that is set on the server.
    * UpdateConfiguration - Updates a configuration on the server.

<a id="run"></a>
## How to run all samples

1. Get this package and all dependencies.

  ```bash
  export PROJECT=github.com/Azure-Samples/azure-sdk-for-go-samples/mysql
  go get -u $PROJECT
  cd ${GOPATH}/src/${PROJECT}
  dep ensure
  ```
1. Create an Azure service principal with the [Azure CLI][] command `az ad sp
   create-for-rbac --output json` and set the following environment variables
   per that command's output. You can also copy `.env.tpl` to `.env` and fill
   it in; the configuration system will utilize this.

  ```bash
  AZURE_CLIENT_ID=
  AZURE_CLIENT_SECRET=
  AZURE_TENANT_ID=
  AZURE_SUBSCRIPTION_ID=
  AZURE_BASE_GROUP_NAME=
  AZURE_LOCATION_DEFAULT=westus2
  ```

1. TODO(joshgav): grant this principal all-powerful rights to your AAD tenant to faciliate identity-related operations.
1. Run the tests: `go test -v -timeout 12h`

  The timeout is optional, but some tests take longer than then default 10m to complete.

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