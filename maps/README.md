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
    * CreateCreatorsAccount - Creates an Azure Maps Creator account resource
    * DeleteCreatorsAccount - Deletes an existing Azure Maps Creator account resource
* Data plane
    * aliasOperations - Demonstrates the process of assigning aliases to uploaded resources
    * conversionOperations - Demonstrates DWG package conversion process
    * datasetOperations - Demonstrates how to create an Azure Maps Creator dataset
    * elevationOperations - Demonstrates Azure Maps Elevation Service functionality
    * featurestateOperations - Demonstrates Azure Maps Creator Feature State service functionality
    * geolocationOperations - Demonstates how to retrieve CoutryRegion ISO code based on provided IP Address
    * renderOperations - Demonstrates how to retrieve map tiles and its associated metadata via Azure Maps Render service
    * renderV2Operations - Demonstates V2 of Azure Maps Render service used to retrieve vector tiles
    * routeOperations - Demonstates Azure Maps Route service functionality
    * searchOperations - Demonstates Azure Maps Search service that includes geocoding, reverse-geocoding and associated search functionality
    * spatialOperations - Demonstrates Azure Maps Creator Spatial operations that also include geofencing
    * tilesetOperations - Demonstrates Azure Maps Creator Tileset service used to create custom private maps tilesets
    * timezoneOperations - Demonstates Azure Maps Timezone service functionality
    * trafficOperations - Demonstates Azure Maps Traffic service functionality
    * uploadOperations - Demonstates the process of uploading data content to Azure Maps service for subsequent use with Azure Maps Geofencing and Azure Maps Creator functionalities
    * weatherOperations - Demonstates Azure Maps Weather service functionality
    * wfsOperations - Demonstates Azure Maps Creator dataset query functionality via WFS service that implements OGC WFS(Web Feature Service) API standard for Features

<a id="run"></a>
## How to run all samples

1. Get this package and all dependencies.

    ```bash
    export PROJECT=github.com/Azure-Samples/azure-sdk-for-go-samples/maps
    go get -u $PROJECT
    cd ${GOPATH}/src/${PROJECT}
    dep ensure
    ```

1. Create an Azure service principal with the [Azure CLI][] command `az ad sp create-for-rbac -n 'azure-sdk-for-go-samples'` and set the following environment variables per that command's output. You can also copy `.env.tpl` to `.env` and fill it in; the configuration system will utilize this.

    ```bash
    AZURE_CLIENT_ID=
    AZURE_CLIENT_SECRET=
    AZURE_TENANT_ID=
    AZURE_SUBSCRIPTION_ID=
    AZURE_BASE_GROUP_NAME=
    AZURE_LOCATION_DEFAULT=westus2
    ```

1. Make sure environment variables are applied in current session. For example via: `export $(xargs < .env)` 
1. Run the tests via `go test -timeout 1h` with shared key authentification. (You can also run tests from the root of azure-sdk-for-go-samples with `export $(xargs < .env) && go test -v ./maps/ -timeout 1h`)

1. **Azure Active Directory Auth**: If you want to use RBAC and Azure Active directory auth, pass `-ad-auth` to test runner: `go test -timeout 1h -ad-auth`. A newly created service principle will also require `Azure Maps Data Contributor` role assignment to allow the access to Azure Maps data plane APIs: extract the principal objectId via running: `az ad sp list --display-name 'azure-sdk-for-go-samples"'`, then use `az role assignment create --assignee 'PRINCIPLE_OBJECT_ID' --role 'Azure Maps Data Contributor' --subscription 'AZURE_SUBSCRIPTION_ID'` to assign the role.

**Note:** expect around 30 minutes for all samples to complete. You might find practical to run an individual example like: `go test -v ./maps/ -run searchOperations`
 
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
      "args": ["-ad-auth"]
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