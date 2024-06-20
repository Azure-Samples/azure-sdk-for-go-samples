# Web App

Continuously build and deploy a Go web app using Azure App Service and Azure
Container Registry.

This package includes a basic Go web app and scripts to set up infrastructure
for it in Azure.

Requires [Azure CLI][].

## Try It!

Set configuration in a local .env file in the package root by copying
`.env.tpl` to `.env`. Change `AZURE_BASE_NAME` to something relatively unique
to you; for example you might include your name. Then run ./[setup.sh][] to
build and deploy your container to Container Registry and App Service.

If run within the Azure Go SDK samples repo, this will carry out a one-off
build and push to your registry, which will trigger refresh of the App Service
Web App.  If you stick with the script defaults, you can visit your app at
`https://${AZURE_BASE_NAME}-webapp.azurewebsites.net/`.

### Continuous Build and Deploy

Follow these steps to set up continuous build and deploy:

1. Copy the contents of this package to your own fresh git repo and push it to GitHub.
2. Specify an image name in env var `IMAGE_NAME` (e.g. in `.env`) that matches
   your GitHub 'org/repo' structure.
3. Run `./setup.sh`. It will arrange continuous build and deploy for you from
   the specified repo/image name.

**NOTE**: Container Registry Build requires a [GitHub personal access
token](https://github.com/settings/tokens); you need to get one from the linked
page and set it in a local environment variable `GH_TOKEN`, e.g. `export
GH_TOKEN=mylongtokenstring`. You can also add it to your local `.env` file for
persistence.

To test continuous integration, now make a change and `git push` it to your
repo. The Container Registry build task should detect the change, rebuild your
container, and notify App Service; which should then refresh and reload your
container image and app.

## More Details

[setup.sh][] ensures an Azure resource group, container registry, app service
plan, and container-based web app are provisioned and connected in the
subscription currently logged in to [Azure CLI][].

It uses the following environment variables to choose names:

* IMAGE\_NAME: Name of container image (aka "repo").
* IMAGE\_TAG: Tag for container image.
* AZURE\_BASE\_NAME: Prefix for Azure resources.
* AZURE\_DEFAULT\_LOCATION: Location for Azure resources. 

These names can be specified in a .env file in the root of the package. If a
`.env` file isn't found, `.env.tpl` is copied to `.env` and used.

Explicit parameters can also be passed, see comments at beginning of
[setup.sh][] for details.

[Azure CLI]: https://github.com/Azure/azure-cli
[setup.sh]: ./setup.sh
