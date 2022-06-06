# Azure SDK for Go Samples

This repo contains a collection of sample usages of the new version of [Azure/azure-sdk-for-go][]. All the samples are placed under [`sdk`](https://github.com/Azure-Samples/azure-sdk-for-go-samples/tree/main/sdk) folder and the folder structures are corresponding to the service packages under the `/sdk` directory of the [Azure/azure-sdk-for-go][] repo.

For general SDK help, please refer to the [SDK README][].

## To run tests

### Prerequisites

You will need Go 1.18 and latest version of resource management modules.

You will need to authenticate to Azure either by using Azure CLI to sign in or setting environment variables.

#### Using Azure CLI to Sign In

You could easily use `az login` in command line to sign in to Azure via your default browser. Detail instructions can be found in [Sign in with Azure CLI](https://docs.microsoft.com/cli/azure/authenticate-azure-cli).

#### Setting Environment Variables

You will need the following values to authenticate to Azure

-   **Subscription ID**
-   **Client ID**
-   **Client Secret**
-   **Tenant ID**

These values can be obtained from the portal, here's the instructions:

- Get Subscription ID

    1.  Login into your Azure account
    2.  Select Subscriptions in the left sidebar
    3.  Select whichever subscription is needed
    4.  Click on Overview
    5.  Copy the Subscription ID

- Get Client ID / Client Secret / Tenant ID

    For information on how to get Client ID, Client Secret, and Tenant ID, please refer to [this document](https://docs.microsoft.com/azure/active-directory/develop/howto-create-service-principal-portal)

- Setting Environment Variables

    After you obtained the values, you need to set the following values as your environment variables

    -   `AZURE_CLIENT_ID`
    -   `AZURE_CLIENT_SECRET`
    -   `AZURE_TENANT_ID`
    -   `AZURE_SUBSCRIPTION_ID`

    To set the following environment variables on your development system:

    Windows (Note: Administrator access is required)

    1.  Open the Control Panel
    2.  Click System Security, then System
    3.  Click Advanced system settings on the left
    4.  Inside the System Properties window, click the `Environment Variables…` button.
    5.  Click on the property you would like to change, then click the `Edit…` button. If the property name is not listed, then click the `New…` button.

    Linux-based OS :

        export AZURE_CLIENT_ID="__CLIENT_ID__"
        export AZURE_CLIENT_SECRET="__CLIENT_SECRET__"
        export AZURE_TENANT_ID="__TENANT_ID__"
        export AZURE_SUBSCRIPTION_ID="__SUBSCRIPTION_ID__"

### Run tests

1. Clone the repository.

    ```
    git clone https://github.com/Azure-Samples/azure-sdk-for-go-samples.git
    ```
2. Set the environment variable.

   ```
   # bash
   export AZURE_SUBSCRIPTION_ID=<your Azure subscription id> 
   # If no value is set, the created resource will be deleted by default.
   # anything other than empty to keep the resources
   export KEEP_RESOURCE=1 
   
   # powershell
   $env:AZURE_SUBSCRIPTION_ID=<your Azure subscription id> 
   $env:KEEP_RESOURCE=1
   ```

3. Choose one sample and run.

    ```
    cd azure-sdk-for-go-samples/sdk/resourcemanager/<service>/<single sample>
    go run main.go
    ```
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

[SDK README]: https://github.com/Azure/azure-sdk-for-go/blob/main/README.md
[Azure update feed]: https://azure.microsoft.com/updates/
[Azure/azure-sdk-for-go]: https://github.com/Azure/azure-sdk-for-go
[LICENSE]: ./LICENSE.txt
[CONTRIBUTING.md]: ./CONTRIBUTING.md