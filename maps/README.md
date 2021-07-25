---
services: maps
platforms: go
author: tarasvozniuk
---

# Azure Maps Samples

This package demonstrates how to use Azure Maps functionality described at [Azure Maps Documentation](https://docs.microsoft.com/en-us/rest/api/maps/)


## Contents

* [How to run all samples](#run)
* Management
    * CreateMapsAccount - Creates an Azure Maps Account resource
    * DeleteMapsAccount - Deletes an existing Azure Maps Account resource
* Data plane
    * uploadOperations - Demonstates the process of uploading data content to Azure Maps service for subsequent use with Azure Maps Geofencing and Azure Maps Creator functionalities

<a id="run"></a>
## How to run all samples

1. Get this package and all dependencies.

    ```bash
    export PROJECT=github.com/Azure-Samples/azure-sdk-for-go-samples/maps
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

1. Assign `Azure Maps Data Contributor` role to newly created service principle to allow the access to Azure Maps data plane APIs with command `az role assignment create --assignee 'PRINCIPLE_OBJECT_ID' --role 'Azure Maps Data Contributor' --subscription 'AZURE_SUBSCRIPTION_ID'`
1. Run the tests via `go test` with Azure Active Directory authentication used in data plane APIs or `go test -sharedkey-auth` for shared key authentication. (You can also run tests from the root of azure-sdk-for-go-samples with `export $(xargs < .env) && go test -v ./maps/`)
  
## Debugging tests in VSCode

Sample debugging configuration at `./azure-sdk-for-go-samples/.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Go",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${fileDirname}",
      "envFile": "${workspaceFolder}/.env",
      "args": ["-sharedkey-auth"]
    }
  ]
}
```

You may then start debugging by hitting F5 when your `some_test.go` is active in the code editor.

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